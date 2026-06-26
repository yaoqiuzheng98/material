package xhs

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// ShareLink 从小红书分享文本中解析出的笔记链接信息。
type ShareLink struct {
	NoteID     string // 笔记 ID，如 6a182c5000000000060378ae
	XSecToken  string // xsec_token
	XSecSource string // xsec_source，如 pc_share / pc_search
	URL        string // 解析出的完整 URL
}

// 匹配小红书笔记 URL 的几种常见路径形态：
//   /discovery/item/{id}
//   /explore/{id}
//   /item/{id}
var noteIDPattern = regexp.MustCompile(`(?:/discovery/item|/explore|/item)/([A-Za-z0-9]+)`)

// urlPattern 用于从分享文本中提取第一条 http(s) 链接。
var urlPattern = regexp.MustCompile(`https?://[^\s）】]+`)

// ParseShareText 从小红书分享文本中解析笔记链接信息。
//
// 支持的输入形态：
//   - 直接是 URL
//   - 包含 URL 的分享文本（如 "67 【标题】 😆 JD5U1VyvHbeWo1v 😆 https://..."）
//
// 返回的 ShareLink 包含 note_id、xsec_token、xsec_source。
func ParseShareText(text string) (*ShareLink, error) {
	rawURL := extractURL(text)
	if rawURL == "" {
		return nil, fmt.Errorf("no xiaohongshu url found in share text")
	}
	return ParseShareURL(rawURL)
}

// ParseShareURL 解析小红书笔记 URL。
func ParseShareURL(rawURL string) (*ShareLink, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}

	// 从路径中提取 note_id
	m := noteIDPattern.FindStringSubmatch(u.Path)
	if len(m) < 2 {
		return nil, fmt.Errorf("cannot extract note_id from path %q", u.Path)
	}
	noteID := m[1]

	q := u.Query()
	sl := &ShareLink{
		NoteID:     noteID,
		XSecToken:  q.Get("xsec_token"),
		XSecSource: q.Get("xsec_source"),
		URL:        rawURL,
	}

	// 规范化：去掉可能的尾部空白
	sl.XSecToken = strings.TrimSpace(sl.XSecToken)
	sl.XSecSource = strings.TrimSpace(sl.XSecSource)

	if sl.NoteID == "" {
		return nil, fmt.Errorf("note_id is empty")
	}
	return sl, nil
}

// extractURL 从文本中提取第一条 http(s) 链接。
func extractURL(text string) string {
	m := urlPattern.FindString(text)
	return strings.TrimRight(m, ".,;!?)]】")
}
