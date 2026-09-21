package helper

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ChishFoxcat/oh-my-1panel/backend/app/dto"
	"github.com/ChishFoxcat/oh-my-1panel/backend/buserr"
	"github.com/ChishFoxcat/oh-my-1panel/backend/constant"
	"github.com/ChishFoxcat/oh-my-1panel/backend/global"
	"github.com/ChishFoxcat/oh-my-1panel/backend/init/session"
)

// Success 返回空数据成功响应。
func Success(c *gin.Context) {
	respond(c, http.StatusOK, "", "", nil)
}

// SuccessWithData 返回带数据的成功响应。
func SuccessWithData(c *gin.Context, data any) {
	respond(c, http.StatusOK, "", "", data)
}

// Error 把业务错误映射为响应，未识别错误按 500 处理。
func Error(c *gin.Context, err error) {
	var bizErr *buserr.BusinessError
	if errors.As(err, &bizErr) {
		respond(c, bizErr.Status, bizErr.Message, bizErr.Detail, nil)
		return
	}
	global.LOGGER.Error("接口处理失败", "path", c.Request.URL.Path, "error", err)
	respond(c, http.StatusInternalServerError, "服务器内部错误："+err.Error(), "", nil)
}

// BadRequest 返回 400。
func BadRequest(c *gin.Context, message string) {
	respond(c, http.StatusBadRequest, message, "", nil)
}

// Unauthorized 返回 401。
func Unauthorized(c *gin.Context, message string) {
	respond(c, http.StatusUnauthorized, message, "", nil)
}

// Forbidden 返回 403。
func Forbidden(c *gin.Context, message string) {
	respond(c, http.StatusForbidden, message, "", nil)
}

// CheckBindAndValidate 绑定并校验请求体，失败时已写入 400 响应。
func CheckBindAndValidate(req any, c *gin.Context) error {
	if err := c.ShouldBindJSON(req); err != nil {
		message := "请求参数错误：" + err.Error()
		respond(c, http.StatusBadRequest, message, "", nil)
		return errors.New(message)
	}
	return nil
}

// IsUnauthorized 判断错误是否为 401 业务错误。
func IsUnauthorized(err error) bool {
	var bizErr *buserr.BusinessError
	return errors.As(err, &bizErr) && bizErr.Status == http.StatusUnauthorized
}

// CurrentSession 返回当前会话，未登录时为 nil。
func CurrentSession(c *gin.Context) *session.Session {
	value, ok := c.Get(constant.SessionContextKey)
	if !ok {
		return nil
	}
	current, ok := value.(*session.Session)
	if !ok {
		return nil
	}
	return current
}

// SetSessionCookies 下发会话 Cookie 与 CSRF Cookie。
func SetSessionCookies(c *gin.Context, current *session.Session) {
	maxAge := int(time.Until(current.ExpiresAt).Seconds())
	secure := global.CONF.CookieSecure
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(constant.SessionName, current.ID, maxAge, "/", "", secure, true)
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(constant.CSRFTokenName, current.CSRFToken, maxAge, "/", "", secure, false)
}

// ClearSessionCookies 清除会话 Cookie 与 CSRF Cookie。
func ClearSessionCookies(c *gin.Context) {
	secure := global.CONF.CookieSecure
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(constant.SessionName, "", -1, "/", "", secure, true)
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(constant.CSRFTokenName, "", -1, "/", "", secure, false)
}

func respond(c *gin.Context, status int, message, detail string, data any) {
	c.JSON(status, dto.Response{Code: status, Message: message, Detail: detail, Data: data})
}
