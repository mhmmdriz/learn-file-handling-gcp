package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/storage"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"google.golang.org/api/option"
)

const (
	projectID  = "alterra-academy-420809"
	bucketName = "learn_file_handling_go"
)

func uploadFilesToGCS(c echo.Context) error {
	// Parse form data, including files
	if err := c.Request().ParseMultipartForm(10 << 20); err != nil {
		return err
	}

	// Loop through uploaded files
	files := c.Request().MultipartForm.File["files"]
	for _, fileHeader := range files {
		// Open file
		file, err := fileHeader.Open()
		if err != nil {
			return err
		}
		defer file.Close()

		// Load GCP credentials securely (consider using KMS or secrets manager)
		ctx := context.Background()
		credentials := os.Getenv("GCP_CREDENTIALS")
		client, err := storage.NewClient(ctx, option.WithCredentialsJSON([]byte(credentials)))
		if err != nil {
			log.Printf("Error creating GCS client: %v", err)
			return err
		}
		defer client.Close()

		// Simpan file ke dalam bucket GCS di dalam folder /images
		dstPath := "images/" + fileHeader.Filename
		dst := client.Bucket(bucketName).Object(dstPath).NewWriter(ctx)
		defer dst.Close()

		// Salin isi file dari source ke destination di GCS
		if _, err = io.Copy(dst, file); err != nil {
			log.Printf("Error uploading file to GCS: %v", err)
			return err
		}

		// Upload berhasil
		fmt.Printf("File %s berhasil diunggah ke GCS\n", dstPath)
	}

	return c.String(http.StatusOK, "Semua file berhasil diunggah ke folder /images di GCS\n")
}

func main() {
	// Load environment variables from .env file
	// if err := godotenv.Load(); err != nil {
	// 	log.Fatalf("Error loading .env file: %v", err)
	// }

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Static("/", "public")
	e.POST("/gcp-upload", uploadFilesToGCS)

	e.Logger.Fatal(e.Start(":8080"))
}
