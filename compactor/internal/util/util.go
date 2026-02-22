// Package util
package util

// todo: move this out of internal and into main package, but different file named utils.go still
import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/seantronsen/openchami-logq/compactor/internal/aws/config"
	"github.com/seantronsen/openchami-logq/compactor/internal/aws/s3/bucket"
)

func GetRequiredEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		slog.Error(fmt.Sprintf("undefined required environment variable: '%s'", name))
		os.Exit(1)
	}
	return value
}

func BuildBucket(ctx context.Context, bucketName string) bucket.Bucket {
	access := GetRequiredEnv("S3_ACCESS_KEY")
	secret := GetRequiredEnv("S3_SECRET_KEY")
	region := GetRequiredEnv("S3_REGION")
	endpoint := GetRequiredEnv("S3_ENDPOINT")
	cfg, err := config.New(ctx, access, secret, region)
	CheckErr(err, true)

	return bucket.New(cfg, region, endpoint, bucketName)
}

func BuildObjectName(prefix string) string {
	return fmt.Sprintf(
		"%s/date=%s/%s.parquet",
		prefix,
		time.Now().UTC().Format(time.DateOnly),
		uuid.NewString(),
	)
}

func CheckErr(err error, shouldExit bool) {
	if err != nil {
		if shouldExit {
			slog.Error(fmt.Sprint(err))
			os.Exit(1)
		} else {
			slog.Warn(fmt.Sprint(err))
		}
	}
}
