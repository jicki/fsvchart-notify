package server

import (
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	statik "github.com/rakyll/statik/fs"

	_ "fsvchart-notify/statik" // Import statik package
)

// 全局版本变量
var Version = "unknown"

// 从VERSION文件读取版本号
func getVersion() string {
	// 尝试读取VERSION文件
	data, err := ioutil.ReadFile("VERSION")
	if err != nil {
		// 如果读取失败，返回默认版本
		return "unknown"
	}
	// 去除空白字符并返回
	return strings.TrimSpace(string(data))
}

// detectContentType 根据文件扩展名返回正确的 MIME 类型
func detectContentType(path string) string {
	switch {
	case strings.HasSuffix(path, ".js"):
		if strings.Contains(path, "module") {
			return "application/javascript; charset=utf-8"
		}
		return "application/javascript; charset=utf-8"
	case strings.HasSuffix(path, ".css"):
		return "text/css; charset=utf-8"
	case strings.HasSuffix(path, ".html"):
		return "text/html; charset=utf-8"
	case strings.HasSuffix(path, ".json"):
		return "application/json; charset=utf-8"
	case strings.HasSuffix(path, ".png"):
		return "image/png"
	case strings.HasSuffix(path, ".jpg"), strings.HasSuffix(path, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(path, ".gif"):
		return "image/gif"
	case strings.HasSuffix(path, ".svg"):
		return "image/svg+xml"
	case strings.HasSuffix(path, ".ico"):
		return "image/x-icon"
	case strings.HasSuffix(path, ".woff"):
		return "font/woff"
	case strings.HasSuffix(path, ".woff2"):
		return "font/woff2"
	case strings.HasSuffix(path, ".ttf"):
		return "font/ttf"
	case strings.HasSuffix(path, ".eot"):
		return "application/vnd.ms-fontobject"
	default:
		return "application/octet-stream"
	}
}

// NewServer 初始化并返回 *http.Server
func NewServer(addr string, port int) *http.Server {
	// 默认 gin.ReleaseMode
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 添加版本信息到响应头
	r.Use(func(c *gin.Context) {
		c.Header("X-FSVChart-Version", Version)
		c.Next()
	})

	// 注册 API 路由
	RegisterRoutes(r)

	// 嵌入的静态资源
	statikFS, err := statik.New()
	if err != nil {
		log.Fatalf("Failed to create statik FS: %v", err)
	}

	// 对于前端路由路径，返回 index.html
	frontendRoutes := []string{"/", "/login", "/profile", "/send-records"}
	for _, route := range frontendRoutes {
		r.GET(route, func(c *gin.Context) {
			f, err := statikFS.Open("/index.html")
			if err != nil {
				c.String(http.StatusNotFound, "404 Not Found")
				return
			}
			defer f.Close()

			info, err := f.Stat()
			if err != nil {
				c.String(http.StatusNotFound, "404 Not Found")
				return
			}

			c.DataFromReader(http.StatusOK, info.Size(), "text/html; charset=utf-8", f, nil)
		})
	}

	// 对于没有匹配的路由，尝试从 statik 文件系统中获取文件
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		f, err := statikFS.Open(path)
		if err != nil {
			c.String(http.StatusNotFound, "404 Not Found")
			return
		}
		defer f.Close()

		info, err := f.Stat()
		if err != nil {
			c.String(http.StatusNotFound, "404 Not Found")
			return
		}

		contentType := detectContentType(path)
		c.DataFromReader(http.StatusOK, info.Size(), contentType, f, nil)
	})

	// 返回标准库的 http.Server
	srv := &http.Server{
		Addr:    addr + ":" + strconv.Itoa(port),
		Handler: r,
	}
	return srv
}
