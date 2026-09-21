package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ChishFoxcat/oh-my-1panel/backend/global"
)

// Logger 以结构化日志记录请求行、状态码与耗时。
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		global.LOGGER.Info("接口请求",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"cost", time.Since(start).String(),
			"ip", c.ClientIP(),
		)
	}
}
