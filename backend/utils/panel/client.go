package panel

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/ChishFoxcat/oh-my-1panel/backend/constant"
	"github.com/ChishFoxcat/oh-my-1panel/backend/utils/encrypt"
)

const (
	publicKeyCookie  = "panel_public_key"
	maxResponseBytes = 16 << 20
	requestTimeout   = 30 * time.Second
)

// Client 是 1Panel 面板 API 的会话化客户端。
//
// 每个客户端持有独立的 Cookie 罐，服务端保存的 psession / pcsrftoken / SecurityEntrance
// 都由该罐维护；调用方只需把某个客户端与一个浏览器会话绑定即可。
type Client struct {
	baseURL      *url.URL
	entranceCode string
	http         *http.Client
	jar          http.CookieJar

	mu        sync.Mutex
	publicKey *rsa.PublicKey
}

// New 构造面板客户端。entrance 为面板安全入口明文，可为空。
func New(baseURL, entrance string, skipTLSVerify bool) (*Client, error) {
	u, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return nil, fmt.Errorf("面板地址解析失败：%w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, errors.New("面板地址必须以 http:// 或 https:// 开头")
	}
	if u.Host == "" {
		return nil, errors.New("面板地址缺少主机名")
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("初始化 Cookie 罐失败：%w", err)
	}
	client := &Client{
		baseURL: u,
		http: &http.Client{
			Jar:     jar,
			Timeout: requestTimeout,
			Transport: &http.Transport{
				Proxy:             http.ProxyFromEnvironment,
				ForceAttemptHTTP2: true,
				TLSClientConfig:   &tls.Config{InsecureSkipVerify: skipTLSVerify},
			},
		},
		jar: jar,
	}
	if entrance != "" {
		client.entranceCode = base64.StdEncoding.EncodeToString([]byte(entrance))
	}
	return client, nil
}

// Close 关闭底层连接，会话销毁时调用。
func (c *Client) Close() {
	c.http.CloseIdleConnections()
}

// LoginSetting 读取面板登录设置（同时触发面板下发 panel_public_key）。
func (c *Client) LoginSetting(ctx context.Context) (*LoginSetting, error) {
	var setting LoginSetting
	if err := c.do(ctx, http.MethodGet, "/core/auth/setting", nil, &setting); err != nil {
		return nil, err
	}
	return &setting, nil
}

// Captcha 获取登录验证码。
func (c *Client) Captcha(ctx context.Context) (*Captcha, error) {
	var captcha Captcha
	if err := c.do(ctx, http.MethodGet, "/core/auth/captcha", nil, &captcha); err != nil {
		return nil, err
	}
	return &captcha, nil
}

// Login 使用账号密码登录面板，密码在发送前按面板约定加密。
func (c *Client) Login(ctx context.Context, req LoginRequest) (*UserLoginInfo, error) {
	if req.Password == "" {
		return nil, errors.New("登录密码不能为空")
	}
	publicKey, err := c.ensurePublicKey(ctx)
	if err != nil {
		return nil, err
	}
	encrypted, err := encrypt.EncryptPassword(req.Password, publicKey)
	if err != nil {
		return nil, err
	}
	payload := req
	payload.Password = encrypted

	var info UserLoginInfo
	if err := c.do(ctx, http.MethodPost, "/core/auth/login", payload, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// MFALogin 完成动态验证码二次登录。
func (c *Client) MFALogin(ctx context.Context, sessionID, code string) (*UserLoginInfo, error) {
	var info UserLoginInfo
	payload := MFARequest{SessionID: sessionID, Code: code}
	if err := c.do(ctx, http.MethodPost, "/core/auth/mfalogin", payload, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// Current 读取面板当前登录用户信息。
func (c *Client) Current(ctx context.Context) (*CurrentUserInfo, error) {
	var info CurrentUserInfo
	if err := c.do(ctx, http.MethodGet, "/core/auth/current", nil, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// Logout 注销面板会话。
func (c *Client) Logout(ctx context.Context) error {
	return c.do(ctx, http.MethodPost, "/core/auth/logout", nil, nil)
}

type envelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func (c *Client) do(ctx context.Context, method, path string, payload any, out any) error {
	target := *c.baseURL
	target.Path = strings.TrimSuffix(c.baseURL.Path, "/") + constant.PanelAPIPrefix + path

	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("请求体序列化失败：%w", err)
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, target.String(), body)
	if err != nil {
		return fmt.Errorf("构造面板请求失败：%w", err)
	}
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.entranceCode != "" {
		req.Header.Set("EntranceCode", c.entranceCode)
	}
	if requiresCSRF(method) {
		if token := c.Cookie(constant.CSRFTokenName); token != "" {
			req.Header.Set(constant.CSRFHeaderName, token)
		}
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("请求面板失败：%w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return fmt.Errorf("读取面板响应失败：%w", err)
	}
	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return fmt.Errorf("面板响应解析失败（HTTP %d）：%s", resp.StatusCode, summary(raw))
	}
	if env.Code != http.StatusOK {
		return &APIError{Code: env.Code, Key: env.Message, Detail: string(env.Data)}
	}
	if out == nil || len(env.Data) == 0 {
		return nil
	}
	if err := json.Unmarshal(env.Data, out); err != nil {
		return fmt.Errorf("面板响应数据解析失败：%w", err)
	}
	return nil
}

// Cookie 返回面板下发的 Cookie 值，不存在时返回空串。
func (c *Client) Cookie(name string) string {
	for _, item := range c.jar.Cookies(c.baseURL) {
		if item.Name == name {
			return item.Value
		}
	}
	return ""
}

func (c *Client) ensurePublicKey(ctx context.Context) (*rsa.PublicKey, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.publicKey != nil {
		return c.publicKey, nil
	}
	// 任意一次请求都会让面板下发 panel_public_key
	if _, err := c.LoginSetting(ctx); err != nil {
		return nil, err
	}
	raw := c.Cookie(publicKeyCookie)
	if raw == "" {
		return nil, errors.New("面板未下发 panel_public_key，无法加密登录密码")
	}
	pemData, err := base64.StdEncoding.DecodeString(decodeCookieValue(raw))
	if err != nil {
		return nil, fmt.Errorf("面板公钥解码失败：%w", err)
	}
	publicKey, err := encrypt.ParseRSAPublicKey(string(pemData))
	if err != nil {
		return nil, err
	}
	c.publicKey = publicKey
	return publicKey, nil
}

// decodeCookieValue 复刻 1Panel 前端 utils/auth.ts 的 urlDecode：
// 面板对 Cookie 值做了查询串编码（+ 表示空格、特殊字符百分号转义）。
func decodeCookieValue(value string) string {
	decoded, err := url.QueryUnescape(strings.ReplaceAll(value, "+", " "))
	if err != nil {
		return value
	}
	return decoded
}

func requiresCSRF(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return false
	default:
		return true
	}
}

func summary(raw []byte) string {
	text := strings.TrimSpace(string(raw))
	if len(text) > 200 {
		return text[:200] + "…"
	}
	return text
}
