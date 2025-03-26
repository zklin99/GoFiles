package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
)

func root(c *gin.Context) {
	c.String(200, "Hello World")
}

type fileInfo struct {
	Name  string `json:"name"`
	size  int64
	IsDir bool `json:"is_dir"`
}

func getFtpFolder(c *gin.Context) {
	path := c.DefaultQuery("path", "/")
	path = "./ftp_folder" + path
	ftpFolderDir, err := os.ReadDir(path)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		}
		files = append(files, _fileInfo)
	}

	c.IndentedJSON(200, files)
	//path := filepath.Dir("./ftp_folder")
	return

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
	r.GET("/ftp", getFtpFolder)
	r.GET("/panic", func(c *gin.Context) {
		panic("触发Panic")
	})
	err = r.Run("0.0.0.0:13939")

	if err != nil {
		log.Fatal(err)
	}
}
