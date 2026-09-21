package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/ChishFoxcat/oh-my-1panel/backend/app/api/v1"
	"github.com/ChishFoxcat/oh-my-1panel/backend/cmd/server/web"
	"github.com/ChishFoxcat/oh-my-1panel/backend/constant"
	"github.com/ChishFoxcat/oh-my-1panel/backend/global"
	"github.com/ChishFoxcat/oh-my-1panel/backend/middleware"
)

// InitRouter 组装 HTTP 路由。
//
// 启用安全入口（OMOP_ENTRANCE）时，接口、Swagger 与前端产物统一挂在 /{入口} 之下，
// 入口以外的请求由 web.Register 注册的 NoRoute 拦掉。
func InitRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(middleware.Logger())
	engine.Use(middleware.SessionLoader())
	engine.Use(middleware.CSRFTokenGuard())

	prefix := global.CONF.WebPrefix()

	api := engine.Group(prefix + constant.APIPrefix)
	system := api.Group("/system")
	{
		system.GET("/info", v1.SystemInfo)
		system.GET("/current", middleware.RequireSession(), v1.SystemCurrent)
	}

	auth := api.Group("/auth")
	{
		auth.GET("/captcha", v1.Captcha)
		auth.POST("/login", v1.Login)
		auth.POST("/mfa", v1.MFA)
		auth.POST("/logout", middleware.RequireSession(), v1.Logout)
	}

	engine.GET(prefix+"/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	web.Register(engine, prefix)
	return engine
}
