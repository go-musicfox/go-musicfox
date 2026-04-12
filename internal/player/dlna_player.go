package player

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
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

type dlnaPlayer struct {
	deviceUrl  string
	controlUrl string
	audioUrl   string
	audioDur   time.Duration
	httpClient *http.Client
	state      types.State
	stateChan  chan types.State
	musicChan  chan URLMusic
	closed     chan struct{}
	ready      chan struct{}

	curPos   time.Duration
	timeChan chan time.Duration
}

func NewDlnaPlayer(deviceUrl string) *dlnaPlayer {
	p := &dlnaPlayer{
		deviceUrl:  deviceUrl,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		state:      types.Stopped,
		stateChan:  make(chan types.State, 10),
		musicChan:  make(chan URLMusic, 10),
		closed:     make(chan struct{}),
		ready:      make(chan struct{}),
		timeChan:   make(chan time.Duration, 1),
	}
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
		if svc.ServiceType == "urn:schemas-upnp-org:service:AVTransport:1" {
			controlUrl := svc.ControlURL
			if controlUrl != "" && !bytes.HasPrefix([]byte(controlUrl), []byte("http")) {
				controlUrl = base + controlUrl
			}
			p.controlUrl = controlUrl
			slog.Info("DLNA: found AVTransport control URL", "url", controlUrl)
			return
		}
	}
	slog.Error("DLNA: AVTransport service not found")
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

func (p *dlnaPlayer) soapRequest(action, body string) error {
	envelope := fmt.Sprintf(soapEnvelopeTpl, body)
	req, err := http.NewRequest("POST", p.controlUrl, bytes.NewBufferString(envelope))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", `text/xml; charset="utf-8"`)
	req.Header.Set("SOAPAction", fmt.Sprintf(`"urn:schemas-upnp-org:service:AVTransport:1#%s"`, action))

	slog.Debug("DLNA: SOAP request", "action", action, "url", p.controlUrl)
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

func (p *dlnaPlayer) Play(music URLMusic) {
	p.audioUrl = music.URL
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
	if err := p.soapRequest("SetAVTransportURI", fmt.Sprintf(setAvTransportUriBody, p.audioUrl)); err != nil {
		slog.Error("DLNA: SetAVTransportURI failed", "error", err)
		return
	}

	slog.Info("DLNA: starting playback")
	if err := p.soapRequest("Play", playBody); err != nil {
		slog.Error("DLNA: Play failed", "error", err)
		return
	}

	p.state = types.Playing
	p.stateChan <- p.state

	go p.pollPositionInfo()
}

func (p *dlnaPlayer) pollPositionInfo() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			curPos, _, err := p.getPositionInfo()
			if err == nil && curPos > 0 {
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
	if err := p.soapRequest("Pause", pauseBody); err != nil {
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
	if err := p.soapRequest("Play", playBody); err != nil {
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
	if err := p.soapRequest("Stop", stopBody); err != nil {
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
	if err := p.soapRequest("Seek", fmt.Sprintf(seekBody, seekTime)); err != nil {
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

func (p *dlnaPlayer) Volume() int { return 0 }

func (p *dlnaPlayer) SetVolume(volume int) {}

func (p *dlnaPlayer) UpVolume() {}

func (p *dlnaPlayer) DownVolume() {}

func (p *dlnaPlayer) Close() {
	close(p.closed)
}

func formatDuration(d time.Duration) string {
	totalSeconds := d.Seconds()
	hours := int(totalSeconds) / 3600
	minutes := (int(totalSeconds) % 3600) / 60
	seconds := totalSeconds - float64(hours*3600+minutes*60)
	return fmt.Sprintf("%02d:%02d:%06.3f", hours, minutes, seconds)
}
