package uploader

import (
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type Config struct {
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	AwsRegion       string
	Endpoint        string
}

func (conf *Config) Validate() error {
	if conf.AwsRegion == "" && conf.Endpoint == "" {
		return errors.New("S3 region or endpoint is required.")
	}

	if conf.BucketName == "" {
		return errors.New("S3 bucket is required.")
	}

	if conf.AccessKeyID == "" {
		return errors.New("AccessKeyID is required.")
	}

	if conf.SecretAccessKey == "" {
		return errors.New("SecretAccessKey is required.")
	}
	return nil
}

func Upload(conf *Config, file io.Reader, fileName string) (string, error) {
	s3Config := aws.NewConfig().
		WithCredentials(credentials.NewStaticCredentials(conf.AccessKeyID, conf.SecretAccessKey, ""))

	forcePathStyle := true
	s3Config.S3ForcePathStyle = &forcePathStyle

	if conf.AwsRegion == "" {
		s3Config = s3Config.WithRegion(" ").WithEndpoint(conf.Endpoint)
	} else {
		s3Config = s3Config.WithRegion(conf.AwsRegion)
	}

	sess := session.New(s3Config)

	uploader := s3manager.NewUploader(sess)

	upParams := &s3manager.UploadInput{
		ACL:    aws.String("public-read"),
		Bucket: &conf.BucketName,
		Key:    &fileName,
		Body:   file,
	}

	result, err := uploader.Upload(upParams)
	if err != nil {
		return "", fmt.Errorf("Failed to upload file, err: %s", err.Error())
	}

	return result.Location, nil
}
