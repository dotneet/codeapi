package storage

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"os"
	"path/filepath"
	"time"
)

type ImageBucket struct {
	Endpoint   string
	BucketName string
	Secret     string
	AccessKey  string
}

func (bucket *ImageBucket) PutObject(runId string, path string) (string, error) {
	endpoint := bucket.Endpoint
	accessKey := bucket.AccessKey
	secretKey := bucket.Secret
	useSSL := false

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return "", err
	}

	// Upload the image file to minio
	bucketName := bucket.BucketName
	objectName := runId + "/" + filepath.Base(path)
	contentType := "image/png"

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	info, err := minioClient.PutObject(context.Background(), bucketName, objectName, file, -1, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", err
	}
	signedUrl, err := minioClient.PresignedGetObject(context.Background(), info.Bucket, info.Key, time.Hour*24*7, nil)
	if err != nil {
		return "", err
	}

	fmt.Printf("Successfully uploaded %s to %s/%s\n", path, bucketName, objectName)
	return signedUrl.String(), nil
}
