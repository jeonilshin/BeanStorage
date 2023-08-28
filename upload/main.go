package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Static("/assets", "./assets")
	r.LoadHTMLGlob("templates/*")
	r.MaxMultipartMemory = 60 << 20

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	r.POST("/", func(c *gin.Context) {
		file, err := c.FormFile("image")
		if err != nil {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"error": "Failed to upload image",
			})
			return
		}
		err = c.SaveUploadedFile(file, "./assets/uploads/"+file.Filename)
		if err != nil {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"error": "Failed to upload image",
			})
			return
		}
		c.HTML(http.StatusOK, "index.html", gin.H{
			"image": "./assets/uploads/" + file.Filename,
		})
	})
	r.Run(":8080")
}