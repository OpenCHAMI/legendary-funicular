// Package util
package util

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/seantronsen/openchami-logq/compactor/internal/aws/config"
	"log/slog"
	"os"
)

func GetRequiredEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		slog.Error(fmt.Sprintf("undefined required environment variable: '%s'", name))
		os.Exit(1)
	}
	return value
}

func BuildAwsCfg(ctx context.Context, s3KeyAccess, s3KeySecret, s3Region string) aws.Config {
	cfg, err := config.New(
		ctx,
		s3KeyAccess,
		s3KeySecret,
		s3Region,
	)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to build configuration for AWS SDK: '%s'", err))
		os.Exit(1)
	}
	return cfg
}
