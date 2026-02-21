// Package s3client
package s3client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Client wraps an S3 client and exposes simple helpers.
type S3Client struct {
	s3 *s3.Client
}

func NewConfig(
	ctx context.Context,
	accessKey string,
	secretKey string,
	region string,
) (*aws.Config, error) {

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		),
	)

	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	return &cfg, nil
}

func NewS3Client(ctx context.Context, cfg aws.Config, region string, endpoint string) S3Client {
	return S3Client{
		s3: s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.UsePathStyle = true
			o.BaseEndpoint = aws.String(endpoint)
		})}
}

// ListObjects lists the objects in a bucket.
// derived from: https://docs.aws.amazon.com/code-library/latest/ug/go_2_s3_code_examples.html#get_started
func (c *S3Client) ListObjects(ctx context.Context, bucket string) ([]types.Object, error) {
	var err error
	var output *s3.ListObjectsV2Output
	var objects []types.Object
	input := &s3.ListObjectsV2Input{Bucket: aws.String(bucket)}

	objectPaginator := s3.NewListObjectsV2Paginator(c.s3, input)
	for objectPaginator.HasMorePages() {
		output, err = objectPaginator.NextPage(ctx)
		if err != nil {
			var noBucket *types.NoSuchBucket
			if errors.As(err, &noBucket) {
				slog.Warn(fmt.Sprintf("bucket '%s' does not exist.\n", bucket))
				err = noBucket
			}
			break
		} else {
			objects = append(objects, output.Contents...)
		}
	}

	return objects, err
}

// GetObject fetches an object and returns its bytes.
func (c *S3Client) GetObject(
	ctx context.Context,
	bucket string,
	key string,
) ([]byte, error) {

	out, err := c.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}
	defer out.Body.Close()

	data, err := io.ReadAll(out.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	return data, nil
}
