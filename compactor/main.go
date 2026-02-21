package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/klauspost/compress/zstd"
	"github.com/parquet-go/parquet-go"
	"github.com/seantronsen/openchami-logq/compactor/s3client"
)

func getRequiredEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		slog.Error(fmt.Sprintf("undefined required environment variable: '%s'", name))
		os.Exit(1)
	}
	return value
}

func buildAwsCfg(ctx context.Context, s3KeyAccess, s3KeySecret, s3Region string) aws.Config {
	cfg, err := s3client.NewConfig(ctx,
		s3KeyAccess,
		s3KeySecret,
		s3Region,
	)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to build configuration for AWS SDK: '%s'", err))
		os.Exit(1)
	}
	return *cfg
}

type JSONLog struct {
	Timestamp *string `json:"timestamp"`
	Service   *string `json:"appname"`
	Host      *string `json:"host"`
	Severity  *string `json:"severity"`
	Message   *string `json:"message"`

	// chami-specific
	XName       *string `json:"xname"`
	IDComponent *string `json:"component_id"`
	IDNode      *string `json:"node_id"`
	IDTrace     *string `json:"trace_id"`
	IDRequest   *string `json:"request_id"`
	RequestUser *string `json:"request_user"`

	// collector specific
	PayloadJSON *string `json:"payload_json"`
	ParseError  *string `json:"parse_error"`
}

// decompressBytes
// derived from: https://github.com/klauspost/compress/tree/master/zstd#usage-1
func decompressBytes(i []byte) ([]byte, error) {
	br := bytes.NewReader(i)
	d, err := zstd.NewReader(br)
	if err != nil {
		return nil, err
	}

	defer d.Close()

	o, err := io.ReadAll(d)
	if err != nil {
		return nil, err
	}
	return o, nil
}

func main() {

	access := getRequiredEnv("S3_ACCESS_KEY")
	secret := getRequiredEnv("S3_SECRET_KEY")
	region := getRequiredEnv("S3_REGION")
	endpoint := getRequiredEnv("S3_ENDPOINT")
	bucketIn := getRequiredEnv("S3_BUCKET_READ")
	// bucketOut := getRequiredEnv("S3_BUCKET_WRITE")
	// target := getRequiredEnv("S3_OBJECT_TARGET")

	// ctx, cancel := context.WithCancel(context.Background())
	ctx := context.Background()
	cfg := buildAwsCfg(ctx, access, secret, region)
	client := s3client.NewS3Client(ctx, cfg, region, endpoint)

	objects, err := client.ListObjects(ctx, bucketIn)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to list objects for bucket '%s': '%s'", bucketIn, err))
		os.Exit(1)
	}

	slog.Info("found the following objects: ")
	for ix, val := range objects {
		slog.Info(fmt.Sprintf("%04d: %s", ix, *val.Key))
	}

	target := objects[len(objects)-1]
	data, err := client.GetObject(ctx, bucketIn, *target.Key)
	if err != nil {
		slog.Error(fmt.Sprintf("unable to get object: %s, error=%s", *target.Key, err))
		os.Exit(1)
	}
	data, err = decompressBytes(data)
	if err != nil {
		slog.Error(fmt.Sprintf("unable to decompress object: %s, error=%s", *target.Key, err))
		os.Exit(1)
	}

	ndjson := string(data)
	var records []JSONLog
	for val := range strings.Lines(ndjson) {
		record := JSONLog{}
		err = json.Unmarshal([]byte(val), &record)
		if err != nil {
			slog.Error(fmt.Sprintf("failed to deserialize json into go record: %s", err))
			os.Exit(1)
		}
		record.PayloadJSON = &val
		dumb, _ := json.MarshalIndent(record, "", "  ")
		fmt.Printf("record: %+v", string(dumb))
		fmt.Println()
		records = append(records, record)
	}
	if err := parquet.WriteFile("tester.parquet", records); err != nil {
		slog.Error(fmt.Sprintf("failed to write records to parquet format into go record: %s", err))
		os.Exit(1)
	}
}
