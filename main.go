package main

import (
	"io/fs"
	"log"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func root(c *gin.Context) {
	c.String(200, "Hello World")
}

func getFtpFolder(c *gin.Context) {
	folder := fs.FS(nil)
	path := filepath.Dir("./ftp_folder")

	c.String(200, path)
}

func main() {
	r := gin.Default()
	r.GET("/", root)
	r.GET("/ftp_folder", getFtpFolder)
	err := r.Run("0.0.0.0:8080")

	if err != nil {
		log.Fatal(err)
	}
}
