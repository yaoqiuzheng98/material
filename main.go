package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"material/config"
	"material/ui"
)

func main() {
	// 加载配置
	cfg, err := loadConfig("config.toml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// 创建 Fyne 应用
	a := app.New()
	w := a.NewWindow("小红书分享文案生成器")
	w.Resize(fyne.NewSize(640, 600))

	gui := ui.NewApp(cfg)
	w.SetContent(gui.Render(w))
	w.SetMaster()

	w.ShowAndRun()
}

// loadConfig 优先从配置文件加载；文件不存在时回退到环境变量。
func loadConfig(path string) (*config.Config, error) {
	if _, err := os.Stat(path); err == nil {
		return config.Load(path)
	}
	if abs, err := filepath.Abs(path); err == nil {
		if _, err := os.Stat(abs); err == nil {
			return config.Load(abs)
		}
	}
	fmt.Fprintln(os.Stderr, "配置文件", path, "不存在，回退到环境变量")
	cfg := config.LoadFromEnv()
	if cfg.LLM.APIKey == "" {
		return nil, fmt.Errorf("未找到配置文件，且环境变量 LLM_API_KEY 未设置")
	}
	return cfg, nil
}
