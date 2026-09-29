package repository

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
}

func MinioFromEnv() MinioConfig {
	return MinioConfig{
		Endpoint:  os.Getenv("MINIO_ENDPOINT"),
		AccessKey: os.Getenv("MINIO_ACCESS_KEY"),
		SecretKey: os.Getenv("MINIO_SECRET_KEY"),
		Bucket:    os.Getenv("MINIO_BUCKET"),
	}
}

func newMinioClient(config MinioConfig) (*minio.Client, error) {
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}

	exists, err := client.BucketExists(context.Background(), config.Bucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		if err := client.MakeBucket(context.Background(), config.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, err
		}

		policy := fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::%s/*"]}]}`, config.Bucket)
		if err := client.SetBucketPolicy(context.Background(), config.Bucket, policy); err != nil {
			return nil, err
		}
	}

	return client, nil
}

func (r *Repository) UploadDrugMedia(objectName, contentType string, file io.Reader, size int64) (string, error) {
	_, err := r.minio.PutObject(context.Background(), r.minioBucket, objectName, file, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%s/%s/%s", r.minioEndpoint, r.minioBucket, objectName), nil
}
