package middleware

import (
	_ "embed"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/ChishFoxcat/oh-my-1panel/backend/constant"
	"github.com/ChishFoxcat/oh-my-1panel/backend/global"
)

// entranceUnavailable 是未携带安全入口凭证时返回的页面，与 1Panel 面板在同样情况下
// 返回的页面逐字节一致，用于隐藏本服务与面板的区别。
//
//go:embed html/entrance_unavailable.html
var entranceUnavailable []byte

// Entrance 实现安全入口门禁，语义与 1Panel 面板一致：
//
//   - 访问 /{入口} 及其子路径：视为入口自身，下发 base64(安全入口) 的 Cookie 后放行；
//   - 已登录会话：放行，因此登录后地址栏不需要再带安全入口；
//   - 持有效入口 Cookie：接口与静态资源放行；根路径下的页面请求 302 回 /{入口}，
//     这样刷新时会把入口自动补回链接；
//   - 其余请求：返回与面板一致的隐藏页。
func Entrance(entrance string) gin.HandlerFunc {
	if entrance == "" {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	prefix := "/" + entrance
	cookieValue := base64.StdEncoding.EncodeToString([]byte(entrance))
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		switch {
		case isEntrancePath(path, prefix):
			setEntranceCookie(c, cookieValue)
			c.Next()
		case hasValidSession(c):
			c.Next()
		case isEntranceCookieValid(c, cookieValue):
			if isDocumentRequest(c) && !isReservedPath(path) {
				c.Redirect(http.StatusFound, prefix)
				c.Abort()
				return
			}
			c.Next()
		default:
			c.Data(http.StatusOK, "text/html; charset=utf-8", entranceUnavailable)
			c.Abort()
		}
	}
}

func isEntrancePath(path, prefix string) bool {
	return path == prefix || strings.HasPrefix(path, prefix+"/")
}

func isEntranceCookieValid(c *gin.Context, want string) bool {
	value, err := c.Cookie(constant.EntranceCookieName)
	return err == nil && value == want
}

func hasValidSession(c *gin.Context) bool {
	id, err := c.Cookie(constant.SessionName)
	if err != nil {
		return false
	}
	_, ok := global.SESSIONS.Get(id)
	return ok
}

// isDocumentRequest 判断是否是浏览器地址栏/刷新产生的页面请求。
func isDocumentRequest(c *gin.Context) bool {
	return c.Request.Method == http.MethodGet &&
		strings.Contains(c.GetHeader("Accept"), "text/html") &&
		!strings.Contains(c.GetHeader("Sec-Fetch-Dest"), "script")
}

func isReservedPath(path string) bool {
	return strings.HasPrefix(path, constant.APIPrefix+"/") ||
		strings.HasPrefix(path, "/assets/") ||
		strings.HasPrefix(path, "/swagger")
}

func setEntranceCookie(c *gin.Context, value string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(constant.EntranceCookieName, value, 0, "/", "", global.CONF.CookieSecure, true)
}
