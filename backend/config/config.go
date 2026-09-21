package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Config 服务配置，全部通过环境变量注入。
type Config struct {
	// HTTPAddr 本服务监听地址
	HTTPAddr string
	// Entrance 本服务的安全入口（单段路径，如 chish），空表示不启用
	Entrance string
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

// entrancePattern 安全入口取值：2-64 位字母、数字、连字符或下划线。
var entrancePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{2,64}$`)

// reservedEntrance 与本服务自身路由冲突或易混淆的入口名。
var reservedEntrance = map[string]bool{
	"api":         true,
	"swagger":     true,
	"assets":      true,
	"favicon.ico": true,
}

// WebPrefix 返回安全入口路径前缀，未启用入口时为空串。
func (c *Config) WebPrefix() string {
	if c.Entrance == "" {
		return ""
	}
	return "/" + c.Entrance
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
		HTTPAddr:           env("OMOP_HTTP_ADDR", "127.0.0.1:9999"),
		Entrance:           env("OMOP_ENTRANCE", ""),
		PanelURL:           strings.TrimRight(strings.TrimSpace(env("OMOP_PANEL_URL", "")), "/"),
		PanelEntrance:      strings.TrimSpace(env("OMOP_PANEL_ENTRANCE", "")),
		PanelSkipTLSVerify: envBool("OMOP_PANEL_SKIP_TLS_VERIFY", false),
		SessionTTL:         envDuration("OMOP_SESSION_TTL", 8*time.Hour),
		PendingTTL:         envDuration("OMOP_PENDING_TTL", 5*time.Minute),
		CookieSecure:       envBool("OMOP_COOKIE_SECURE", false),
		WebDir:             env("OMOP_WEB_DIR", "../frontend/dist"),
	}
	if c.PanelURL == "" {
		return nil, errors.New("必须设置 OMOP_PANEL_URL，例如 OMOP_PANEL_URL=http://192.168.139.151:1314")
	}
	if _, err := c.PanelBaseURL(); err != nil {
		return nil, err
	}
	if err := validateEntrance(c.Entrance); err != nil {
		return nil, err
	}
	if c.SessionTTL <= 0 {
		return nil, errors.New("OMOP_SESSION_TTL 必须为正数")
	}
	if c.PendingTTL <= 0 {
		return nil, errors.New("OMOP_PENDING_TTL 必须为正数")
	}
	return c, nil
}

func validateEntrance(entrance string) error {
	if entrance == "" {
		return nil
	}
	if !entrancePattern.MatchString(entrance) {
		return errors.New("OMOP_ENTRANCE 只允许 2-64 位的字母、数字、连字符或下划线")
	}
	if reservedEntrance[strings.ToLower(entrance)] {
		return fmt.Errorf("OMOP_ENTRANCE 不能使用保留路径 %q", entrance)
	}
	return nil
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
