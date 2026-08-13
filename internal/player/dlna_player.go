// Package player provides DLNA (Digital Living Network Alliance) player implementation
// for streaming audio to DLNA-compatible devices.
package player

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/types"
)

const soapEnvelopeTpl = `<?xml version="1.0"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
  <s:Body>
    %s
  </s:Body>
</s:Envelope>`

const setAvTransportURIBody = `<u:SetAVTransportURI xmlns:u="urn:schemas-upnp-org:service:AVTransport:1">
  <InstanceID>0</InstanceID>
  <CurrentURI>%s</CurrentURI>
  <CurrentURIMetaData></CurrentURIMetaData>
</u:SetAVTransportURI>`

const playBody = `<u:Play xmlns:u="urn:schemas-upnp-org:service:AVTransport:1">
  <InstanceID>0</InstanceID>
  <Speed>1</Speed>
</u:Play>`

const pauseBody = `<u:Pause xmlns:u="urn:schemas-upnp-org:service:AVTransport:1">
  <InstanceID>0</InstanceID>
</u:Pause>`

const stopBody = `<u:Stop xmlns:u="urn:schemas-upnp-org:service:AVTransport:1">
  <InstanceID>0</InstanceID>
</u:Stop>`

const seekBody = `<u:Seek xmlns:u="urn:schemas-upnp-org:service:AVTransport:1">
  <InstanceID>0</InstanceID>
  <Unit>REL_TIME</Unit>
  <Target>%s</Target>
</u:Seek>`

const getPositionInfoBody = `<u:GetPositionInfo xmlns:u="urn:schemas-upnp-org:service:AVTransport:1">
  <InstanceID>0</InstanceID>
</u:GetPositionInfo>`

const getTransportInfoBody = `<u:GetTransportInfo xmlns:u="urn:schemas-upnp-org:service:AVTransport:1">
  <InstanceID>0</InstanceID>
</u:GetTransportInfo>`

const getVolumeBody = `<u:GetVolume xmlns:u="urn:schemas-upnp-org:service:RenderingControl:1">
  <InstanceID>0</InstanceID>
  <Channel>Master</Channel>
</u:GetVolume>`

const setVolumeBody = `<u:SetVolume xmlns:u="urn:schemas-upnp-org:service:RenderingControl:1">
  <InstanceID>0</InstanceID>
  <Channel>Master</Channel>
  <DesiredVolume>%d</DesiredVolume>
</u:SetVolume>`

type cmdType int

const (
	cmdPlay cmdType = iota
	cmdPause
	cmdResume
	cmdStop
	cmdSeek
	cmdSetVolume
)

type command struct {
	cmd    cmdType
	param  any
	result chan any
}

type dlnaPlayer struct {
	deviceURL           string
	controlURL          string
	renderingControlURL string
	audioURL            string
	audioDur            time.Duration
	httpClient          *http.Client
	// mu 保护 state/curPos/audioURL/audioDur/cachedVolume 及播放计时字段：
	// worker goroutine（executeCmd/pollStateTask）写入，UI 线程读取
	mu        sync.RWMutex
	state     types.State
	stateChan chan types.State
	closed    chan struct{}
	closeOnce sync.Once // 保证 closed 只关闭一次（永不置 nil，sendCmd 依赖它）
	cmdQueue  chan command

	curPos   time.Duration
	timeChan chan time.Duration

	httpServer *http.Server
	httpPort   int
	localIP    string
	fileMap    map[int64]string
	fileMapMu  sync.RWMutex

	startTime     time.Time
	pausedTime    time.Duration
	pauseStart    time.Time
	wasEverPlayed bool

	cachedVolume int
}

