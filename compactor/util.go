package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/parquet-go/parquet-go"
	"github.com/seantronsen/openchami-logq/compactor/internal/aws/config"
	"github.com/seantronsen/openchami-logq/compactor/internal/aws/s3/bucket"
	"github.com/seantronsen/openchami-logq/compactor/internal/pipeline"
	"github.com/seantronsen/openchami-logq/compactor/internal/record"
	"github.com/seantronsen/openchami-logq/compactor/internal/zio"
)

func checkErr(err error, shouldExit bool) {
	if err != nil {
		if shouldExit {
			slog.Error(fmt.Sprint(err))
			os.Exit(1)
		} else {
			slog.Warn(fmt.Sprint(err))
		}
	}
}

func getRequiredEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		slog.Error(fmt.Sprintf("undefined required environment variable: '%s'", name))
		os.Exit(1)
	}
	return value
}

func buildObjectKey(prefix string) string {
	return fmt.Sprintf(
		"%s/date=%s/%s.parquet",
		prefix,
		time.Now().UTC().Format(time.DateOnly),
		uuid.NewString(),
	)
}

func buildBucket(ctx context.Context, bucketName string) bucket.Bucket {
	access := getRequiredEnv("S3_ACCESS_KEY")
	secret := getRequiredEnv("S3_SECRET_KEY")
	region := getRequiredEnv("S3_REGION")
	endpoint := getRequiredEnv("S3_ENDPOINT")
	cfg, err := config.New(ctx, access, secret, region)
	checkErr(err, true)

	return bucket.New(cfg, region, endpoint, bucketName)
}

// attempt to add all records from input objects to the outgoing parquet stream
// if failure occurs for a single object, continue without adding those records and log the error
// e.g., leave that object for later after the subsequent bug fix
func cvtS3Objects2Parquet[T record.Record](
	ctx context.Context,
	source bucket.Bucket,
	keys []string,
	parser record.Parser[T],
	out io.Writer,
) ([]string, error) {

	var objectsSink []string
	writer := parquet.NewGenericWriter[T](out, parquet.Compression(&parquet.Zstd))
	defer writer.Close()

	for _, key := range keys {
		i, err := source.Open(ctx, key)
		if err != nil {
			return objectsSink, err
		}

		pipe, err := pipeline.NewR(i, zio.NewUnzstdStream)
		if err != nil {
			return objectsSink, err
		}

		records, err := record.FromStream(pipe, parser)
		if err != nil {
			slog.Warn(fmt.Sprintf("skipping %s, reason: %s", key, err))
			continue
		}

		_, err = writer.Write(records)
		if err != nil {
			return objectsSink, err
		}

		pipe.Close()
		objectsSink = append(objectsSink, key)
		slog.Debug(fmt.Sprintf("compacted %s", key))
	}
	return objectsSink, nil
}
