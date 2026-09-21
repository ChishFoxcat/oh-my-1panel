package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/ChishFoxcat/oh-my-1panel/backend/app/api/v1/helper"
	"github.com/ChishFoxcat/oh-my-1panel/backend/constant"
	"github.com/ChishFoxcat/oh-my-1panel/backend/global"
)

// SessionLoader 尝试装载浏览器会话，未登录也继续执行（供公开接口读取登录态）。
func SessionLoader() gin.HandlerFunc {
	return func(c *gin.Context) {
		if id, err := c.Cookie(constant.SessionName); err == nil {
			if current, ok := global.SESSIONS.Get(id); ok {
				c.Set(constant.SessionContextKey, current)
			}
		}
		c.Next()
	}
}

// RequireSession 要求已登录，否则返回 401。
func RequireSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		if helper.CurrentSession(c) == nil {
			helper.ClearSessionCookies(c)
			helper.Unauthorized(c, "请先登录！")
			c.Abort()
			return
		}
		c.Next()
	}
}

// CSRFTokenGuard 校验非安全方法的 CSRF 令牌，规则与 1Panel 保持一致：
// 登录、动态验证码等无会话接口豁免，其余写操作必须携带 X-CSRF-Token。
func CSRFTokenGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !requiresCSRFTokenCheck(c) {
			c.Next()
			return
		}
		current := helper.CurrentSession(c)
		token := strings.TrimSpace(c.GetHeader(constant.CSRFHeaderName))
		if current == nil || token == "" ||
			subtle.ConstantTimeCompare([]byte(token), []byte(current.CSRFToken)) != 1 {
			helper.Forbidden(c, "CSRF 令牌校验失败，请刷新页面后重试！")
			c.Abort()
			return
		}
		c.Next()
	}
}

func requiresCSRFTokenCheck(c *gin.Context) bool {
	switch c.Request.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return false
	}
	path := c.Request.URL.Path
	if !strings.HasPrefix(path, constant.APIPrefix+"/") {
		return false
	}
	switch path {
	case constant.APIPrefix + "/auth/login",
		constant.APIPrefix + "/auth/mfa",
		constant.APIPrefix + "/auth/captcha":
		return false
	}
	_, err := c.Cookie(constant.SessionName)
	return err == nil
}