func NewDlnaPlayer(deviceURL, localIP string) (Player, error) {
	p := &dlnaPlayer{
		deviceURL:  deviceURL,
		localIP:    localIP,
		httpClient: &http.Client{Timeout: 1 * time.Second},
		state:      types.Stopped,
		stateChan:  make(chan types.State, 10),
		closed:     make(chan struct{}),
		timeChan:   make(chan time.Duration, 1),
		fileMap:    make(map[int64]string),
		cmdQueue:   make(chan command, 10),
	}
	if err := p.startHTTPServer(); err != nil {
		return nil, err
	}
	if err := p.initControlURL(); err != nil {
		p.Close()
		return nil, err
	}
	go p.worker()
	return p, nil
}

type dlnaRoot struct {
	XMLName xml.Name   `xml:"root"`
	URLBase string     `xml:"URLBase"`
	Device  dlnaDevice `xml:"device"`
}

type dlnaDevice struct {
	Services []dlnaService `xml:"serviceList>service"`
}

type dlnaService struct {
	ServiceType string `xml:"serviceType"`
	ControlURL  string `xml:"controlURL"`
}

func (p *dlnaPlayer) initControlURL() error {
	slog.Debug("DLNA: fetching device description", "url", p.deviceURL)
	resp, err := p.httpClient.Get(p.deviceURL)
	if err != nil {
		slog.Error("DLNA: failed to fetch device description", "error", err)
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error("DLNA: failed to close response body", "error", err)
		}
	}()
	xmlData, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("DLNA: failed to read device description", "error", err)
		return err
	}

	var r dlnaRoot
	if err := xml.Unmarshal(xmlData, &r); err != nil {
		slog.Error("DLNA: failed to parse device description", "error", err)
		slog.Debug("DLNA: raw XML", "data", string(xmlData))
		return err
	}

	base := r.URLBase
	if base == "" {
		base = p.deviceURL[:len(p.deviceURL)-len("/description.xml")] + "/"
	}
	base = strings.TrimRight(base, "/")
	slog.Debug("DLNA: using base URL", "base", base)

	for _, svc := range r.Device.Services {
		slog.Debug("DLNA: found service", "type", svc.ServiceType, "controlURL", svc.ControlURL)
		controlURL := p.normalizeURL(base, svc.ControlURL)
		switch svc.ServiceType {
		case "urn:schemas-upnp-org:service:AVTransport:1":
			p.controlURL = controlURL
			slog.Info("DLNA: found AVTransport control URL", "url", controlURL)
		case "urn:schemas-upnp-org:service:RenderingControl:1":
			p.renderingControlURL = controlURL
			slog.Info("DLNA: found RenderingControl control URL", "url", controlURL)
		}
	}

	if p.controlURL == "" {
		return errors.New("DLNA: AVTransport service not found or invalid")
	}
	if p.renderingControlURL == "" {
		slog.Error("DLNA: RenderingControl service not found")
	}

	return nil
}

func (p *dlnaPlayer) worker() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			p.pollStateTask()
		case cmd := <-p.cmdQueue:
			p.executeCmd(cmd)
		case <-p.closed:
			return
		}
	}
}

func (p *dlnaPlayer) pollStateTask() {
	state, err := p.getTransportInfo()
	if err != nil {
		// 查询失败不覆盖状态：网络抖动时设备可能短暂不可达
		return
	}
	if state == "STOPPED" || state == "NO_MEDIA_PRESENT" {
		p.mu.Lock()
		changed := p.state != types.Stopped
		p.state = types.Stopped
		p.mu.Unlock()
		// 仅状态确实变化时上报：设备持续返回 STOPPED（如无法播放该 URL）时
		// 避免每 5s 触发一次 UI 的自动切歌
		if changed {
			p.sendState(types.Stopped)
		}
		return
	}
	curPos, _, _ := p.getPositionInfo()
	if curPos > 0 {
		p.mu.Lock()
		p.curPos = curPos
		p.mu.Unlock()
		select {
		case p.timeChan <- curPos:
		default:
		}
	}
	if vol, err := p.getVolume(); err == nil {
		p.mu.Lock()
		p.cachedVolume = vol
		p.mu.Unlock()
	}
}

