package v1

import (
	"github.com/gin-gonic/gin"

	"github.com/ChishFoxcat/oh-my-1panel/backend/app/api/v1/helper"
	"github.com/ChishFoxcat/oh-my-1panel/backend/app/service"
)

// SystemInfo 面板公开信息
// @Tags System
// @Summary 面板信息与登录态
// @Description 登录页与首屏路由守卫使用：返回面板名称、主题、语言、是否需要验证码；携带有效会话 Cookie 时同时返回当前用户。
// @Produce json
// @Success 200 {object} dto.Response{data=dto.SystemInfo} "面板信息"
// @Failure 502 {object} dto.Response "面板不可达"
// @Router /system/info [get]
func SystemInfo(c *gin.Context) {
	result, err := service.NewAuthService().SystemInfo(c.Request.Context(), helper.CurrentSession(c))
	if err != nil {
		helper.Error(c, err)
		return
	}
	helper.SuccessWithData(c, result)
}

// SystemCurrent 面板当前用户信息
// @Tags System
// @Summary 当前登录用户
// @Description 透传面板 GET /core/auth/current，返回角色、权限与认证来源；面板会话失效时返回 401 并销毁本地会话。
// @Produce json
// @Success 200 {object} dto.Response{data=dto.CurrentUser} "当前用户"
// @Failure 401 {object} dto.Response "未登录或面板会话已过期"
// @Failure 502 {object} dto.Response "面板不可达"
// @Router /system/current [get]
func SystemCurrent(c *gin.Context) {
	result, err := service.NewAuthService().Current(c.Request.Context(), helper.CurrentSession(c))
	if err != nil {
		if helper.IsUnauthorized(err) {
			helper.ClearSessionCookies(c)
		}
		helper.Error(c, err)
		return
	}
	helper.SuccessWithData(c, result)
}
