package player

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
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

type dlnaPlayer struct {
	deviceUrl  string
	controlUrl string
	audioUrl   string
	httpClient *http.Client
	state      types.State
	stateChan  chan types.State
	musicChan  chan URLMusic
	closed     chan struct{}
	ready      chan struct{}
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
}

func (p *dlnaPlayer) CurMusic() URLMusic {
	return URLMusic{URL: p.audioUrl}
}

func (p *dlnaPlayer) Pause() {
	p.state = types.Paused
	p.stateChan <- p.state
}

func (p *dlnaPlayer) Resume() {
	p.state = types.Playing
	p.stateChan <- p.state
}

func (p *dlnaPlayer) Stop() {
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

func (p *dlnaPlayer) Seek(duration time.Duration) {}

func (p *dlnaPlayer) PassedTime() time.Duration { return 0 }

func (p *dlnaPlayer) PlayedTime() time.Duration { return 0 }

func (p *dlnaPlayer) TimeChan() <-chan time.Duration {
	return make(chan time.Duration)
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
