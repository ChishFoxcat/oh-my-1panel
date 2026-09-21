package panel

import (
	"errors"
	"net/http"
)

// 1Panel 侧的错误键，用于把面板返回的错误码翻译为中文提示。
const (
	KeyAuth                = "ErrAuth"
	KeyEntrance            = "ErrEntrance"
	KeyCaptchaCode         = "ErrCaptchaCode"
	KeyLoginLocked         = "ErrLoginLocked"
	KeyNotLogin            = "ErrNotLogin"
	KeySessionDataNotFound = "ErrSessionDataNotFound"
	KeyPasswordExpired     = "ErrPasswordExpired"
	KeyRecordNotFound      = "ErrRecordNotFound"
	KeyMFAFailed           = "ErrMFAFailed"
)

var messages = map[string]string{
	KeyAuth:                "您输入的用户名或密码不正确，请重新输入！",
	KeyEntrance:            "安全入口信息错误，请检查后重试！",
	KeyCaptchaCode:         "验证码错误！",
	KeyLoginLocked:         "失败次数过多，当前登录已被临时锁定 5 分钟。",
	KeyNotLogin:            "当前会话已过期，请重新登录！",
	KeySessionDataNotFound: "当前会话已过期，请重新登录！",
	KeyPasswordExpired:     "当前密码已过期，请先重置密码！",
	KeyRecordNotFound:      "面板数据缺失，请检查面板初始化状态！",
	KeyMFAFailed:           "动态验证码错误，请重新输入！",
}

// APIError 是 1Panel 业务信封中 code != 200 的统一错误。
type APIError struct {
	Code   int
	Key    string
	Detail string
}

func (e *APIError) Error() string {
	if msg, ok := messages[e.Key]; ok {
		return msg
	}
	if e.Key != "" {
		return e.Key
	}
	return http.StatusText(e.Code)
}

// IsSessionExpired 判断错误是否为面板会话失效。
func IsSessionExpired(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.Code == http.StatusUnauthorized && apiErr.Key != KeyAuth && apiErr.Key != KeyEntrance
}

// Message 返回错误键对应的中文提示，未收录时返回原键。
func Message(key string) string {
	if msg, ok := messages[key]; ok {
		return msg
	}
	return key
}
