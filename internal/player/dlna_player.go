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
	"os"
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

const setAvTransportUriBody = `<u:SetAVTransportURI xmlns:u="urn:schemas-upnp-org:service:AVTransport:1">
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
	deviceUrl           string
	controlUrl          string
	renderingControlUrl string
	audioUrl            string
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
}

func NewDlnaPlayer(deviceUrl, localIP string) *dlnaPlayer {
	p := &dlnaPlayer{
		deviceUrl:  deviceUrl,
		localIP:    localIP,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		state:      types.Stopped,
		stateChan:  make(chan types.State, 10),
		musicChan:  make(chan URLMusic, 10),
		closed:     make(chan struct{}),
		ready:      make(chan struct{}),
		timeChan:   make(chan time.Duration, 1),
		fileMap:    make(map[int64]string),
	}
	p.startHTTPServer()
	go p.initControlUrl()
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

func (p *dlnaPlayer) initControlUrl() {
	defer close(p.ready)

	slog.Debug("DLNA: fetching device description", "url", p.deviceUrl)
	resp, err := p.httpClient.Get(p.deviceUrl)
	if err != nil {
		slog.Error("DLNA: failed to fetch device description", "error", err)
		return
	}
	defer resp.Body.Close()
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
		base = p.deviceUrl[:len(p.deviceUrl)-len("/description.xml")] + "/"
	}
	base = strings.TrimRight(base, "/")
	slog.Debug("DLNA: using base URL", "base", base)

	for _, svc := range r.Device.Services {
		slog.Debug("DLNA: found service", "type", svc.ServiceType, "controlUrl", svc.ControlURL)
		switch svc.ServiceType {
		case "urn:schemas-upnp-org:service:AVTransport:1":
			controlUrl := svc.ControlURL
			if controlUrl != "" && !bytes.HasPrefix([]byte(controlUrl), []byte("http")) {
				controlUrl = base + controlUrl
			}
			p.controlUrl = controlUrl
			slog.Info("DLNA: found AVTransport control URL", "url", controlUrl)
		case "urn:schemas-upnp-org:service:RenderingControl:1":
			controlUrl := svc.ControlURL
			if controlUrl != "" && !bytes.HasPrefix([]byte(controlUrl), []byte("http")) {
				controlUrl = base + controlUrl
			}
			p.renderingControlUrl = controlUrl
			slog.Info("DLNA: found RenderingControl control URL", "url", controlUrl)
		}
	}

	if p.controlUrl == "" {
		slog.Error("DLNA: AVTransport service not found")
	}
	if p.renderingControlUrl == "" {
		slog.Error("DLNA: RenderingControl service not found")
	}

	// 初始化完成後啟動輪詢，保持連接活躍
	go p.pollPositionInfo()
}

func (p *dlnaPlayer) startHTTPServer() {
	if p.localIP == "" {
		panic("DLNA: localIP is required in config [player.dlna].localIP")
	}
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

func (p *dlnaPlayer) getPositionInfo() (time.Duration, time.Duration, error) {
	body := getPositionInfoBody
	envelope := fmt.Sprintf(soapEnvelopeTpl, body)
	req, err := http.NewRequest("POST", p.controlUrl, bytes.NewBufferString(envelope))
	if err != nil {
		return 0, 0, err
	}
	req.Header.Set("Content-Type", `text/xml; charset="utf-8"`)
	req.Header.Set("SOAPAction", `"urn:schemas-upnp-org:service:AVTransport:1#GetPositionInfo"`)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return 0, 0, fmt.Errorf("GetPositionInfo failed: %d", resp.StatusCode)
	}

	respBody, _ := io.ReadAll(resp.Body)

	type positionInfoResponse struct {
		Track         string `xml:"Track"`
		TrackDuration string `xml:"TrackDuration"`
		RelTime       string `xml:"RelTime"`
		AbsTime       string `xml:"AbsTime"`
		RelCount      string `xml:"RelCount"`
		AbsCount      string `xml:"AbsCount"`
	}

	type envelopeResponse struct {
		Body struct {
			GetPositionInfoResponse positionInfoResponse `xml:"GetPositionInfoResponse"`
		} `xml:"Body"`
	}

	var env envelopeResponse
	if err := xml.Unmarshal(respBody, &env); err != nil {
		slog.Debug("DLNA: failed to parse position info", "error", err, "body", string(respBody))
		return 0, 0, err
	}

	curPos := parseTime(env.Body.GetPositionInfoResponse.RelTime)
	totalDur := parseTime(env.Body.GetPositionInfoResponse.TrackDuration)

	slog.Debug("DLNA: position info", "relTime", env.Body.GetPositionInfoResponse.RelTime, "trackDuration", env.Body.GetPositionInfoResponse.TrackDuration, "parsedCurPos", curPos, "parsedTotal", totalDur)

	return curPos, totalDur, nil
}

func parseTime(t string) time.Duration {
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

func (p *dlnaPlayer) soapRequest(service, action, body string) error {
	envelope := fmt.Sprintf(soapEnvelopeTpl, body)
	var controlUrl string
	switch service {
	case "AVTransport":
		controlUrl = p.controlUrl
	case "RenderingControl":
		controlUrl = p.renderingControlUrl
	default:
		return fmt.Errorf("unknown service: %s", service)
	}
	if controlUrl == "" {
		return fmt.Errorf("control URL for service %s not available", service)
	}
	req, err := http.NewRequest("POST", controlUrl, bytes.NewBufferString(envelope))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", `text/xml; charset="utf-8"`)
	req.Header.Set("SOAPAction", fmt.Sprintf(`"urn:schemas-upnp-org:service:%s:1#%s"`, service, action))

	slog.Debug("DLNA: SOAP request", "action", action, "url", controlUrl)
	resp, err := p.httpClient.Do(req)
	if err != nil {
		slog.Error("DLNA: SOAP request failed", "action", action, "error", err)
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	slog.Debug("DLNA: SOAP response", "action", action, "status", resp.StatusCode, "body", string(respBody))

	if resp.StatusCode >= 400 {
		return fmt.Errorf("soap request failed: %d", resp.StatusCode)
	}
	return nil
}

func isFileReady(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	if fi.Size() == 0 {
		return false
	}
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	_ = f.Close()
	return true
}

func (p *dlnaPlayer) waitForFileReady(path string) {
	for i := 0; i < 10; i++ {
		if isFileReady(path) {
			return
		}
		time.Sleep(1 * time.Second)
	}
}

func (p *dlnaPlayer) Play(music URLMusic) {
	// 清理 fileMap，避免内存泄漏
	p.fileMapMu.Lock()
	p.fileMap = make(map[int64]string)
	p.fileMapMu.Unlock()

	audioURL := music.URL

	if strings.HasPrefix(audioURL, "file://") {
		localPath := strings.TrimPrefix(audioURL, "file://")
		p.waitForFileReady(localPath)
		if !isFileReady(localPath) {
			slog.Error("DLNA: cache file not ready", "path", localPath)
			return
		}
		p.fileMapMu.Lock()
		p.fileMap[music.Id] = localPath
		p.fileMapMu.Unlock()
		audioURL = fmt.Sprintf("http://%s:%d/dlna/%d", p.localIP, p.httpPort, music.Id)
		slog.Debug("DLNA: converted file URL", "original", music.URL, "converted", audioURL)
	}

	p.audioUrl = audioURL
	p.audioDur = music.Duration
	p.musicChan <- music

	select {
	case <-p.ready:
	case <-time.After(2 * time.Second):
		slog.Error("DLNA: timeout waiting for control URL init")
		return
	}

	if p.controlUrl == "" {
		slog.Error("DLNA: control URL not found")
		return
	}

	slog.Info("DLNA: setting AVTransport URI", "audioUrl", p.audioUrl)
	if err := p.soapRequest("AVTransport", "SetAVTransportURI", fmt.Sprintf(setAvTransportUriBody, p.audioUrl)); err != nil {
		slog.Error("DLNA: SetAVTransportURI failed", "error", err)
		return
	}

	// 同步获取当前位置，确保 PassedTime() 立即准确
	if curPos, _, err := p.getPositionInfo(); err == nil {
		p.curPos = curPos
	}

	slog.Info("DLNA: starting playback")
	if err := p.soapRequest("AVTransport", "Play", playBody); err != nil {
		slog.Error("DLNA: Play failed", "error", err)
		return
	}

	p.state = types.Playing
	p.stateChan <- p.state
}

func (p *dlnaPlayer) pollPositionInfo() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// 保持連接活躍：DLNA 設備空閒會關閉連接，需持續輪詢避免
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
		case <-p.closed:
			return
		}
	}
}

func (p *dlnaPlayer) getTransportInfo() (string, error) {
	envelope := fmt.Sprintf(soapEnvelopeTpl, getTransportInfoBody)
	req, err := http.NewRequest("POST", p.controlUrl, bytes.NewBufferString(envelope))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", `text/xml; charset="utf-8"`)
	req.Header.Set("SOAPAction", `"urn:schemas-upnp-org:service:AVTransport:1#GetTransportInfo"`)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("GetTransportInfo failed: %d", resp.StatusCode)
	}

	respBody, _ := io.ReadAll(resp.Body)

	type transportInfoResponse struct {
		CurrentTransportState string `xml:"CurrentTransportState"`
	}

	type envelopeResponse struct {
		Body struct {
			GetTransportInfoResponse transportInfoResponse `xml:"GetTransportInfoResponse"`
		} `xml:"Body"`
	}

	var env envelopeResponse
	if err := xml.Unmarshal(respBody, &env); err != nil {
		return "", err
	}

	return env.Body.GetTransportInfoResponse.CurrentTransportState, nil
}

func (p *dlnaPlayer) CurMusic() URLMusic {
	music := URLMusic{
		URL: p.audioUrl,
	}
	music.Duration = p.audioDur
	return music
}

func (p *dlnaPlayer) Pause() {
	if p.controlUrl == "" {
		return
	}
	if err := p.soapRequest("AVTransport", "Pause", pauseBody); err != nil {
		slog.Error("DLNA: Pause failed", "error", err)
		return
	}
	p.state = types.Paused
	p.stateChan <- p.state
}

func (p *dlnaPlayer) Resume() {
	if p.controlUrl == "" {
		return
	}
	if err := p.soapRequest("AVTransport", "Play", playBody); err != nil {
		slog.Error("DLNA: Resume failed", "error", err)
		return
	}
	p.state = types.Playing
	p.stateChan <- p.state
}

func (p *dlnaPlayer) Stop() {
	if p.controlUrl == "" {
		return
	}
	if err := p.soapRequest("AVTransport", "Stop", stopBody); err != nil {
		slog.Error("DLNA: Stop failed", "error", err)
		return
	}
	p.curPos = 0
	p.state = types.Stopped
	p.stateChan <- p.state
}

func (p *dlnaPlayer) Toggle() {
	if p.state == types.Playing {
		p.Pause()
	} else {
		p.Resume()
	}
}

func (p *dlnaPlayer) Seek(duration time.Duration) {
	if p.controlUrl == "" {
		return
	}
	seekTime := formatDuration(duration)
	if err := p.soapRequest("AVTransport", "Seek", fmt.Sprintf(seekBody, seekTime)); err != nil {
		slog.Debug("DLNA: Seek not supported", "error", err)
	}
}

func (p *dlnaPlayer) PassedTime() time.Duration {
	return p.curPos
}

func (p *dlnaPlayer) PlayedTime() time.Duration { return 0 }

func (p *dlnaPlayer) TimeChan() <-chan time.Duration {
	return p.timeChan
}

func (p *dlnaPlayer) State() types.State { return p.state }

func (p *dlnaPlayer) StateChan() <-chan types.State { return p.stateChan }

func (p *dlnaPlayer) getVolume() (int, error) {
	if p.renderingControlUrl == "" {
		return 0, errors.New("RenderingControl service not available")
	}
	body := getVolumeBody
	envelope := fmt.Sprintf(soapEnvelopeTpl, body)
	req, err := http.NewRequest("POST", p.renderingControlUrl, bytes.NewBufferString(envelope))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", `text/xml; charset="utf-8"`)
	req.Header.Set("SOAPAction", `"urn:schemas-upnp-org:service:RenderingControl:1#GetVolume"`)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return 0, fmt.Errorf("GetVolume failed: %d", resp.StatusCode)
	}

	respBody, _ := io.ReadAll(resp.Body)

	type volumeResponse struct {
		CurrentVolume string `xml:"CurrentVolume"`
	}

	type envelopeResponse struct {
		Body struct {
			GetVolumeResponse volumeResponse `xml:"GetVolumeResponse"`
		} `xml:"Body"`
	}

	var env envelopeResponse
	if err := xml.Unmarshal(respBody, &env); err != nil {
		slog.Debug("DLNA: failed to parse volume", "error", err, "body", string(respBody))
		return 0, err
	}

	volume, _ := strconv.Atoi(env.Body.GetVolumeResponse.CurrentVolume)
	return volume, nil
}

func (p *dlnaPlayer) setVolume(volume int) error {
	if p.renderingControlUrl == "" {
		return errors.New("RenderingControl service not available")
	}
	body := fmt.Sprintf(setVolumeBody, volume)
	envelope := fmt.Sprintf(soapEnvelopeTpl, body)
	req, err := http.NewRequest("POST", p.renderingControlUrl, bytes.NewBufferString(envelope))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", `text/xml; charset="utf-8"`)
	req.Header.Set("SOAPAction", `"urn:schemas-upnp-org:service:RenderingControl:1#SetVolume"`)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("SetVolume failed: %d", resp.StatusCode)
	}

	return nil
}

func (p *dlnaPlayer) Volume() int {
	vol, err := p.getVolume()
	if err != nil {
		slog.Debug("DLNA: failed to get volume", "error", err)
		return 0
	}
	return vol
}

func (p *dlnaPlayer) SetVolume(volume int) {
	if volume < 0 {
		volume = 0
	}
	if volume > 100 {
		volume = 100
	}
	if err := p.setVolume(volume); err != nil {
		slog.Error("DLNA: SetVolume failed", "error", err)
	}
}

func (p *dlnaPlayer) UpVolume() {
	current := p.Volume()
	if current < 100 {
		p.SetVolume(current + 1)
	}
}

func (p *dlnaPlayer) DownVolume() {
	current := p.Volume()
	if current > 0 {
		p.SetVolume(current - 1)
	}
}

func (p *dlnaPlayer) Close() {
	close(p.closed)
	if p.httpServer != nil {
		p.httpServer.Close()
	}
}

func formatDuration(d time.Duration) string {
	totalSeconds := d.Seconds()
	hours := int(totalSeconds) / 3600
	minutes := (int(totalSeconds) % 3600) / 60
	seconds := totalSeconds - float64(hours*3600+minutes*60)
	return fmt.Sprintf("%02d:%02d:%06.3f", hours, minutes, seconds)
}
