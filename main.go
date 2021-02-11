package main

import (
	"archive/zip"
	"context"
	"fmt"
	"github.com/cheggaaa/pb"
	"github.com/cloudfoundry-community/gautocloud"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/minio/minio-go/v7"
	"github.com/philips-software/gautocloud-connectors/hsdp"
	"path/filepath"
)

type UnpackRequest struct {
	SourceFile      string `json:"sourceFile"`
	DestinationPath string `json:"destinationPath"`
}

func main() {
	var svc *hsdp.S3MinioClient

	if err := gautocloud.Inject(&svc); err != nil {
		fmt.Printf("Error binding S3 Bucket: %v", err)
		return
	}
	e := echo.New()
	e.Use(middleware.Logger())
	e.POST("/unpack", unpackHandler(svc))

	_ = e.Start(":8080")
}

func unpackHandler(svc *hsdp.S3MinioClient) echo.HandlerFunc {
	return func(c echo.Context) error {
		unpackRequest := new(UnpackRequest)
		if err := c.Bind(unpackRequest); err != nil {
			return err
		}

		fmt.Printf("sourceFile: %s\n", unpackRequest.SourceFile)
		fmt.Printf("destinationPath: %s\n", unpackRequest.DestinationPath)

		object, err := svc.GetObject(context.Background(), svc.Bucket, unpackRequest.SourceFile, minio.GetObjectOptions{})
		if err != nil {
			return fmt.Errorf("GetObject: %v", err)
		}
		defer func() {
			_ = object.Close()
		}()

		stats, err := object.Stat()
		if err != nil {
			return fmt.Errorf("Stat: %v", err)
		}
		reader, err := zip.NewReader(object, stats.Size)
		if err != nil {
			return fmt.Errorf("zip.NewReader: %v", err)
		}
		for i := 0; i < len(reader.File); i++ {
			zipEntry := reader.File[i]
			destPath := filepath.Join(unpackRequest.DestinationPath, zipEntry.Name)
			fmt.Printf("Should open and write %s with size compressed size %d\n", destPath, zipEntry.CompressedSize64)
			src, err := zipEntry.Open()
			if err != nil {
				fmt.Printf("ERROR: Open: %v\n", err)
				continue
			}
			fmt.Printf("PutObject running...\n")
			progress := pb.New64(int64(zipEntry.UncompressedSize64))
			progress.Start()

			info, err := svc.PutObject(context.Background(), svc.Bucket, destPath, src, int64(zipEntry.UncompressedSize64), minio.PutObjectOptions{
				ContentType: "application/octet-stream",
				Progress:    progress,
			})
			if err != nil {
				fmt.Printf("ERROR: PutObject: %v\n", err)
				_ = src.Close()
				continue
			}
			fmt.Printf("INFO: %v\n", info)
			_ = src.Close()
		}
		return nil
	}
}
