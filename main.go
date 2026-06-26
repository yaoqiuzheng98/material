package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"material/config"
	"material/web"
)

func main() {
	// 加载配置
	cfg, err := loadConfig("config.toml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// 启动 HTTP server
	addr := ":8080"
	fmt.Printf("服务已启动，请在浏览器打开 http://localhost%s\n", addr)
	if err := web.Serve(addr, cfg); err != nil {
		log.Fatalf("server error: %v", err)
	}
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
