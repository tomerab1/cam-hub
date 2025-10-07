package objectstorage

import (
	"context"
	"io"
	"log/slog"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type ContentType = string

const (
	UrlVideo ContentType = "video/mp4"
	UrlPng   ContentType = "image/png"
)

type MinIOStore struct {
	client *minio.Client
	logger *slog.Logger
	ctx    context.Context
}

func NewMinIOStore(ctx context.Context, logger *slog.Logger, useSSL bool) (*MinIOStore, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKeyID := os.Getenv("MINIO_ROOT_USER")
	secretAccessKey := os.Getenv("MINIO_ROOT_PASSWORD")

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	return &MinIOStore{
		client: minioClient,
		logger: logger,
		ctx:    ctx,
	}, nil
}

func (store *MinIOStore) CreateBucket(bucketName string) error {
	if err := store.client.MakeBucket(store.ctx, bucketName, minio.MakeBucketOptions{
		ObjectLocking: true,
	}); err != nil {
		return err
	}

	return nil
}

func (store *MinIOStore) BucketExists(bucketName string) (bool, error) {
	return store.client.BucketExists(store.ctx, bucketName)
}

func (store *MinIOStore) RemoveBucket(bucketName string) error {
	return store.client.RemoveBucket(store.ctx, bucketName)
}

func (store *MinIOStore) CopyObjectWithinBucket(bucket, srcPrefix, dstPrefix, objName string) (minio.UploadInfo, error) {
	srcKey := path.Join(srcPrefix, objName)
	dstKey := path.Join(dstPrefix, objName)

	srcOpts := minio.CopySrcOptions{
		Bucket: bucket,
		Object: srcKey,
	}
	dstOpts := minio.CopyDestOptions{
		Bucket: bucket,
		Object: dstKey,
	}
	return store.client.CopyObject(store.ctx, dstOpts, srcOpts)
}

func (store *MinIOStore) PutObject(bucketName, objName string, reader io.Reader, objSz int64) (minio.UploadInfo, error) {
	return store.client.PutObject(
		store.ctx,
		bucketName,
		objName,
		reader,
		objSz,
		minio.PutObjectOptions{
			ContentType: "application/octet-stream",
		},
	)
}

func (store *MinIOStore) RemoveObjects(bucketName, objPrefix string) error {
	objsCh := make(chan minio.ObjectInfo)

	go func() {
		defer close(objsCh)

		for obj := range store.client.ListObjects(store.ctx, bucketName, minio.ListObjectsOptions{
			Prefix:    objPrefix,
			Recursive: true,
		}) {
			if obj.Err != nil {
				store.logger.Error("removeObjects: failed to list object", "bucket_name", bucketName, "obj_prefix", objPrefix)
				return
			}

			objsCh <- obj
		}
	}()

	for err := range store.client.RemoveObjects(store.ctx, bucketName, objsCh, minio.RemoveObjectsOptions{}) {
		return err.Err
	}

	return nil
}

func (store *MinIOStore) RemoveObject(bucketName, objName string) error {
	return store.client.RemoveObject(store.ctx, bucketName, objName, minio.RemoveObjectOptions{})
}

func (store *MinIOStore) GetObject(bucketName, objectName string) (*minio.Object, error) {
	return store.client.GetObject(store.ctx, bucketName, objectName, minio.GetObjectOptions{})
}

func (store *MinIOStore) PresignedViewUrl(bucketName, objName string, contentType ContentType, expiry time.Duration) (string, error) {
	urlVals := make(url.Values)
	urlVals.Set("response-content-type", contentType)
	urlVals.Set("response-content-disposition", `inline; filename="`+objName+`"`)

	url, err := store.client.PresignedGetObject(store.ctx, bucketName, objName, expiry, urlVals)
	if err != nil {
		return "", err
	}

	return url.String(), nil
}
