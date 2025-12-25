package http

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// StaticFileService 静态文件服务
type StaticFileService struct {
	staticDir string
	hasStatic bool
}

// NewStaticFileService 创建静态文件服务
func NewStaticFileService() (*StaticFileService, error) {
	// 尝试多个可能的静态文件目录位置
	possiblePaths := []string{
		"static",                    // 相对于可执行文件
		"./static",                  // 当前目录
		"../static",                 // 上级目录
		filepath.Join("..", "..", "static"), // 上上级目录（用于开发环境）
	}

	var staticDir string
	found := false

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			staticDir = path
			found = true
			break
		}
	}

	return &StaticFileService{
		staticDir: staticDir,
		hasStatic: found,
	}, nil
}

// RegisterRoutes 注册静态文件路由
func (s *StaticFileService) RegisterRoutes(engine *gin.Engine) {
	if !s.hasStatic {
		// 如果没有静态文件，返回简单提示
		engine.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Assistant Qisumi API Server",
				"status":  "running",
				"frontend": "not built",
			})
		})
		return
	}

	// 静态资源文件（直接返回）
	engine.Static("/assets", filepath.Join(s.staticDir, "assets"))

	// SPA 路由：所有非API路径都返回 index.html
	engine.NoRoute(func(c *gin.Context) {
		// 如果是API路径，返回404
		if c.Request.URL.Path == "/api/" || len(c.Request.URL.Path) > 5 && c.Request.URL.Path[:5] == "/api/" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "API endpoint not found",
			})
			return
		}

		// 其他路径返回 index.html（支持 SPA 路由）
		indexFile := filepath.Join(s.staticDir, "index.html")
		c.File(indexFile)
	})

	// 根路径也返回 index.html
	engine.GET("/", func(c *gin.Context) {
		c.File(filepath.Join(s.staticDir, "index.html"))
	})
}

// HasStaticFiles 检查是否有静态文件
func (s *StaticFileService) HasStaticFiles() bool {
	return s.hasStatic
}
