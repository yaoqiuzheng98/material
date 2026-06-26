package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Config 顶层配置结构。
type Config struct {
	XHS XHSConfig `toml:"xhs"`
	LLM LLMConfig `toml:"llm"`
}

// XHSConfig 小红书请求相关配置。
//
// 注意：x-s / x-s-common / x-t / x-rap-param 等签名头由小红书前端 JS 生成且会过期，
// 需要定期更新或自行实现签名逻辑。这里把它们放在配置里方便替换。
type XHSConfig struct {
	Cookie string `toml:"cookie"`

	XS          string `toml:"x_s"`
	XSCommon    string `toml:"x_s_common"`
	XT          string `toml:"x_t"`
	XB3TraceID  string `toml:"x_b3_traceid"`
	XRapParam   string `toml:"x_rap_param"`
	XXrayTrace  string `toml:"x_xray_traceid"`
	XYDirection string `toml:"xy_direction"`
}

// LLMConfig LLM 服务配置（DeepSeek / OpenAI 兼容协议）。
type LLMConfig struct {
	BaseURL string `toml:"base_url"` // DeepSeek: https://api.deepseek.com
	APIKey  string `toml:"api_key"`
	Model   string `toml:"model"` // 如 deepseek-chat / deepseek-reasoner
}

// Load 从 TOML 文件加载配置。
func Load(path string) (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("decode toml %q: %w", path, err)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// LoadFromEnv 在配置文件缺失时，尝试从环境变量加载（便于 CI / 临时运行）。
func LoadFromEnv() *Config {
	return &Config{
		LLM: LLMConfig{
			BaseURL: envOr("LLM_BASE_URL", "https://api.deepseek.com"),
			APIKey:  os.Getenv("LLM_API_KEY"),
			Model:   envOr("LLM_MODEL", "deepseek-chat"),
		},
	}
}

func (c *Config) validate() error {
	if c.LLM.APIKey == "" {
		return fmt.Errorf("llm.api_key is required")
	}
	if c.LLM.BaseURL == "" {
		c.LLM.BaseURL = "https://api.deepseek.com"
	}
	if c.LLM.Model == "" {
		c.LLM.Model = "deepseek-chat"
	}
	return nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
