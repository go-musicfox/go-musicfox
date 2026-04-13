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

type dlnaPlayer struct {
	deviceURL           string
	controlURL          string
	renderingControlURL string
	audioURL            string
	audioDur            time.Duration
	httpClient          *http.Client
	state               types.State
	stateChan           chan types.State
	musicChan           chan URLMusic
	closed              chan struct{}
	ready               chan struct{}

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

func NewDlnaPlayer(deviceURL, localIP string) *dlnaPlayer {
	p := &dlnaPlayer{
		deviceURL:  deviceURL,
		localIP:    localIP,
		httpClient: &http.Client{Timeout: 1 * time.Second},
		state:      types.Stopped,
		stateChan:  make(chan types.State, 10),
		musicChan:  make(chan URLMusic, 10),
		closed:     make(chan struct{}),
		ready:      make(chan struct{}),
		timeChan:   make(chan time.Duration, 1),
		fileMap:    make(map[int64]string),
	}
	p.startHTTPServer()
	go p.initControlURL()
	return p
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

func (p *dlnaPlayer) initControlURL() {
	defer close(p.ready)

	slog.Debug("DLNA: fetching device description", "url", p.deviceURL)
	resp, err := p.httpClient.Get(p.deviceURL)
	if err != nil {
		slog.Error("DLNA: failed to fetch device description", "error", err)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error("DLNA: failed to close response body", "error", err)
		}
	}()
	xmlData, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("DLNA: failed to read device description", "error", err)
		return
	}

	var r dlnaRoot
	if err := xml.Unmarshal(xmlData, &r); err != nil {
		slog.Error("DLNA: failed to parse device description", "error", err)
		slog.Debug("DLNA: raw XML", "data", string(xmlData))
		return
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
		panic("DLNA: AVTransport service not found or invalid")
	}
	if p.renderingControlURL == "" {
		slog.Error("DLNA: RenderingControl service not found")
	}

	// 初始化完成後啟動輪詢，保持連接活躍
	go p.pollState()
}

func (p *dlnaPlayer) startHTTPServer() {
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
		return
	}
	panic("DLNA: failed to start HTTP server after 10 attempts")
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
	defer resp.Body.Close()

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
		if t == "" || t == "NOT_IMPLEMENTED" {
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

func (p *dlnaPlayer) Play(music URLMusic) {
	// 清理 fileMap，避免内存泄漏
	p.fileMapMu.Lock()
	p.fileMap = make(map[int64]string)
	p.fileMapMu.Unlock()

	audioURL := music.URL

	if strings.HasPrefix(audioURL, "file://") {
		localPath := strings.TrimPrefix(audioURL, "file://")
		p.fileMapMu.Lock()
		p.fileMap[music.Id] = localPath
		p.fileMapMu.Unlock()
		audioURL = fmt.Sprintf("http://%s:%d/dlna/%d", p.localIP, p.httpPort, music.Id)
		slog.Debug("DLNA: converted file URL", "original", music.URL, "converted", audioURL)
	}

	p.audioURL = audioURL
	p.audioDur = music.Duration
	p.musicChan <- music

	select {
	case <-p.ready:
	case <-time.After(2 * time.Second):
		slog.Error("DLNA: timeout waiting for control URL init")
		return
	}

	slog.Info("DLNA: setting AVTransport URI", "audioURL", p.audioURL)
	if _, err := p.doSOAP("AVTransport", "SetAVTransportURI", fmt.Sprintf(setAvTransportURIBody, p.audioURL)); err != nil {
		slog.Error("DLNA: SetAVTransportURI failed", "error", err)
		return
	}

	slog.Info("DLNA: starting playback")
	if _, err := p.doSOAP("AVTransport", "Play", playBody); err != nil {
		slog.Error("DLNA: Play failed", "error", err)
		return
	}

	p.state = types.Playing
	p.stateChan <- p.state

	// 初始化播放计时
	p.startTime = time.Now()
	p.pausedTime = 0
	p.wasEverPlayed = true
}

func (p *dlnaPlayer) pollState() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			state, _ := p.getTransportInfo()
			if state == "STOPPED" || state == "NO_MEDIA_PRESENT" {
				p.state = types.Stopped
				p.stateChan <- p.state
				continue
			}
			curPos, _, _ := p.getPositionInfo()
			if curPos > 0 {
				p.curPos = curPos
				select {
				case p.timeChan <- curPos:
				default:
				}
			}
			if vol, err := p.getVolume(); err == nil {
				p.cachedVolume = vol
			}
		case <-p.closed:
			return
		}
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
	music := URLMusic{
		URL: p.audioURL,
	}
	music.Duration = p.audioDur
	return music
}

func (p *dlnaPlayer) Pause() {
	if _, err := p.doSOAP("AVTransport", "Pause", pauseBody); err != nil {
		slog.Error("DLNA: Pause failed", "error", err)
		return
	}
	p.state = types.Paused
	p.stateChan <- p.state

	// 记录暂停时刻
	p.pauseStart = time.Now()
}

func (p *dlnaPlayer) Resume() {
	if _, err := p.doSOAP("AVTransport", "Play", playBody); err != nil {
		slog.Error("DLNA: Resume failed", "error", err)
		return
	}
	p.state = types.Playing
	p.stateChan <- p.state

	// 累加暂停时长
	p.pausedTime += time.Since(p.pauseStart)
}

func (p *dlnaPlayer) Stop() {
	if _, err := p.doSOAP("AVTransport", "Stop", stopBody); err != nil {
		slog.Error("DLNA: Stop failed", "error", err)
		return
	}
	p.curPos = 0
	p.state = types.Stopped
	p.stateChan <- p.state

	// 重置播放计时
	p.startTime = time.Time{}
	p.pausedTime = 0
	p.wasEverPlayed = false
}

func (p *dlnaPlayer) Toggle() {
	if p.state == types.Playing {
		p.Pause()
	} else {
		p.Resume()
	}
}

func (p *dlnaPlayer) Seek(duration time.Duration) {
	seekTime := formatDuration(duration)
	if _, err := p.doSOAP("AVTransport", "Seek", fmt.Sprintf(seekBody, seekTime)); err != nil {
		slog.Debug("DLNA: Seek not supported", "error", err)
	}
}

func (p *dlnaPlayer) PassedTime() time.Duration {
	return p.curPos
}

func (p *dlnaPlayer) PlayedTime() time.Duration {
	if !p.wasEverPlayed {
		return 0
	}
	return time.Since(p.startTime) - p.pausedTime
}

func (p *dlnaPlayer) TimeChan() <-chan time.Duration {
	return p.timeChan
}

func (p *dlnaPlayer) State() types.State { return p.state }

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
	return p.cachedVolume
}

func (p *dlnaPlayer) SetVolume(volume int) {
	body := fmt.Sprintf(setVolumeBody, volume)
	if _, err := p.doSOAP("RenderingControl", "SetVolume", body); err != nil {
		slog.Error("DLNA: SetVolume failed", "error", err)
		return
	}
	p.cachedVolume = volume
}

func (p *dlnaPlayer) UpVolume() {
	p.SetVolume(p.Volume() + 1)
}

func (p *dlnaPlayer) DownVolume() {
	p.SetVolume(p.Volume() - 1)
}

func (p *dlnaPlayer) Close() {
	close(p.closed)
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
