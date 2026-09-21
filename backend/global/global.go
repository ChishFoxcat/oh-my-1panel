package global

import (
	"log/slog"

	"github.com/ChishFoxcat/oh-my-1panel/backend/config"
	"github.com/ChishFoxcat/oh-my-1panel/backend/init/session"
)

var (
	// CONF 服务配置
	CONF *config.Config
	// LOGGER 全局日志
	LOGGER *slog.Logger
	// SESSIONS 浏览器会话存储
	SESSIONS *session.Store
)
