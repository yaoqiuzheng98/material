package xhs

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FeedResponse /api/sns/web/v1/feed 的响应结构。
type FeedResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Items []FeedItem `json:"items"`
	} `json:"data"`
}

// FeedItem 一条笔记条目。
type FeedItem struct {
	ID        string   `json:"id"`
	XSecToken string   `json:"xsec_token"`
	NoteCard  NoteCard `json:"note_card"`
}

// NoteCard 笔记详情卡片。
type NoteCard struct {
	Type         string       `json:"type"` // normal / video
	Title        string       `json:"title"`
	Desc         string       `json:"desc"`
	NoteID       string       `json:"note_id"`
	User         NoteUser     `json:"user"`
	InteractInfo InteractInfo `json:"interact_info"`
	TagList      []Tag        `json:"tag_list"`
	ImageList    []NoteImage  `json:"image_list"`
	Video        *NoteVideo   `json:"video,omitempty"`
}

// NoteUser 笔记作者。
type NoteUser struct {
	UserID   string `json:"user_id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// InteractInfo 互动数据。
type InteractInfo struct {
	LikedCount     string `json:"liked_count"`
	CollectedCount string `json:"collected_count"`
	CommentCount   string `json:"comment_count"`
	ShareCount     string `json:"share_count"`
}

// Tag 标签。
type Tag struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// NoteImage 图片。
type NoteImage struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// NoteVideo 视频。
type NoteVideo struct {
	Media struct {
		Stream struct {
			HLS []struct {
				MasterURL string `json:"master_url"`
			} `json:"hls"`
		} `json:"stream"`
	} `json:"media"`
}

// ParseFeedResponse 解析 feed 接口的原始 JSON 响应。
func ParseFeedResponse(body []byte) (*FeedResponse, error) {
	var fr FeedResponse
	if err := json.Unmarshal(body, &fr); err != nil {
		return nil, fmt.Errorf("unmarshal feed response: %w", err)
	}
	if fr.Code != 0 && fr.Code != -1 {
		// 小红书成功时 code 通常为 0，部分场景为 -1 但仍返回数据
		return nil, fmt.Errorf("feed api code=%d msg=%s", fr.Code, fr.Msg)
	}
	if len(fr.Data.Items) == 0 {
		return nil, fmt.Errorf("feed api returned no items: %s", string(body))
	}
	return &fr, nil
}

// FirstNote 取响应中的第一条笔记卡片。
func (fr *FeedResponse) FirstNote() *NoteCard {
	if len(fr.Data.Items) == 0 {
		return nil
	}
	return &fr.Data.Items[0].NoteCard
}

// ToLLMText 把笔记内容整理成适合喂给 LLM 的纯文本。
//
// 包含：标题、正文、作者、互动数据、标签、图片数量（视频笔记则标注为视频）。
func (nc *NoteCard) ToLLMText() string {
	var b strings.Builder
	if nc.Title != "" {
		fmt.Fprintf(&b, "标题：%s\n", nc.Title)
	}
	if nc.Desc != "" {
		fmt.Fprintf(&b, "正文：%s\n", nc.Desc)
	}
	if nc.User.Nickname != "" {
		fmt.Fprintf(&b, "作者：%s\n", nc.User.Nickname)
	}
	fmt.Fprintf(&b, "类型：%s\n", nc.Type)

	// 互动数据
	ii := nc.InteractInfo
	if ii.LikedCount != "" || ii.CollectedCount != "" || ii.CommentCount != "" || ii.ShareCount != "" {
		fmt.Fprintf(&b, "互动数据：点赞 %s / 收藏 %s / 评论 %s / 分享 %s\n",
			ii.LikedCount, ii.CollectedCount, ii.CommentCount, ii.ShareCount)
	}

	// 标签
	if len(nc.TagList) > 0 {
		names := make([]string, 0, len(nc.TagList))
		for _, t := range nc.TagList {
			if t.Name != "" {
				names = append(names, "#"+t.Name)
			}
		}
		if len(names) > 0 {
			fmt.Fprintf(&b, "标签：%s\n", strings.Join(names, " "))
		}
	}

	// 媒体信息
	if nc.Type == "video" {
		b.WriteString("媒体：视频笔记\n")
	} else if len(nc.ImageList) > 0 {
		fmt.Fprintf(&b, "媒体：%d 张图片\n", len(nc.ImageList))
	}

	return strings.TrimRight(b.String(), "\n")
}
