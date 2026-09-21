package dto

// SystemInfo 是登录页与首屏路由守卫所需的公开信息。
type SystemInfo struct {
	PanelName   string    `json:"panelName" example:"1Panel"`
	PanelURL    string    `json:"panelURL" example:"http://192.168.139.151:1314"`
	Language    string    `json:"language" example:"zh"`
	Theme       string    `json:"theme" example:"dark"`
	NeedCaptcha bool      `json:"needCaptcha" example:"false"`
	Logged      bool      `json:"logged" example:"false"`
	User        *UserInfo `json:"user,omitempty"`
	Version     string    `json:"version" example:"0.1.0"`
}
