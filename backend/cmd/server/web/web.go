package web

import (
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/ChishFoxcat/oh-my-1panel/backend/constant"
	"github.com/ChishFoxcat/oh-my-1panel/backend/global"
)

// entranceUnavailable 是入口外请求返回的页面，与 1Panel 面板在未输入安全入口时
// 返回的页面逐字节一致（取自面板自身），用于隐藏本服务与面板的区别。
//
//go:embed html/entrance_unavailable.html
var entranceUnavailable []byte

// Register 托管前端构建产物。
//
// prefix 是安全入口路径前缀（如 /chish，未启用入口时为空串）：
//   - 入口路径内：按文件系统提供静态资源，未命中文件时回退到注入 <base> 的 index.html；
//   - 入口路径外：返回与面板一致的中性页面，不暴露本服务的存在与形态。
func Register(engine *gin.Engine, prefix string) {
	dir := global.CONF.WebDir
	// 中间件必须先于静态路由注册，否则不会进入该路由的处理链
	engine.Use(cacheControl("/assets/", "public, max-age=2628000, immutable"))
	if info, err := os.Stat(filepath.Join(dir, "assets")); err == nil && info.IsDir() {
		engine.Static(prefix+"/assets", filepath.Join(dir, "assets"))
	}

	index, err := buildIndex(dir, prefix)
	if err != nil {
		global.LOGGER.Warn("未找到前端产物，仅提供接口与文档服务",
			"dir", dir, "hint", "在 frontend 目录执行 npm run build，或用 OMOP_WEB_DIR 指定产物目录")
	}
	global.LOGGER.Info("前端产物托管", "dir", dir, "entrance", prefix, "mounted", err == nil)

	engine.NoRoute(func(c *gin.Context) {
		relative, inside := relativePath(c.Request.URL.Path, prefix)
		if !inside {
			c.Data(http.StatusOK, "text/html; charset=utf-8", entranceUnavailable)
			return
		}
		if isReservedPath(relative) {
			apiNotFound(c)
			return
		}
		if index != nil {
			if file, found := lookup(dir, relative); found {
				c.File(file)
				return
			}
			// 静态资源缺失时不要回退到 SPA 页面，避免 404 变成 200
			if isStaticAsset(relative) {
				blocked(c)
				return
			}
			c.Header("Cache-Control", "no-cache")
			c.Data(http.StatusOK, "text/html; charset=utf-8", index)
			return
		}
		blocked(c)
	})
}

// buildIndex 读取 index.html 并注入 <base href>，使前端产物的相对资源路径
// 与前端路由基址跟当前安全入口一致（入口可变而无需重新构建）。
func buildIndex(dir, prefix string) ([]byte, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "index.html"))
	if err != nil {
		return nil, fmt.Errorf("读取 index.html 失败：%w", err)
	}
	content := string(raw)
	const head = "<head>"
	if at := strings.Index(strings.ToLower(content), head); at >= 0 {
		insertAt := at + len(head)
		tag := fmt.Sprintf("\n    <base href=\"%s/\">", prefix)
		content = content[:insertAt] + tag + content[insertAt:]
	}
	return []byte(content), nil
}

// relativePath 判断请求路径是否位于入口内，并返回去掉入口前缀的相对路径。
func relativePath(path, prefix string) (string, bool) {
	if prefix == "" {
		return path, true
	}
	if path == prefix {
		return "/", true
	}
	if strings.HasPrefix(path, prefix+"/") {
		return strings.TrimPrefix(path, prefix), true
	}
	return "", false
}

func isReservedPath(relative string) bool {
	return relative == constant.APIPrefix ||
		strings.HasPrefix(relative, constant.APIPrefix+"/") ||
		strings.HasPrefix(relative, "/swagger")
}

// staticExtensions 参与静态资源判定的扩展名，用于把"文件缺失"与"前端路由"区分开。
var staticExtensions = map[string]bool{
	".js": true, ".mjs": true, ".css": true, ".map": true, ".json": true, ".txt": true,
	".ico": true, ".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".svg": true,
	".webp": true, ".woff": true, ".woff2": true, ".ttf": true, ".wasm": true,
}

func isStaticAsset(relative string) bool {
	return strings.HasPrefix(relative, "/assets/") ||
		staticExtensions[strings.ToLower(filepath.Ext(relative))]
}

// lookup 在产物目录内定位真实文件，拒绝越出目录的路径。
func lookup(dir, relative string) (string, bool) {
	clean := filepath.Clean(strings.TrimPrefix(relative, "/"))
	if clean == "." || strings.HasPrefix(clean, "..") || strings.Contains(clean, ".."+string(filepath.Separator)) {
		return "", false
	}
	target := filepath.Join(dir, clean)
	if info, err := os.Stat(target); err == nil && !info.IsDir() {
		return target, true
	}
	return "", false
}

func cacheControl(fragment, value string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.Request.URL.Path, fragment) {
			c.Header("Cache-Control", value)
		}
		c.Next()
	}
}

// blocked 对入口外的请求返回中性 404，不泄露任何服务特征。
func blocked(c *gin.Context) {
	http.NotFound(c.Writer, c.Request)
}

func apiNotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "message": "接口不存在", "data": nil})
}
