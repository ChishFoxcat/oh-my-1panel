package dto

// UserInfo 是登录用户的最小信息。
type UserInfo struct {
	Name    string `json:"name" example:"admin"`
	Role    string `json:"role" example:"ADMIN"`
	IsAdmin bool   `json:"isAdmin" example:"true"`
}

// LoginRequest 账号密码登录请求。
type LoginRequest struct {
	Name      string `json:"name" binding:"required" example:"admin"`
	Password  string `json:"password" binding:"required" example:"1panel123456"`
	Captcha   string `json:"captcha" example:"42"`
	CaptchaID string `json:"captchaID" example:"Wk3kQz1s"`
}

// LoginResponse 登录结果：开启动态验证码时返回 mfaRequired 与 mfaSession。
type LoginResponse struct {
	MfaRequired bool      `json:"mfaRequired" example:"false"`
	MfaSession  string    `json:"mfaSession,omitempty" example:"bE1lSP2rq3Zg"`
	ExpiresAt   string    `json:"expiresAt,omitempty" example:"2026-09-21T18:30:00+08:00"`
	User        *UserInfo `json:"user,omitempty"`
}

// MFARequest 动态验证码二次校验请求。
type MFARequest struct {
	SessionID string `json:"sessionId" binding:"required" example:"bE1lSP2rq3Zg"`
	Code      string `json:"code" binding:"required" example:"123456"`
}

// CaptchaResponse 登录验证码。
type CaptchaResponse struct {
	CaptchaID string `json:"captchaID" example:"Wk3kQz1s"`
	ImagePath string `json:"imagePath" example:"data:image/png;base64,iVBORw0KGgo="`
}

// CurrentUser 面板当前登录用户信息。
type CurrentUser struct {
	Name        string   `json:"name" example:"admin"`
	Role        string   `json:"role" example:"ADMIN"`
	IsAdmin     bool     `json:"isAdmin" example:"true"`
	MfaStatus   string   `json:"mfaStatus" example:"Disable"`
	AuthSource  string   `json:"authSource" example:"local"`
	Permissions []string `json:"permissions"`
}
