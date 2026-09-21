package constant

const (
	// SessionName 浏览器会话 Cookie 名称（HttpOnly）
	SessionName = "ompsession"
	// CSRFTokenName 浏览器可读的 CSRF Cookie 名称
	CSRFTokenName = "ompcsrftoken"
	// CSRFHeaderName 非安全方法必须携带的 CSRF 头
	CSRFHeaderName = "X-CSRF-Token"
	// SessionContextKey 会话在 gin 上下文中的键
	SessionContextKey = "OMP_SESSION"

	// APIPrefix 本服务对外接口前缀
	APIPrefix = "/api/v1"
	// PanelAPIPrefix 1Panel 接口前缀
	PanelAPIPrefix = "/api/v2"

	// StatusEnable 1Panel 通用启用状态
	StatusEnable = "Enable"
	// RoleAdmin 1Panel 管理员角色
	RoleAdmin = "ADMIN"
	// RoleNormal 1Panel 普通用户角色
	RoleNormal = "NORMAL"
)
