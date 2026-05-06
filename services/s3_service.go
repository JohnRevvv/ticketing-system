package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var s3Client *s3.Client
var bucketName string

// ✅ INIT S3 for access key
// func InitS3() error {
// 	cfg, err := config.LoadDefaultConfig(context.TODO(),
// 		config.WithRegion(os.Getenv("AWS_REGION")),
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	s3Client = s3.NewFromConfig(cfg)
// 	bucketName = os.Getenv("AWS_S3_BUCKET")

// 	return nil
// }

func InitS3() error {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(os.Getenv("AWS_REGION")),
	)
	if err != nil {
		return err
	}

	// SDK automatically uses IAM Role credentials (NO keys needed)
	s3Client = s3.NewFromConfig(cfg)
	bucketName = os.Getenv("AWS_S3_BUCKET")

	return nil
}

// ✅ UPLOAD
func UploadToS3(file *multipart.FileHeader, ticketID string) (string, string, error) {

	if s3Client == nil || bucketName == "" {
		return "", "", fmt.Errorf("S3 is not initialized")
	}

	src, err := file.Open()
	if err != nil {
		return "", "", err
	}
	defer src.Close()

	cleanFileName := sanitizeFileName(file.Filename)
	key := fmt.Sprintf("attachments/%s_%d_%s", ticketID, time.Now().UnixNano(), cleanFileName)

	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(key),
		Body:        src,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", "", err
	}

	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s",
		bucketName,
		os.Getenv("AWS_REGION"),
		key,
	)

	return cleanFileName, url, nil
}

// helper
func sanitizeFileName(name string) string {
	return name
}