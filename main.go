package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
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
	ftpFolderDir, err := os.ReadDir("./ftp_folder")
	if err != nil {
		log.Println(err)
	}
	log.Printf(
		"ftp_folder: %v, 文件数量：%d",
		ftpFolderDir, len(ftpFolderDir),
	)

	files := make([]fileInfo, 0, len(ftpFolderDir))

	print(files)
	for _, file := range ftpFolderDir {
		_fileInfo := fileInfo{
			Name:  file.Name(),
			IsDir: file.IsDir(),
		}
		files = append(files, _fileInfo)
	}

	c.IndentedJSON(200, files)
	path := filepath.Dir("./ftp_folder")

	c.String(200, path)
}

func main() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("当前目录为：%v", dir)

	r := gin.Default()
	r.GET("/", root)
	r.GET("/ftp_folder", getFtpFolder)
	err = r.Run("0.0.0.0:13939")

	if err != nil {
		log.Fatal(err)
	}
}