func (p *dlnaPlayer) executeCmd(cmd command) {
	switch cmd.cmd {
	case cmdPlay:
		music := cmd.param.(URLMusic)
		p.mu.Lock()
		p.audioURL = music.URL
		p.audioDur = music.Duration
		p.mu.Unlock()
		p.fileMapMu.Lock()
		for k := range p.fileMap {
			delete(p.fileMap, k)
		}
		p.fileMapMu.Unlock()
		audioURL := music.URL
		if strings.HasPrefix(audioURL, "file://") {
			localPath := strings.TrimPrefix(audioURL, "file://")
			p.fileMapMu.Lock()
			p.fileMap[music.Id] = localPath
			p.fileMapMu.Unlock()
			audioURL = fmt.Sprintf("http://%s:%d/dlna/%d", p.localIP, p.httpPort, music.Id)
		}
		p.mu.Lock()
		p.audioURL = audioURL
		p.mu.Unlock()

		slog.Info("DLNA: setting AVTransport URI", "audioURL", audioURL)
		p.doSOAP("AVTransport", "SetAVTransportURI", fmt.Sprintf(setAvTransportURIBody, audioURL))

		slog.Info("DLNA: starting playback")
		p.doSOAP("AVTransport", "Play", playBody)

		p.mu.Lock()
		p.state = types.Playing
		p.startTime = time.Now()
		p.pausedTime = 0
		p.wasEverPlayed = true
		p.mu.Unlock()
		p.sendState(types.Playing)
		cmd.result <- true

	case cmdPause:
		p.doSOAP("AVTransport", "Pause", pauseBody)
		p.mu.Lock()
		p.state = types.Paused
		p.pauseStart = time.Now()
		p.mu.Unlock()
		p.sendState(types.Paused)
		cmd.result <- true

	case cmdResume:
		p.doSOAP("AVTransport", "Play", playBody)
		p.mu.Lock()
		p.state = types.Playing
		p.pausedTime += time.Since(p.pauseStart)
		p.mu.Unlock()
		p.sendState(types.Playing)
		cmd.result <- true

	case cmdStop:
		p.doSOAP("AVTransport", "Stop", stopBody)
		p.mu.Lock()
		p.curPos = 0
		p.state = types.Stopped
		p.startTime = time.Time{}
		p.pausedTime = 0
		p.wasEverPlayed = false
		p.mu.Unlock()
		p.sendState(types.Stopped)
		cmd.result <- true

	case cmdSeek:
		duration := cmd.param.(time.Duration)
		seekTime := formatDuration(duration)
		p.doSOAP("AVTransport", "Seek", fmt.Sprintf(seekBody, seekTime))
		cmd.result <- true

	case cmdSetVolume:
		volume := cmd.param.(int)
		p.mu.RLock()
		curVol := p.cachedVolume
		p.mu.RUnlock()
		if volume >= 0 && volume <= 100 && volume != curVol {
			body := fmt.Sprintf(setVolumeBody, volume)
			p.doSOAP("RenderingControl", "SetVolume", body)
			p.mu.Lock()
			p.cachedVolume = volume
			p.mu.Unlock()
		}
		cmd.result <- true
	}
}

func (p *dlnaPlayer) startHTTPServer() error {
	for i := 0; i < 10; i++ {
		listener, err := net.Listen("tcp", p.localIP+":0")
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		p.httpPort = listener.Addr().(*net.TCPAddr).Port
		mux := http.NewServeMux()
		mux.HandleFunc("/dlna/", p.serveLocalFile)
		p.httpServer = &http.Server{Handler: mux}
		go func() {
			if err := p.httpServer.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
				slog.Error("DLNA: HTTP server error", "error", err)
			}
		}()
		slog.Info("DLNA: HTTP server started", "bind", p.localIP, "port", p.httpPort)
		return nil
	}
	return errors.New("DLNA: failed to start HTTP server after 10 attempts")
}

