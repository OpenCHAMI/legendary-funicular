// Package bucket
package bucket

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"io"
)

type Bucket struct {
	s3   *s3.Client
	name string
}

func New(cfg aws.Config, region string, endpoint, bucketName string) Bucket {
	return Bucket{
		s3: s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.UsePathStyle = true
			o.Region = region
			o.BaseEndpoint = aws.String(endpoint)
		}),
		name: bucketName,
	}
}

func (b *Bucket) buildObjectPaginator() *s3.ListObjectsV2Paginator {
	return s3.NewListObjectsV2Paginator(
		b.s3,
		&s3.ListObjectsV2Input{Bucket: aws.String(b.name)},
	)
}

func (b *Bucket) List(ctx context.Context) ([]types.Object, error) {
	var objects []types.Object
	objectPaginator := b.buildObjectPaginator()

	for objectPaginator.HasMorePages() {
		output, err := objectPaginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		objects = append(objects, output.Contents...)
	}
	return objects, nil
}

func (b *Bucket) Open(
	ctx context.Context,
	key string,
) (io.ReadCloser, error) {
	if out, err := b.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(b.name),
		Key:    aws.String(key),
	}); err != nil {
		return nil, err
	} else {
		return out.Body, nil
	}
}

func (b *Bucket) Read(
	ctx context.Context,
	key string,
) ([]byte, error) {
	stream, err := b.Open(ctx, key)
	if err != nil {
		return nil, err
	}

	defer stream.Close()
	data, err := io.ReadAll(stream)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (b *Bucket) Put(
	ctx context.Context,
	key string,
	stream io.Reader,
) error {
	_, err := b.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(b.name),
		Key:    aws.String(key),
		Body:   stream,
	})
	return err
}

func (b *Bucket) Delete(ctx context.Context, key string) error {
	_, err := b.s3.DeleteObject(
		ctx,
		&s3.DeleteObjectInput{
			Bucket: aws.String(b.name),
			Key:    aws.String(key),
		},
	)
	return err
}
