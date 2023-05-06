package storage

import (
	"context"
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
	client     *minio.Client
}

func (bucket *ImageBucket) getClient() (*minio.Client, error) {
	if bucket.client != nil {
		return bucket.client, nil
	}

	endpoint := bucket.Endpoint
	accessKey := bucket.AccessKey
	secretKey := bucket.Secret
	useSSL := false

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	bucket.client = minioClient
	return minioClient, nil
}

func (bucket *ImageBucket) PutObject(runId string, path string) (string, error) {
	minioClient, err := bucket.getClient()
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

	return info.Key, nil
}

func (bucket *ImageBucket) GetSignedUrl(key string) (string, error) {
	expires := time.Hour * 24 * 7
	minioClient, err := bucket.getClient()
	if err != nil {
		return "", err
	}
	signedUrl, err := minioClient.PresignedGetObject(context.Background(), bucket.BucketName, key, expires, nil)
	if err != nil {
		return "", err
	}
	return signedUrl.String(), nil
}
