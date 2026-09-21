package v1

import (
	"github.com/gin-gonic/gin"

	"github.com/ChishFoxcat/oh-my-1panel/backend/app/api/v1/helper"
	"github.com/ChishFoxcat/oh-my-1panel/backend/app/dto"
	"github.com/ChishFoxcat/oh-my-1panel/backend/app/service"
)

// Login 账号密码登录
// @Tags Auth
// @Summary 用户登录
// @Description 使用 1Panel 面板账号密码登录。密码在服务端按面板约定（RSA + AES 混合加密）加密后转发，浏览器不接触面板凭据。
// @Description 面板开启 MFA 时返回 mfaRequired=true 与 mfaSession，需继续调用 /auth/mfa 完成登录。
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "登录请求"
// @Success 200 {object} dto.Response{data=dto.LoginResponse} "登录结果"
// @Failure 400 {object} dto.Response "参数错误"
// @Failure 401 {object} dto.Response "账号密码错误、验证码错误或安全入口错误"
// @Failure 502 {object} dto.Response "面板不可达"
// @Router /auth/login [post]
func Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	result, current, err := service.NewAuthService().Login(c.Request.Context(), req)
	if err != nil {
		helper.Error(c, err)
		return
	}
	if current != nil {
		helper.SetSessionCookies(c, current)
	}
	helper.SuccessWithData(c, result)
}

// MFA 动态验证码二次校验
// @Tags Auth
// @Summary 动态验证码校验
// @Description 携带登录返回的 mfaSession 与 6 位动态验证码完成登录。验证码错误时中间会话保留，可直接重试。
// @Accept json
// @Produce json
// @Param request body dto.MFARequest true "动态验证码请求"
// @Success 200 {object} dto.Response{data=dto.LoginResponse} "登录结果"
// @Failure 400 {object} dto.Response "参数错误"
// @Failure 401 {object} dto.Response "验证码错误或中间会话已过期"
// @Failure 502 {object} dto.Response "面板不可达"
// @Router /auth/mfa [post]
func MFA(c *gin.Context) {
	var req dto.MFARequest
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	result, current, err := service.NewAuthService().MFA(c.Request.Context(), req)
	if err != nil {
		helper.Error(c, err)
		return
	}
	helper.SetSessionCookies(c, current)
	helper.SuccessWithData(c, result)
}

// Captcha 获取登录验证码
// @Tags Auth
// @Summary 获取登录验证码
// @Description 面板在登录失败次数过多或开启图形校验时要求验证码，返回值中的 imagePath 可直接用于 <img src>。
// @Produce json
// @Success 200 {object} dto.Response{data=dto.CaptchaResponse} "验证码"
// @Failure 502 {object} dto.Response "面板不可达"
// @Router /auth/captcha [get]
func Captcha(c *gin.Context) {
	result, err := service.NewAuthService().Captcha(c.Request.Context())
	if err != nil {
		helper.Error(c, err)
		return
	}
	helper.SuccessWithData(c, result)
}

// Logout 退出登录
// @Tags Auth
// @Summary 退出登录
// @Description 注销面板会话并清除本地会话 Cookie。需要携带 X-CSRF-Token。
// @Produce json
// @Success 200 {object} dto.Response "退出成功"
// @Failure 401 {object} dto.Response "未登录"
// @Failure 403 {object} dto.Response "CSRF 令牌校验失败"
// @Router /auth/logout [post]
func Logout(c *gin.Context) {
	current := helper.CurrentSession(c)
	service.NewAuthService().Logout(c.Request.Context(), current)
	helper.ClearSessionCookies(c)
	helper.Success(c)
}
