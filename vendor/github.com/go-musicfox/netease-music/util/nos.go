package util

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// NosUploadFile 是上传接口接收的文件数据。
type NosUploadFile struct {
	Name     string
	MimeType string
	Data     []byte
	MD5      string
	Size     int64
}

// Prepare 计算缺失的 MD5 和大小。
func (f *NosUploadFile) Prepare() {
	if f.MD5 == "" && len(f.Data) > 0 {
		sum := md5.Sum(f.Data)
		f.MD5 = hex.EncodeToString(sum[:])
	}
	if f.Size == 0 {
		f.Size = int64(len(f.Data))
	}
}

// FileExtension 返回小写扩展名，默认 mp3。
func FileExtension(name string) string {
	if i := strings.LastIndex(name, "."); i >= 0 && i+1 < len(name) {
		return strings.ToLower(name[i+1:])
	}
	return "mp3"
}

// SanitizeFilename 去掉扩展名、空格和点。
func SanitizeFilename(name string) string {
	if i := strings.LastIndex(name, "."); i >= 0 {
		name = name[:i]
	}
	name = strings.ReplaceAll(name, " ", "")
	name = strings.ReplaceAll(name, ".", "_")
	if name == "" {
		return "unknown"
	}
	return name
}

type nosTokenResult struct {
	Result struct {
		ObjectKey  string `json:"objectKey"`
		Token      string `json:"token"`
		DocID      string `json:"docId"`
		ResourceID string `json:"resourceId"`
	} `json:"result"`
}

// AllocNosToken 申请 NOS 上传 token。
func AllocNosToken(options *Options, data map[string]string) (*nosTokenResult, []byte, error) {
	code, body, _ := CreateRequest("POST", "https://music.163.com/api/nos/token/alloc", data, options)
	if code != 200 {
		return nil, body, fmt.Errorf("nos token alloc failed: %s", string(body))
	}
	var result nosTokenResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, body, err
	}
	if result.Result.ObjectKey == "" || result.Result.Token == "" {
		return nil, body, fmt.Errorf("nos token alloc empty: %s", string(body))
	}
	return &result, body, nil
}

type nosLBSResult struct {
	Upload []string `json:"upload"`
}

// NosUploadServer 获取 NOS 上传服务器地址。
func NosUploadServer(bucket string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://wanproxy.127.net/lbs?version=1.0&bucketname=" + bucket)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result nosLBSResult
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	if len(result.Upload) == 0 || result.Upload[0] == "" {
		return "", fmt.Errorf("nos lbs empty: %s", string(body))
	}
	return strings.TrimRight(result.Upload[0], "/"), nil
}

// NosPost 向 NOS 地址发送原始字节并返回响应体和头。
func NosPost(url string, data []byte, headers map[string]string) ([]byte, http.Header, error) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return body, resp.Header, fmt.Errorf("nos upload failed: status=%d body=%s", resp.StatusCode, string(body))
	}
	return body, resp.Header, nil
}

// NosPut 向 NOS 分片地址发送数据。
func NosPut(url string, data []byte, headers map[string]string) ([]byte, http.Header, error) {
	req, err := http.NewRequest("PUT", url, bytes.NewReader(data))
	if err != nil {
		return nil, nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return body, resp.Header, fmt.Errorf("nos put failed: status=%d body=%s", resp.StatusCode, string(body))
	}
	return body, resp.Header, nil
}
