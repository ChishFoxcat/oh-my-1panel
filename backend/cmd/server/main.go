package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ChishFoxcat/oh-my-1panel/backend/config"
	_ "github.com/ChishFoxcat/oh-my-1panel/backend/docs"
	"github.com/ChishFoxcat/oh-my-1panel/backend/global"
	"github.com/ChishFoxcat/oh-my-1panel/backend/init/session"
	"github.com/ChishFoxcat/oh-my-1panel/backend/router"
)

// @title oh-my-1panel API
// @version 0.1.0
// @description 1Panel 新版 WebUI 的后端服务（BFF）。浏览器只与本服务通信，由服务端持有面板凭据、
// @description 加密登录密码、维护会话并按需聚合转发 1Panel 接口。
// @description 统一响应信封为 {"code": <HTTP 状态码>, "message": "", "data": {}}。
// @description 启用安全入口（OMOP_ENTRANCE）时，实际请求路径为 /{入口}/api/v1/...。
// @BasePath /api/v1
// @schemes http https
func main() {
	if err := run(); err != nil {
		slog.Error("服务启动失败", "error", err)
		os.Exit(1)
	}
}

func run() error {
	conf, err := config.Load()
	if err != nil {
		return err
	}
	global.CONF = conf
	global.LOGGER = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	global.SESSIONS = session.NewStore(conf.SessionTTL, conf.PendingTTL)
	defer global.SESSIONS.Close()

	global.LOGGER.Info("服务启动",
		"addr", conf.HTTPAddr,
		"panel", conf.PanelURL,
		"entrance", conf.PanelEntrance != "",
		"sessionTTL", conf.SessionTTL.String(),
	)

	server := &http.Server{
		Addr:              conf.HTTPAddr,
		Handler:           router.InitRouter(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       5 * time.Minute,
		WriteTimeout:      5 * time.Minute,
		IdleTimeout:       2 * time.Minute,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		global.LOGGER.Info("收到退出信号，正在关闭服务")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	}
}
