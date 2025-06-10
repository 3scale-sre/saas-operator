package util

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	AWSAccessKeyEnvvar string = "AWS_ACCESS_KEY_ID"
	AWSSecretKeyEnvvar string = "AWS_SECRET_ACCESS_KEY"
	AWSRegionEnvvar    string = "AWS_REGION"
)

func S3Client(ctx context.Context, accessKeyID, secretAccessKey, region string, serviceEndpoint *string) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
	)
	if err != nil {
		return nil, err
	}

	if serviceEndpoint != nil {
		return s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = serviceEndpoint
			o.UsePathStyle = true
		}), nil
	} else {
		return s3.NewFromConfig(cfg), nil
	}
}
