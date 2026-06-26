package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client 是一个轻量的 OpenAI 兼容 Chat Completions 客户端，
// 适用于 DeepSeek、OpenAI、Moonshot、本地 vLLM 等服务。
type Client struct {
	BaseURL     string // 如 https://api.deepseek.com
	APIKey      string
	Model       string
	HTTP        *http.Client
	MaxTokens   int // 0 表示不限制
	Temperature float64
}

// NewClient 创建默认配置的 client。
func NewClient(baseURL, apiKey, model string) *Client {
	return &Client{
		BaseURL:     strings.TrimRight(baseURL, "/"),
		APIKey:      apiKey,
		Model:       model,
		HTTP:        &http.Client{Timeout: 60 * time.Second},
		MaxTokens:   1024,
		Temperature: 1.0,
	}
}

// Message 一条对话消息。
type Message struct {
	Role    string `json:"role"` // system / user / assistant
	Content string `json:"content"`
}

// chatRequest OpenAI Chat Completions 请求体。
type chatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	Stream      bool      `json:"stream"`
}

// chatResponse OpenAI Chat Completions 响应体。
type chatResponse struct {
	Choices []struct {
		Message      Message `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *APIError `json:"error,omitempty"`
}

// APIError OpenAI 兼容协议的错误结构。
type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    any    `json:"code"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("llm api error: %s (type=%s)", e.Message, e.Type)
}

// Chat 发起一次 Chat Completions 请求，返回助手回复内容。
func (c *Client) Chat(ctx context.Context, messages []Message) (string, error) {
	if c.HTTP == nil {
		c.HTTP = &http.Client{Timeout: 60 * time.Second}
	}

	reqBody := chatRequest{
		Model:       c.Model,
		Messages:    messages,
		MaxTokens:   c.MaxTokens,
		Temperature: c.Temperature,
		Stream:      false,
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	url := c.BaseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("accept", "application/json")
	if c.APIKey != "" {
		req.Header.Set("authorization", "Bearer "+c.APIKey)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var cr chatResponse
		_ = json.Unmarshal(body, &cr)
		if cr.Error != nil {
			return "", fmt.Errorf("status %d: %w", resp.StatusCode, cr.Error)
		}
		return "", fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	var cr chatResponse
	if err := json.Unmarshal(body, &cr); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if len(cr.Choices) == 0 {
		return "", fmt.Errorf("empty choices in response: %s", string(body))
	}
	return cr.Choices[0].Message.Content, nil
}

// GenerateShareCopy 根据小红书笔记内容生成分享文案。
//
// 入参 noteText 是已经整理好的笔记文本（标题 + 正文 + 标签 + 互动数据等）。
// 返回 LLM 生成的分享文案。
func (c *Client) GenerateShareCopy(ctx context.Context, noteText string) (string, error) {
	messages := []Message{
		{
			Role: "system",
			Content: "你是一名资深的内容运营，擅长根据小红书笔记内容生成适合二次传播的分享文案。" +
				"要求：1) 抓住笔记核心卖点或情绪点；2) 风格活泼、有网感、带emoji；" +
				"3) 控制在150字以内；4) 结尾可带1-2个相关话题标签；5) 只输出文案本身，不要解释。",
		},
		{
			Role:    "user",
			Content: "请根据以下小红书笔记内容，生成一段分享文案：\n\n" + noteText,
		},
	}
	return c.Chat(ctx, messages)
}