func (p *dlnaPlayer) serveLocalFile(w http.ResponseWriter, r *http.Request) {
	songIDStr := strings.TrimPrefix(r.URL.Path, "/dlna/")
	songID, _ := strconv.ParseInt(songIDStr, 10, 64)
	p.fileMapMu.RLock()
	path := p.fileMap[songID]
	p.fileMapMu.RUnlock()
	http.ServeFile(w, r, path)
}

func (p *dlnaPlayer) doSOAP(service, action, body string) ([]byte, error) {
	controlURL := p.controlURL
	if service == "RenderingControl" {
		controlURL = p.renderingControlURL
		if controlURL == "" {
			return nil, errors.New("RenderingControl service not available")
		}
	}
	envelope := fmt.Sprintf(soapEnvelopeTpl, body)
	req, err := http.NewRequest("POST", controlURL, bytes.NewBufferString(envelope))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", `text/xml; charset="utf-8"`)
	req.Header.Set("SOAPAction", fmt.Sprintf(`"urn:schemas-upnp-org:service:%s:1#%s"`, service, action))

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error("DLNA: failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("soap request failed: %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func (p *dlnaPlayer) getPositionInfo() (time.Duration, time.Duration, error) {
	respBody, err := p.doSOAP("AVTransport", "GetPositionInfo", getPositionInfoBody)
	if err != nil {
		return 0, 0, err
	}

	type envelopeResponse struct {
		Body struct {
			GetPositionInfoResponse struct {
				TrackDuration string `xml:"TrackDuration"`
				RelTime       string `xml:"RelTime"`
			} `xml:"GetPositionInfoResponse"`
		} `xml:"Body"`
	}

	var env envelopeResponse
	if err := xml.Unmarshal(respBody, &env); err != nil {
		return 0, 0, err
	}

	parse := func(t string) time.Duration {
		if t == "" || strings.HasPrefix(t, "NOT_IMPLEMENTED") {
			return 0
		}
		parts := strings.Split(t, ":")
		if len(parts) != 3 {
			return 0
		}
		h, _ := strconv.Atoi(parts[0])
		m, _ := strconv.Atoi(parts[1])
		f := strings.Split(parts[2], ".")
		s, _ := strconv.Atoi(f[0])
		var ms int
		if len(f) > 1 {
			ms, _ = strconv.Atoi(f[1])
		}
		return time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(s)*time.Second + time.Duration(ms)*time.Millisecond
	}

	curPos := parse(env.Body.GetPositionInfoResponse.RelTime)
	totalDur := parse(env.Body.GetPositionInfoResponse.TrackDuration)
	return curPos, totalDur, nil
}

// sendState sends state update non-blockingly, with 2 second timeout
func (p *dlnaPlayer) sendState(state types.State) {
	select {
	case p.stateChan <- state:
	case <-time.After(time.Second * 2):
		slog.Warn("DLNA: stateChan send timeout, drop state update")
	}
}

func (p *dlnaPlayer) Play(music URLMusic) {
	p.sendCmd(command{
		cmd:    cmdPlay,
		param:  music,
		result: make(chan any, 1),
	})
}

// sendCmd 向 worker 投递命令并等待结果。worker 在 Close 后退场、不再消费
// cmdQueue，直接发送会在队列满后永久阻塞 UI 线程；select p.closed 保证
// 播放器已关闭时立即返回。
func (p *dlnaPlayer) sendCmd(cmd command) {
	select {
	case p.cmdQueue <- cmd:
	case <-p.closed:
		return
	}
	select {
	case <-cmd.result:
	case <-p.closed:
	}
}

func (p *dlnaPlayer) getTransportInfo() (string, error) {
	respBody, err := p.doSOAP("AVTransport", "GetTransportInfo", getTransportInfoBody)
	if err != nil {
		return "", err
	}

	type envelopeResponse struct {
		Body struct {
			GetTransportInfoResponse struct {
				CurrentTransportState string `xml:"CurrentTransportState"`
			} `xml:"GetTransportInfoResponse"`
		} `xml:"Body"`
	}

	var env envelopeResponse
	if err := xml.Unmarshal(respBody, &env); err != nil {
		return "", err
	}
	return env.Body.GetTransportInfoResponse.CurrentTransportState, nil
}

func (p *dlnaPlayer) normalizeURL(base, controlURL string) string {
	if controlURL != "" && !bytes.HasPrefix([]byte(controlURL), []byte("http")) {
		controlURL = base + controlURL
	}
	return controlURL
}

func (p *dlnaPlayer) CurMusic() URLMusic {
	p.mu.RLock()
	defer p.mu.RUnlock()
	music := URLMusic{
		URL: p.audioURL,
	}
	music.Duration = p.audioDur
	return music
}

func (p *dlnaPlayer) Pause() {
	p.sendCmd(command{cmd: cmdPause, result: make(chan any, 1)})
}

func (p *dlnaPlayer) Resume() {
	p.sendCmd(command{cmd: cmdResume, result: make(chan any, 1)})
}

func (p *dlnaPlayer) Stop() {
	p.sendCmd(command{cmd: cmdStop, result: make(chan any, 1)})
}

func (p *dlnaPlayer) Toggle() {
	if p.State() == types.Playing {
		p.Pause()
	} else {
		p.Resume()
	}
}

func (p *dlnaPlayer) Seek(duration time.Duration) {
	p.sendCmd(command{cmd: cmdSeek, param: duration, result: make(chan any, 1)})
}

func (p *dlnaPlayer) PassedTime() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.curPos
}

func (p *dlnaPlayer) PlayedTime() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if !p.wasEverPlayed {
		return 0
	}
	return time.Since(p.startTime) - p.pausedTime
}

