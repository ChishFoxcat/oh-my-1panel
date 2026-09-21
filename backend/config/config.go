package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config 服务配置，全部通过环境变量注入。
type Config struct {
	// HTTPAddr 本服务监听地址
	HTTPAddr string
	// PanelURL 1Panel 面板地址，例如 http://192.168.139.151:1314
	PanelURL string
	// PanelEntrance 面板安全入口（登录接口 EntranceCode 头的明文来源），可为空
	PanelEntrance string
	// PanelSkipTLSVerify 跳过面板 TLS 证书校验（面板自签名证书场景）
	PanelSkipTLSVerify bool
	// SessionTTL 浏览器会话有效期（滑动续期）
	SessionTTL time.Duration
	// PendingTTL MFA 待验证会话有效期
	PendingTTL time.Duration
	// CookieSecure 会话 Cookie 是否仅 HTTPS 传输
	CookieSecure bool
	// WebDir 前端静态资源目录
	WebDir string
}

// PanelBaseURL 面板地址的规范化形式（去除末尾斜杠、校验协议与主机）。
func (c *Config) PanelBaseURL() (*url.URL, error) {
	u, err := url.Parse(c.PanelURL)
	if err != nil {
		return nil, fmt.Errorf("面板地址解析失败：%w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, errors.New("面板地址必须以 http:// 或 https:// 开头")
	}
	if u.Host == "" {
		return nil, errors.New("面板地址缺少主机名")
	}
	return u, nil
}

// Load 从环境变量装载配置。
func Load() (*Config, error) {
	c := &Config{
		HTTPAddr:           env("OMP_HTTP_ADDR", "127.0.0.1:9999"),
		PanelURL:           strings.TrimRight(strings.TrimSpace(env("OMP_PANEL_URL", "")), "/"),
		PanelEntrance:      strings.TrimSpace(env("OMP_PANEL_ENTRANCE", "")),
		PanelSkipTLSVerify: envBool("OMP_PANEL_SKIP_TLS_VERIFY", false),
		SessionTTL:         envDuration("OMP_SESSION_TTL", 8*time.Hour),
		PendingTTL:         envDuration("OMP_PENDING_TTL", 5*time.Minute),
		CookieSecure:       envBool("OMP_COOKIE_SECURE", false),
		WebDir:             env("OMP_WEB_DIR", "../frontend/dist"),
	}
	if c.PanelURL == "" {
		return nil, errors.New("必须设置 OMP_PANEL_URL，例如 OMP_PANEL_URL=http://192.168.139.151:1314")
	}
	if _, err := c.PanelBaseURL(); err != nil {
		return nil, err
	}
	if c.SessionTTL <= 0 {
		return nil, errors.New("OMP_SESSION_TTL 必须为正数")
	}
	if c.PendingTTL <= 0 {
		return nil, errors.New("OMP_PENDING_TTL 必须为正数")
	}
	return c, nil
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(v) == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(strings.TrimSpace(v))
	if err != nil {
		return fallback
	}
	return parsed
}

func envDuration(key string, fallback time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(v) == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(strings.TrimSpace(v))
	if err != nil {
		return fallback
	}
	return parsed
}
