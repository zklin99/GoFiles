package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// 配置项，可以后续移到配置文件中
var (
	serverPort = "13939"
	ftpRootDir = "./ftp_folder"
)

func root(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{})
}

type fileInfo struct {
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	IsDir    bool      `json:"is_dir"`
	Path     string    `json:"path"`
	ModTime  time.Time `json:"mod_time"`
	MimeType string    `json:"mime_type"`
}

// 格式化文件大小为人类可读格式
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(size)/float64(div), "KMGTPE"[exp])
}

// 安全地拼接和验证路径
func safePath(requestPath string) (string, error) {
	// 确保路径总是以/开头
	if !strings.HasPrefix(requestPath, "/") {
		requestPath = "/" + requestPath
	}

	// 使用path.Join安全地拼接路径
	fullPath := filepath.Join(ftpRootDir, requestPath)

	// 确保路径没有跳出ftpRootDir
	relPath, err := filepath.Rel(ftpRootDir, fullPath)
	if err != nil {
		return "", fmt.Errorf("invalid path")
	}

	if strings.HasPrefix(relPath, "..") {
		return "", fmt.Errorf("access denied: path outside allowed directory")
	}

	return fullPath, nil
}

func getFtpFolder(c *gin.Context) {
	requestPath := c.DefaultQuery("path", "/")
	if !strings.HasSuffix(requestPath, "/") && requestPath != "/" {
		requestPath += "/"
	}

	fullPath, err := safePath(requestPath)
	if err != nil {
		c.HTML(http.StatusForbidden, "error.html", gin.H{
			"error":  "Access denied: " + err.Error(),
			"title":  "Error",
			"header": requestPath,
		})
		return
	}

	pathInfo, err := os.Stat(fullPath)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error":  "File not found: " + err.Error(),
			"title":  "Error",
			"header": requestPath,
		})
		return
	}

	if !pathInfo.IsDir() {
		fileName := pathInfo.Name()
		fileSize := pathInfo.Size()
		log.Printf("提供文件下载: %s, 文件大小: %d 字节", fullPath, fileSize)

		// 检测文件MIME类型
		mimeType := mime.TypeByExtension(filepath.Ext(fileName))
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}

		// URL编码文件名以支持中文
		encodedFileName := url.QueryEscape(fileName)

		// 设置头信息让浏览器下载文件
		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", fileName, encodedFileName))
		c.Header("Content-Type", mimeType)
		c.Header("Content-Length", fmt.Sprintf("%d", fileSize))
		// 添加缓存控制
		c.Header("Cache-Control", "max-age=31536000")
		c.File(fullPath)
		return
	}

	dirEntries, err := os.ReadDir(fullPath)
	if err != nil {
		log.Printf("读取目录错误: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error":  err.Error(),
			"title":  "文件列���",
			"header": requestPath,
		})
		return
	}
	log.Printf(
		"目录: %s, 文件数量：%d\n",
		requestPath, len(dirEntries),
	)

	files := make([]fileInfo, 0, len(dirEntries))

	// 如果不是根目录，添加返回上级目录的选项
	if requestPath != "/" {
		parentPath := path.Dir(strings.TrimSuffix(requestPath, "/")) + "/"
		if parentPath == "//" {
			parentPath = "/"
		}
		files = append(files, fileInfo{
			Name:  "..",
			IsDir: true,
			Path:  parentPath,
		})
	}

	for _, entry := range dirEntries {
		info, err := entry.Info()
		if err != nil {
			log.Printf("获取文件信息错误: %s - %v", entry.Name(), err)
			continue
		}

		mimeType := ""
		if !info.IsDir() {
			mimeType = mime.TypeByExtension(filepath.Ext(info.Name()))
			if mimeType == "" {
				mimeType = "application/octet-stream"
			}
		}

		_fileInfo := fileInfo{
			Name:     info.Name(),
			Size:     info.Size(),
			IsDir:    info.IsDir(),
			Path:     requestPath + info.Name(),
			ModTime:  info.ModTime(),
			MimeType: mimeType,
		}
		files = append(files, _fileInfo)
	}

	c.HTML(http.StatusOK, "files.html", gin.H{
		"title":          "文件列表",
		"header":         requestPath,
		"files":          files,
		"formatFileSize": formatFileSize,
	})
}

func main() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("当前目录为：%v", dir)

	// 确保ftp_folder目录存在
	if _, err := os.Stat(ftpRootDir); os.IsNotExist(err) {
		log.Printf("创建目录: %s", ftpRootDir)
		if err := os.MkdirAll(ftpRootDir, 0755); err != nil {
			log.Fatalf("无法创建目录 %s: %v", ftpRootDir, err)
		}
	}

	// 设置发布模式减少日志
	// gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 设置恢复中间件
	r.Use(gin.Recovery())

	// 路由配置
	r.GET("/", root)
	r.GET("/files", getFtpFolder)

	// 仅在开发环境中使用
	r.GET("/panic", func(c *gin.Context) {
		panic("触发Panic")
	})

	r.LoadHTMLGlob("templates/*")

	// 优雅关闭服务器
	srv := &http.Server{
		Addr:    "0.0.0.0:" + serverPort,
		Handler: r,
	}

	log.Printf("服务器启动在 http://0.0.0.0:%s", serverPort)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
