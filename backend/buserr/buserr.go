package buserr

import "net/http"

// BusinessError 是可直接展示给用户的中文业务错误，同时携带建议的 HTTP 状态码。
// Detail 为可选的机器可读标识（面板错误键），供前端做分支处理。
type BusinessError struct {
	Status  int
	Message string
	Detail  string
}

func (e *BusinessError) Error() string {
	return e.Message
}

// New 构造 400 业务错误。
func New(message string) *BusinessError {
	return &BusinessError{Status: http.StatusBadRequest, Message: message}
}

// Unauthorized 构造 401 业务错误。
func Unauthorized(message string) *BusinessError {
	return &BusinessError{Status: http.StatusUnauthorized, Message: message}
}

// Forbidden 构造 403 业务错误。
func Forbidden(message string) *BusinessError {
	return &BusinessError{Status: http.StatusForbidden, Message: message}
}

// Internal 构造 500 业务错误。
func Internal(message string) *BusinessError {
	return &BusinessError{Status: http.StatusInternalServerError, Message: message}
}

// WithStatus 修改错误状态码。
func (e *BusinessError) WithStatus(status int) *BusinessError {
	e.Status = status
	return e
}

// WithDetail 附加机器可读标识。
func (e *BusinessError) WithDetail(detail string) *BusinessError {
	e.Detail = detail
	return e
}
