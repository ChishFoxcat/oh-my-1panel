package dto

// Response 是所有接口的统一响应信封，code 与 HTTP 状态码保持一致。
// Detail 仅在错误来自面板时携带面板错误键（如 ErrCaptchaCode），供前端做分支处理。
type Response struct {
	Code    int    `json:"code" example:"200"`
	Message string `json:"message" example:""`
	Detail  string `json:"detail,omitempty" example:"ErrCaptchaCode"`
	Data    any    `json:"data"`
}
