package main

import (
	"log"
	"net/http"
	"context"
	"os"
	"io/ioutil"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load("./env/env.go")

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	r := gin.Default()
	r.Use(cors.Default())

	r.Static("/assets", "./assets")
	r.LoadHTMLGlob("templates/*")
	r.MaxMultipartMemory = 900 * 1024 * 1024

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Printf("error: %v", err)
		return
	}

	client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(client)

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	r.POST("/", func(c *gin.Context) {
		file, err := c.FormFile("image")
		if err != nil {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"error": "Failed to upload file",
			})
			return
		}

		f, openErr := file.Open()

		if openErr != nil {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"error": "Failed to upload file",
			})
			return
		}

		_, uploadErr := uploader.Upload(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(os.Getenv("AWS_BUCKET_NAME")),
			Key:    aws.String(file.Filename),
			Body:   f,
			ACL:    "public-read",
		})

		if uploadErr != nil {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"error": "Failed to upload file",
			})
			return
		}

		c.HTML(http.StatusOK, "success.html", gin.H{})
	})

	r.GET("/success", func(c *gin.Context) {
		c.HTML(http.StatusOK, "success.html", gin.H{})
	})

	r.POST("/save-credentials", func(c *gin.Context) {
		var formData struct {
			Region     string `json:"region"`
			AccessKey  string `json:"accessKey"`
			SecretKey  string `json:"secretKey"`
			BucketName string `json:"bucketName"`
		}

		if err := c.BindJSON(&formData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON data"})
			return
		}

		content := `AWS_REGION=` + formData.Region + `
AWS_ACCESS_KEY_ID=` + formData.AccessKey + `
AWS_SECRET_ACCESS_KEY=` + formData.SecretKey + `
AWS_BUCKET_NAME=` + formData.BucketName + `
`

		err := ioutil.WriteFile("./env/env.go", []byte(content), 0644)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save credentials"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Credentials saved successfully"})
	})
	r.Run()
}
