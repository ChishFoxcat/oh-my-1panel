package panel

// LoginSetting 对应 1Panel dto.LoginSetting（GET /api/v2/core/auth/setting）。
type LoginSetting struct {
	IsDemo         bool   `json:"isDemo"`
	IsIntl         bool   `json:"isIntl"`
	IsOffline      bool   `json:"isOffline"`
	IsFxplay       bool   `json:"isFxplay"`
	IsEnterprise   bool   `json:"isEnterprise"`
	Language       string `json:"language"`
	MenuTabs       string `json:"menuTabs"`
	MenuAccordion  string `json:"menuAccordion"`
	PanelName      string `json:"panelName"`
	Theme          string `json:"theme"`
	NeedCaptcha    bool   `json:"needCaptcha"`
	PasskeySetting bool   `json:"passkeySetting"`
}

// Captcha 对应 1Panel dto.CaptchaResponse（GET /api/v2/core/auth/captcha）。
type Captcha struct {
	CaptchaID string `json:"captchaID"`
	ImagePath string `json:"imagePath"`
}

// LoginRequest 对应 1Panel dto.Login（POST /api/v2/core/auth/login）。
type LoginRequest struct {
	Name      string `json:"name"`
	Password  string `json:"password"`
	Captcha   string `json:"captcha,omitempty"`
	CaptchaID string `json:"captchaID,omitempty"`
	Language  string `json:"language"`
}

// MFARequest 对应 1Panel dto.MFALogin（POST /api/v2/core/auth/mfalogin）。
type MFARequest struct {
	SessionID string `json:"sessionId"`
	Code      string `json:"code"`
}

// UserLoginInfo 对应 1Panel dto.UserLoginInfo。
type UserLoginInfo struct {
	Name       string `json:"name"`
	Role       string `json:"role"`
	MfaStatus  string `json:"mfaStatus"`
	MfaSession string `json:"mfaSession"`
	Token      string `json:"token"`
}

// CurrentUserInfo 对应 1Panel dto.CurrentUserInfo（GET /api/v2/core/auth/current）。
type CurrentUserInfo struct {
	Name              string                `json:"name"`
	Role              string                `json:"role"`
	MfaStatus         string                `json:"mfaStatus"`
	MfaInterval       int                   `json:"mfaInterval"`
	AuthSource        string                `json:"authSource"`
	AuthSourceStatus  string                `json:"authSourceStatus"`
	ComplexitySetting string                `json:"complexitySetting"`
	ApiInterfaceStatus string               `json:"apiInterfaceStatus"`
	Permissions       []string              `json:"permissions"`
	NodeRoles         []CurrentUserNodeRole `json:"nodeRoles"`
}

// CurrentUserNodeRole 对应 1Panel dto.CurrentUserNodeRole。
type CurrentUserNodeRole struct {
	NodeID   int    `json:"nodeId"`
	NodeName string `json:"nodeName"`
	RoleID   int    `json:"roleId"`
	RoleName string `json:"roleName"`
}
