package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"net/url"
	"os"
)

func root(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{})
}

type fileInfo struct {
	Name  string `json:"name"`
	size  int64
	IsDir bool `json:"is_dir"`
	Path  string
}

func getFtpFolder(c *gin.Context) {
	path1 := c.DefaultQuery("path", "/")
	path := "./ftp_folder" + path1

	pathInfo, err := os.Stat(path)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error":  err.Error(),
			"title":  "文件列表",
			"header": path1,
		})
		return
	}
	if !pathInfo.IsDir() {
		fileName := pathInfo.Name()
		fileSize := pathInfo.Size()
		log.Printf("提供文件下载: %s, 文件大小: %d 字节", path, fileSize)
		// URL编码文件名以支持中文
		encodedFileName := url.QueryEscape(fileName)

		// 设置头信息让浏览器下载文件
		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", fileName, encodedFileName))
		c.Header("Content-Type", "application/octet-stream")
		c.File(path)
		return
	}

	ftpFolderDir, err := os.ReadDir(path)
	if err != nil {
		log.Printf("读取目录错误: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error":  err.Error(),
			"title":  "文件列表",
			"header": path1,
		})
		return
	}
	log.Printf(
		"ftp_folder: %v, 文件数量：%d\n",
		ftpFolderDir, len(ftpFolderDir),
	)

	files := make([]fileInfo, 0, len(ftpFolderDir))

	for _, file := range ftpFolderDir {
		_fileInfo := fileInfo{
			Name:  file.Name(),
			IsDir: file.IsDir(),
			Path:  path1 + file.Name(),
		}
		files = append(files, _fileInfo)
	}

	c.HTML(http.StatusOK, "files.html", gin.H{
		"title":  "文件列表",
		"header": path1,
		"files":  files,
	})

}

func main() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("当前目录为：%v", dir)

	r := gin.Default()

	//r.Use(gin.Recovery())

	r.GET("/", root)
	r.GET("/files", getFtpFolder)
	r.GET("/panic", func(c *gin.Context) {
		panic("触发Panic")
	})
	r.LoadHTMLGlob("templates/*")
	err = r.Run("0.0.0.0:13939")

	if err != nil {
		log.Fatal(err)
	}
}
