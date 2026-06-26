package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Config 顶层配置结构。
//
// 注意：XHS 部分写死在 DefaultXHSConfig() 中，不来自配置文件。
type Config struct {
	XHS XHSConfig
	LLM LLMConfig `toml:"llm"`
}

// XHSConfig 小红书请求相关配置。
//
// x-s / x-s-common / x-t / x-rap-param 等签名头由小红书前端 JS 生成且会过期，
// 当前值写死在 DefaultXHSConfig() 中，来自浏览器抓包，过期后需要替换。
type XHSConfig struct {
	Cookie string

	XS          string
	XSCommon    string
	XT          string
	XB3TraceID  string
	XRapParam   string
	XXrayTrace  string
	XYDirection string
}

// DefaultXHSConfig 返回写死的小红书签名头 / Cookie（来自浏览器抓包，会过期）。
func DefaultXHSConfig() XHSConfig {
	return XHSConfig{
		Cookie:      "abRequestId=1315df89-4bc7-532d-9d80-d1d6480c3aa9; ets=1781684307876; xsecappid=xhs-pc-web; a1=19ed4a9080bsloqb6wreg43lrnaxughy2gb0vqq4o50000413861; webId=40ecd5e69ffbf1dd0470fd80e4103b06; gid=yjdf40jyifJ4yjdf40j8YkAA8DM1lADKEFdkD4l4q1FI0828C7kxvh8884yqYKy82yyWyK4S; webBuild=6.25.1; acw_tc=0ad5963317824594860804957eee1bdaf592186b2310f258b0b20a112f41ad; loadts=1782459700370; web_session=040069b8e8af29c21f3c10650b384b41d7f7d5; id_token=VjEAAFOT5kII98X2WQm3gJIHhBX8jim4TXB5x8LFTQHMiVmqoOT8l4etX6K5wuMVI0CokkVeb+r1LL25EBfxK1lorIC38ZLXZ+05MwKvhwlSmQb61TvtD2UDBMthdPWLSP9ikDWn; x-rednote-datactry=CN; x-rednote-holderctry=CN; unread={%22ub%22:%226a2bd36c0000000035031b8c%22%2C%22ue%22:%226a27ebc7000000000702fc65%22%2C%22uc%22:26}; websectiga=16f444b9ff5e3d7e258b5f7674489196303a0b160e16647c6c2b4dcb609f4134; sec_poison_id=fb1cf07d-9a46-48b4-93cb-fbd9f41afc30; acw_tc=0ad5963317824603666343204eee1964b7c5451a3fb8d046facc11338f210a",
		XS:          "XYS_2UQhPsHCH0c1PUhMHjIj2erjwjQhyoPTqBPT49pjHjIj2eHjwjQgynEDJ74AHjIj2ePjwjQTJdPIPAZlg94aGLTlGDEz+dbYyF4/LpQb4emk8o8oGpprL7+awepn/sRx2bSonLDUy0bA+FDF8Bhh4gH9zBWEq9GILgSI+sTtaDYhcFRL40blJ08M4rMPzfkaJS+gcS8laeSD/BEm8FEhc9+PP9bGJFbSzgq7qLTBaFTmPrkHaMY/4bYPwoi7PaT+c9EIqMQCLDkcpnbLP9lsLDT/Jfznnfl0yLLIaSQQyAmOarEaLSz+GAQQ+/4s87zOyfrAP/YH+/bmzf4UpjHVHdWFH0ijJ9Qx8n+FHdF=",
		XSCommon:    "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",
		XT:          "1782460308397",
		XB3TraceID:  "39db3761e422e11c",
		XRapParam:   "ByQBBgAAAAEAAAAUAAACBC5BcBcAACg8AAAAIQAAAAAAAAAAbHhocXliSsgKv5bfIBDOPuIsFwPmTQAAABC2na2hXRhaf6h5fgLiySpfr+QUayOFBenFGP8aRQWrQxyyeT4FdqkOCr2/Eka7xghRUv6LEZvuhAedPpC8Ou02/lJCs/0WZvTnxWG6ZeYfndJbv85FgLp+Dt7UWrfG0Lw49JiNV7SO5IlKafsjkX3g9BtgbHFMhPdHQT3OUmBQ8PwleNSCaSBT0bX8jy8qXOI5zYmobCd1rUX7ZV+/bHZ0belkOPhym8U4HMAn4DVbGYM6yd22dAhTzUCiILxseQ+XCyibEceBgDvVxQ51l5gCIMjydaa/4Ret0mX1m9R1u9wdR4yUECJBjT6YqrqNbDqWVYXdsbmpArPRiZxGUXJ0KioYmX8VkmrFwENS4sFczcVZGIbi+Zkt9bWAoN8PKwp/91f7hB1W/OioPosPsj5reg6Us7o5zh9JbRvldRuMscSZ9pIEb8VWt+4NXIB8vsk5DZDq5budaIgMfQmIxUewVtmnVc+clIUxkBJKgBVI5ixq7b0HOd6k6qeJUBpldACUvKWBaC15J0R3nxrZ4jU3cGSYQf3SzvDiXKjRHmYKBKssYxc9xuFZcM36shsWtlw5SGmsBguef6M/YX3E535FBohOichdugqg1FOOSgDxV9T7mh4TlUUOLSwodM94mLnQS4n4JHIknV2x/z8EvNbcx3/HS2IRjN82GUlcFgVPEgAAAf8=",
		XXrayTrace:  "cf8174edbdb0f0db79a6765935ee130f",
		XYDirection: "59",
	}
}

// LLMConfig LLM 服务配置（DeepSeek / OpenAI 兼容协议）。
type LLMConfig struct {
	BaseURL string `toml:"base_url"` // DeepSeek: https://api.deepseek.com
	APIKey  string `toml:"api_key"`
	Model   string `toml:"model"` // 如 deepseek-chat / deepseek-reasoner
}

// Load 从 TOML 文件加载配置。XHS 部分使用写死的默认值。
func Load(path string) (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("decode toml %q: %w", path, err)
	}
	cfg.XHS = DefaultXHSConfig()
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// LoadFromEnv 在配置文件缺失时，尝试从环境变量加载（便于 CI / 临时运行）。
// XHS 部分使用写死的默认值。
func LoadFromEnv() *Config {
	return &Config{
		XHS: DefaultXHSConfig(),
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
