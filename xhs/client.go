package xhs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	FeedAPIURL = "https://edith.xiaohongshu.com/api/sns/web/v1/feed"
	Origin     = "https://www.xiaohongshu.com"
)

// Client 用于请求小红书笔记详情接口。
//
// 注意：x-s / x-s-common / x-t / x-rap-param 等是签名头，
// 由小红书前端 JS 生成且会过期，需要外部注入或自行实现签名逻辑。
type Client struct {
	HTTP *http.Client

	// Cookie 浏览器登录后的 Cookie 字符串
	Cookie string

	// 签名相关头部，每次请求前需要更新
	XS         string
	XSCommon   string
	XT         string
	XB3TraceID string
	XRapParam  string

	// 可选，未设置时使用随机值或留空
	XXrayTraceID string
	XYDirection  string
}

// NewClient 创建一个默认配置的 Client。
func NewClient() *Client {
	return &Client{
		HTTP: &http.Client{Timeout: 15 * time.Second},
	}
}

// FeedRequest 笔记详情请求体。
type FeedRequest struct {
	SourceNoteID string   `json:"source_note_id"`
	ImageFormats []string `json:"image_formats"`
	Extra        *Extra   `json:"extra,omitempty"`
	XSecSource   string   `json:"xsec_source,omitempty"`
	XSecToken    string   `json:"xsec_token,omitempty"`
}

// Extra 请求附加参数。
type Extra struct {
	NeedBodyTopic string `json:"need_body_topic,omitempty"`
}

// NewFeedRequest 构造一个常用的笔记详情请求体。
func NewFeedRequest(sourceNoteID, xsecSource, xsecToken string) *FeedRequest {
	return &FeedRequest{
		SourceNoteID: sourceNoteID,
		ImageFormats: []string{"jpg", "webp", "avif"},
		Extra:        &Extra{NeedBodyTopic: "1"},
		XSecSource:   xsecSource,
		XSecToken:    xsecToken,
	}
}

// GetNoteDetail 请求笔记详情。
//
// 参数:
//   - sourceNoteID: 笔记 ID
//   - xsecSource: 来源，如 "pc_search"
//   - xsecToken: 笔记的 xsec_token
func (c *Client) GetNoteDetail(ctx context.Context, sourceNoteID, xsecSource, xsecToken string) ([]byte, error) {
	reqBody := NewFeedRequest(sourceNoteID, xsecSource, xsecToken)
	return c.GetNoteDetailWithBody(ctx, reqBody)
}

// GetNoteDetailWithBody 使用自定义请求体请求笔记详情，返回原始响应。
func (c *Client) GetNoteDetailWithBody(ctx context.Context, reqBody *FeedRequest) ([]byte, error) {
	if c.HTTP == nil {
		c.HTTP = &http.Client{Timeout: 15 * time.Second}
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, FeedAPIURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return body, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

// setHeaders 设置与 curl 一致的请求头。
func (c *Client) setHeaders(req *http.Request) {
	h := req.Header
	h.Set("accept", "application/json, text/plain, */*")
	h.Set("accept-language", "zh-CN,zh;q=0.9")
	h.Set("content-type", "application/json;charset=UTF-8")
	h.Set("origin", Origin)
	h.Set("priority", "u=1, i")
	h.Set("referer", Origin+"/")
	h.Set("sec-ch-ua", `"Google Chrome";v="149", "Chromium";v="149", "Not)A;Brand";v="24"`)
	h.Set("sec-ch-ua-mobile", "?0")
	h.Set("sec-ch-ua-platform", `"Windows"`)
	h.Set("sec-fetch-dest", "empty")
	h.Set("sec-fetch-mode", "cors")
	h.Set("sec-fetch-site", "same-site")
	h.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Safari/537.36")

	// 签名 / 链路追踪头
	if c.XS != "" {
		h.Set("x-s", c.XS)
	}
	if c.XSCommon != "" {
		h.Set("x-s-common", c.XSCommon)
	}
	if c.XT != "" {
		h.Set("x-t", c.XT)
	}
	if c.XB3TraceID != "" {
		h.Set("x-b3-traceid", c.XB3TraceID)
	}
	if c.XRapParam != "" {
		h.Set("x-rap-param", c.XRapParam)
	}
	if c.XXrayTraceID != "" {
		h.Set("x-xray-traceid", c.XXrayTraceID)
	}
	if c.XYDirection != "" {
		h.Set("xy-direction", c.XYDirection)
	}

	if c.Cookie != "" {
		req.Header.Set("Cookie", c.Cookie)
	}
}