func (p *dlnaPlayer) TimeChan() <-chan time.Duration {
	return p.timeChan
}

func (p *dlnaPlayer) State() types.State {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state
}

func (p *dlnaPlayer) StateChan() <-chan types.State { return p.stateChan }

func (p *dlnaPlayer) getVolume() (int, error) {
	respBody, err := p.doSOAP("RenderingControl", "GetVolume", getVolumeBody)
	if err != nil {
		return 0, err
	}

	type envelopeResponse struct {
		Body struct {
			GetVolumeResponse struct {
				CurrentVolume string `xml:"CurrentVolume"`
			} `xml:"GetVolumeResponse"`
		} `xml:"Body"`
	}

	var env envelopeResponse
	if err := xml.Unmarshal(respBody, &env); err != nil {
		return 0, err
	}

	volume, _ := strconv.Atoi(env.Body.GetVolumeResponse.CurrentVolume)
	return volume, nil
}

func (p *dlnaPlayer) Volume() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.cachedVolume
}

func (p *dlnaPlayer) SetVolume(volume int) {
	p.sendCmd(command{cmd: cmdSetVolume, param: volume, result: make(chan any, 1)})
}

func (p *dlnaPlayer) UpVolume() {
	p.SetVolume(p.Volume() + 1)
}

func (p *dlnaPlayer) DownVolume() {
	p.SetVolume(p.Volume() - 1)
}

func (p *dlnaPlayer) Close() {
	// closed 通道永不置 nil：sendCmd 依赖它对已关闭的播放器立即返回
	p.closeOnce.Do(func() {
		close(p.closed)
	})
	if p.httpServer != nil {
		if err := p.httpServer.Close(); err != nil {
			slog.Error("DLNA: failed to close HTTP server", "error", err)
		}
	}
}

func formatDuration(d time.Duration) string {
	totalSeconds := d.Seconds()
	hours := int(totalSeconds) / 3600
	minutes := (int(totalSeconds) % 3600) / 60
	seconds := totalSeconds - float64(hours*3600+minutes*60)
	return fmt.Sprintf("%02d:%02d:%06.3f", hours, minutes, seconds)
}
