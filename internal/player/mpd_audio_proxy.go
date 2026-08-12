package player

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// audioProxyUserAgent 模拟浏览器 UA 请求上游, 规避网易云 CDN 对非浏览器
// 流式请求（如连接重置、忽略 Range）的异常行为。
const audioProxyUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36 Edg/124.0.0.0"

// mpdAudioProxy 是 MPD 引擎专用的本地音频代理。
//
// 背景: MPD 的 curl input 插件以 libcurl 默认 UA、无 Referer、并强制携带
// "Icy-Metadata: 1" 请求源站, 且仅凭响应头存在 Accept-Ranges 就判定流可
// seek——断线续传/seek 时发出的 Range 请求若被源站忽略（返回 200 全量数据),
// MPD 不校验 Content-Range 会静默接受, 导致字节错位。FLAC 帧有严格 CRC 校验,
// 错位即丢帧, 听感为杂音/爆音（MP3 解码器容错强, 故低音质正常）。下载为本地
// 文件后数据完整, 播放自然正常。
//
// 方案: MPD 改为从本代理拉流。代理向上游注入浏览器 UA + Referer（网易云域名),
// 并正确处理 Range: 上游忽略 Range 时, 代理自行丢弃偏移字节并按 206 对齐返回,
// 从根上消除数据错位。代理不可用时回退为直连（行为与旧版一致）。
type mpdAudioProxy struct {
	server  *http.Server
	baseURL string
	client  *http.Client
}

func newMPDAudioProxy() (*mpdAudioProxy, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("mpd audio proxy: listen failed: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	p := &mpdAudioProxy{
		baseURL: fmt.Sprintf("http://127.0.0.1:%d", port),
		client:  &http.Client{},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/play", p.serve)
	p.server = &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	go func() {
		if err := p.server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("mpd audio proxy: server error", "error", err)
		}
	}()
	return p, nil
}

// wrapURL 将源站 URL 包装为本代理地址, 供 mpd 通过 addid 拉流。
func (p *mpdAudioProxy) wrapURL(raw string) string {
	return p.baseURL + "/play?url=" + url.QueryEscape(raw)
}

// isOurs 判断 URL 是否已指向本代理, 避免重复包装。
func (p *mpdAudioProxy) isOurs(u string) bool {
	return strings.HasPrefix(u, p.baseURL)
}

func (p *mpdAudioProxy) Close() error {
	if p.server == nil {
		return nil
	}
	return p.server.Close()
}

func (p *mpdAudioProxy) serve(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("url")
	if target == "" {
		http.Error(w, "missing url query param", http.StatusBadRequest)
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, target, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.Header.Set("User-Agent", audioProxyUserAgent)
	if isNetEaseHost(target) {
		req.Header.Set("Referer", "https://music.163.com/")
	}
	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		http.Error(w, "upstream error: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 上游已正确处理 Range（206）, 或无 Range 请求: 原样透传。
	if resp.StatusCode == http.StatusPartialContent || rangeHeader == "" {
		copyAudioHeaders(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
		return
	}

	// 上游异常（403/404 等）: 原样返回状态码与响应。
	if resp.StatusCode != http.StatusOK {
		copyAudioHeaders(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
		return
	}

	// 上游忽略 Range 返回 200 全量数据: 代理自行丢弃偏移字节并按 206 对齐返回。
	offset, err := parseRangeStart(rangeHeader)
	if err != nil {
		copyAudioHeaders(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
		return
	}

	if total := resp.ContentLength; total > 0 {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", offset, total-1, total))
		w.Header().Set("Content-Length", strconv.FormatInt(total-offset, 10))
	}
	w.Header().Set("Accept-Ranges", "bytes")
	w.WriteHeader(http.StatusPartialContent)
	if _, err := io.CopyN(io.Discard, resp.Body, offset); err != nil {
		// 上游数据不足, 中止本次响应, 由 MPD 重试。
		return
	}
	_, _ = io.Copy(w, resp.Body)
}

// copyAudioHeaders 拷贝音频流相关的响应头。
func copyAudioHeaders(dst, src http.Header) {
	for _, k := range []string{"Content-Type", "Content-Length", "Accept-Ranges", "Content-Range"} {
		if v := src.Get(k); v != "" {
			dst.Set(k, v)
		}
	}
}

// parseRangeStart 解析单区间 Range 头 "bytes=N-" 的起始偏移。仅支持 MPD
// curl 插件实际会发送的形式, 其余形式返回错误由调用方走安全回退。
func parseRangeStart(rangeHeader string) (int64, error) {
	if !strings.HasPrefix(rangeHeader, "bytes=") {
		return 0, errors.New("unsupported range")
	}
	start, _, ok := strings.Cut(strings.TrimPrefix(rangeHeader, "bytes="), "-")
	if !ok || start == "" {
		return 0, errors.New("unsupported range")
	}
	offset, err := strconv.ParseInt(start, 10, 64)
	if err != nil || offset < 0 {
		return 0, errors.New("invalid range start")
	}
	return offset, nil
}

// isNetEaseHost 判断目标 URL 是否属于网易云音乐域名, 用于注入 Referer。
func isNetEaseHost(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	return strings.HasSuffix(host, "music.126.net") || strings.HasSuffix(host, "music.163.com")
}
