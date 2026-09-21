package web

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/ChishFoxcat/oh-my-1panel/backend/constant"
	"github.com/ChishFoxcat/oh-my-1panel/backend/global"
)

// Register 托管前端构建产物，并为单页应用提供 history 回退。
func Register(engine *gin.Engine) {
	dir := global.CONF.WebDir
	indexFile := filepath.Join(dir, "index.html")
	if _, err := os.Stat(indexFile); err != nil {
		global.LOGGER.Warn("未找到前端产物，仅提供接口与文档服务", "dir", dir, "hint", "在 frontend 目录执行 npm run build，或用 OMP_WEB_DIR 指定产物目录")
		engine.NoRoute(apiNotFound)
		return
	}
	global.LOGGER.Info("已托管前端产物", "dir", dir)

	assets := filepath.Join(dir, "assets")
	if info, err := os.Stat(assets); err == nil && info.IsDir() {
		engine.Use(cacheControl("/assets", "public, max-age=2628000, immutable"))
		engine.Static("/assets", assets)
	}

	engine.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if isReserved(path) {
			apiNotFound(c)
			return
		}
		if file, ok := lookup(dir, path); ok {
			c.File(file)
			return
		}
		c.Header("Cache-Control", "no-cache")
		c.File(indexFile)
	})
}

func lookup(dir, path string) (string, bool) {
	clean := filepath.Clean(strings.TrimPrefix(path, "/"))
	if clean == "." || strings.HasPrefix(clean, "..") || strings.Contains(clean, ".."+string(filepath.Separator)) {
		return "", false
	}
	target := filepath.Join(dir, clean)
	if info, err := os.Stat(target); err == nil && !info.IsDir() {
		return target, true
	}
	return "", false
}

func isReserved(path string) bool {
	return strings.HasPrefix(path, constant.APIPrefix+"/") || strings.HasPrefix(path, "/swagger")
}

func cacheControl(prefix, value string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, prefix) {
			c.Header("Cache-Control", value)
		}
		c.Next()
	}
}

func apiNotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "message": "接口不存在", "data": nil})
}
