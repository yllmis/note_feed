package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	VaultPath string      `yaml:"vault_path"`
	LLM       LLMConfig   `yaml:"llm"`
	Search    SearchConfig `yaml:"search"`
	Push      PushConfig  `yaml:"push"`
	DB        DBConfig    `yaml:"db"`
}

type LLMConfig struct {
	APIKey  string `yaml:"api_key"`
	Model   string `yaml:"model"`
	Timeout string `yaml:"timeout"`
	BaseURL string `yaml:"base_url"`
}

type SearchConfig struct {
	Google GoogleConfig `yaml:"google"`
}

type GoogleConfig struct {
	APIKey string `yaml:"api_key"`
	CseID  string `yaml:"cse_id"`
}

type PushConfig struct {
	Email EmailConfig `yaml:"email"`
}

type EmailConfig struct {
	SMTPHost string `yaml:"smtp_host"`
	SMTPPort int    `yaml:"smtp_port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
	To       string `yaml:"to"`
}

type DBConfig struct {
	Path string `yaml:"path"`
}

func Default() *Config {
	return &Config{
		VaultPath: "",
		LLM: LLMConfig{
			Model:   "deepseek-chat",
			Timeout: "30s",
			BaseURL: "https://api.deepseek.com",
		},
		Search: SearchConfig{},
		Push: PushConfig{
			Email: EmailConfig{
				SMTPHost: "smtp.qq.com",
				SMTPPort: 587,
			},
		},
		DB: DBConfig{
			Path: "./data/push.db",
		},
	}
}

func Load(path string) (*Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	expanded := os.ExpandEnv(string(data))

	if err := yaml.Unmarshal([]byte(expanded), cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	cfg.LLM.APIKey = resolveEnvRef(cfg.LLM.APIKey)
	cfg.Push.Email.Username = resolveEnvRef(cfg.Push.Email.Username)
	cfg.Push.Email.Password = resolveEnvRef(cfg.Push.Email.Password)
	cfg.Push.Email.From = resolveEnvRef(cfg.Push.Email.From)
	cfg.Push.Email.To = resolveEnvRef(cfg.Push.Email.To)
	cfg.Search.Google.APIKey = resolveEnvRef(cfg.Search.Google.APIKey)
	cfg.Search.Google.CseID = resolveEnvRef(cfg.Search.Google.CseID)

	return cfg, nil
}

func resolveEnvRef(v string) string {
	if strings.HasPrefix(v, "${") && strings.HasSuffix(v, "}") {
		name := strings.TrimSuffix(strings.TrimPrefix(v, "${"), "}")
		if val := os.Getenv(name); val != "" {
			return val
		}
	}
	return v
}
