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

var credentials = os.Getenv("GCP_CREDENTIALS")

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

func deleteFilesFromGCS(c echo.Context) error {
	// Parse request body to get file names to be deleted
	var fileNames []string
	if err := c.Bind(&fileNames); err != nil {
		return err
	}

	// Create GCS client
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsJSON([]byte(credentials)))
	if err != nil {
		return err
	}
	defer client.Close()

	// Delete files from GCS
	for _, fileName := range fileNames {
		// Construct object handle
		obj := client.Bucket(bucketName).Object(fileName)

		// Delete object
		if err := obj.Delete(ctx); err != nil {
			return err
		}

		fmt.Printf("File %s berhasil dihapus dari GCS\n", fileName)
	}

	return c.String(http.StatusOK, "Semua file berhasil dihapus dari GCS\n")
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
	e.DELETE("/gcp-delete", deleteFilesFromGCS)

	e.Logger.Fatal(e.Start(":8080"))
	// Test Update
}
