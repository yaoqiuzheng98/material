package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"material/config"
	"material/llm"
	"material/xhs"
)

func main() {
	shareText := flag.String("share", "", "小红书分享文本或链接（包含 note_id 和 xsec_token）")
	cfgPath := flag.String("config", "config.toml", "配置文件路径")
	flag.Parse()

	if *shareText == "" {
		log.Fatal("请通过 -share 传入小红书分享文本或链接")
	}

	// 1. 加载配置
	cfg, err := loadConfig(*cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// 2. 从分享文本中解析 note_id / xsec_token / xsec_source
	link, err := xhs.ParseShareText(*shareText)
	if err != nil {
		log.Fatalf("parse share text: %v", err)
	}
	fmt.Printf("解析链接成功:\n  note_id     = %s\n  xsec_token  = %s\n  xsec_source = %s\n\n",
		link.NoteID, link.XSecToken, link.XSecSource)

	// 3. 构造小红书 client 并注入签名头 / Cookie
	xhsClient := xhs.NewClient()
	applyXHSConfig(xhsClient, &cfg.XHS)

	// 4. 请求笔记详情
	body, err := xhsClient.GetNoteDetail(context.Background(),
		link.NoteID, link.XSecSource, link.XSecToken)
	if err != nil {
		log.Fatalf("get note detail: %v", err)
	}

	// 5. 解析笔记详情
	feed, err := xhs.ParseFeedResponse(body)
	if err != nil {
		log.Fatalf("parse feed response: %v", err)
	}
	note := feed.FirstNote()
	if note == nil {
		log.Fatal("feed response has no note")
	}
	noteText := note.ToLLMText()
	fmt.Printf("笔记内容：\n%s\n\n", noteText)

	// 6. 调用 LLM 生成分享文案
	llmClient := llm.NewClient(cfg.LLM.BaseURL, cfg.LLM.APIKey, cfg.LLM.Model)
	fmt.Println("正在生成分享文案...")
	copyText, err := llmClient.GenerateShareCopy(context.Background(), noteText)
	if err != nil {
		log.Fatalf("generate share copy: %v", err)
	}

	fmt.Println("\n========== 分享文案 ==========")
	fmt.Println(copyText)
	fmt.Println("==============================")
}

// loadConfig 优先从配置文件加载；文件不存在时回退到环境变量。
func loadConfig(path string) (*config.Config, error) {
	if _, err := os.Stat(path); err == nil {
		return config.Load(path)
	}
	// 尝试与可执行文件同目录
	if abs, err := filepath.Abs(path); err == nil {
		if _, err := os.Stat(abs); err == nil {
			return config.Load(abs)
		}
	}
	fmt.Fprintf(os.Stderr, "配置文件 %q 不存在，回退到环境变量\n", path)
	cfg := config.LoadFromEnv()
	if cfg.LLM.APIKey == "" {
		return nil, fmt.Errorf("未找到配置文件，且环境变量 LLM_API_KEY 未设置")
	}
	return cfg, nil
}

// applyXHSConfig 把配置中的签名头 / Cookie 注入到 client。
func applyXHSConfig(c *xhs.Client, x *config.XHSConfig) {
	c.Cookie = x.Cookie
	c.XS = x.XS
	c.XSCommon = x.XSCommon
	c.XT = x.XT
	c.XB3TraceID = x.XB3TraceID
	c.XRapParam = x.XRapParam
	c.XXrayTraceID = x.XXrayTrace
	c.XYDirection = x.XYDirection
}
