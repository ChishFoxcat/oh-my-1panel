package web

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/ChishFoxcat/oh-my-1panel/backend/constant"
	"github.com/ChishFoxcat/oh-my-1panel/backend/global"
)

// Register 托管前端构建产物。
//
// 资源与接口都在根路径提供（前端产物按根路径构建），安全入口只作为进入凭证：
// 门禁由 middleware.Entrance 负责，这里只决定"入口内/根路径下要返回什么"。
func Register(engine *gin.Engine, entrance string) {
	dir := global.CONF.WebDir
	// 中间件必须先于静态路由注册，否则不会进入该路由的处理链
	engine.Use(cacheControl("/assets/", "public, max-age=2628000, immutable"))
	if info, err := os.Stat(filepath.Join(dir, "assets")); err == nil && info.IsDir() {
		engine.Static("/assets", filepath.Join(dir, "assets"))
	}

	indexPath := filepath.Join(dir, "index.html")
	_, indexErr := os.Stat(indexPath)
	if indexErr != nil {
		global.LOGGER.Warn("未找到前端产物，仅提供接口与文档服务",
			"dir", dir, "hint", "在 frontend 目录执行 npm run build，或用 OMOP_WEB_DIR 指定产物目录")
	} else {
		global.LOGGER.Info("前端产物托管", "dir", dir, "entrance", entrance)
	}

	prefix := ""
	if entrance != "" {
		prefix = "/" + entrance
	}

	engine.NoRoute(func(c *gin.Context) {
		relative, _ := stripEntrance(c.Request.URL.Path, prefix)
		if isReservedPath(relative) {
			apiNotFound(c)
			return
		}
		if indexErr != nil {
			notFound(c)
			return
		}
		if file, found := lookup(dir, relative); found {
			c.File(file)
			return
		}
		// 静态资源缺失时不要回退到 SPA 页面，避免 404 变成 200
		if isStaticAsset(relative) {
			notFound(c)
			return
		}
		// index.html 每次读取：构建产物更新后无需重启服务
		page, err := renderIndex(indexPath, prefix)
		if err != nil {
			global.LOGGER.Error("渲染 index.html 失败", "error", err)
			notFound(c)
			return
		}
		c.Header("Cache-Control", "no-cache")
		c.Data(http.StatusOK, "text/html; charset=utf-8", page)
	})
}

// stripEntrance 去掉安全入口路径前缀，返回应用自身路径；第二个返回值表示请求是否位于入口内。
func stripEntrance(path, prefix string) (string, bool) {
	if prefix == "" {
		return path, true
	}
	if path == prefix {
		return "/", true
	}
	if strings.HasPrefix(path, prefix+"/") {
		return strings.TrimPrefix(path, prefix), true
	}
	return path, false
}

// renderIndex 读取 index.html 并注入运行期配置：
//   - <base href="/"> 让相对资源路径始终相对站点根解析（安全入口路径下也成立）；
//   - <meta name="omop-entrance"> 告知前端安全入口挂载点，供路由与"登录后去掉入口"使用。
//
// 入口可变，因此这里每次请求都重新渲染，不缓存。
func renderIndex(path, prefix string) ([]byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取 index.html 失败：%w", err)
	}
	content := string(raw)
	const head = "<head>"
	if at := strings.Index(strings.ToLower(content), head); at >= 0 {
		insertAt := at + len(head)
		tags := fmt.Sprintf("\n    <base href=\"/\">\n    <meta name=\"omop-entrance\" content=\"%s\">", prefix)
		content = content[:insertAt] + tags + content[insertAt:]
	}
	return []byte(content), nil
}

func isReservedPath(path string) bool {
	return path == constant.APIPrefix ||
		strings.HasPrefix(path, constant.APIPrefix+"/") ||
		strings.HasPrefix(path, "/swagger")
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

func notFound(c *gin.Context) {
	http.NotFound(c.Writer, c.Request)
}

func apiNotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "message": "接口不存在", "data": nil})
}
