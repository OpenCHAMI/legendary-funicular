package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/parquet-go/parquet-go"
	"github.com/seantronsen/openchami-logq/compactor/internal/aws/s3/bucket"
	"github.com/seantronsen/openchami-logq/compactor/internal/iopipe"
	"github.com/seantronsen/openchami-logq/compactor/internal/syslog"
	"github.com/seantronsen/openchami-logq/compactor/internal/util"
	"github.com/seantronsen/openchami-logq/compactor/internal/zio"
)

func main() {

	access := util.GetRequiredEnv("S3_ACCESS_KEY")
	secret := util.GetRequiredEnv("S3_SECRET_KEY")
	region := util.GetRequiredEnv("S3_REGION")
	endpoint := util.GetRequiredEnv("S3_ENDPOINT")
	bucketIn := util.GetRequiredEnv("S3_BUCKET_READ")
	// bucketOut := util.GetRequiredEnv("S3_BUCKET_WRITE")

	// ctx, cancel := context.WithCancel(context.Background())
	ctx := context.Background()
	cfg := util.BuildAwsCfg(ctx, access, secret, region)
	source := bucket.New(cfg, region, endpoint, bucketIn)
	// sink := bucket.New(cfg, region, endpoint, bucketOut)

	objects, err := source.List(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to list objects for bucket '%s': '%s'", bucketIn, err))
		os.Exit(1)
	}

	slog.Info("found the following objects: ")
	for ix, val := range objects {
		slog.Info(fmt.Sprintf("%04d: %s", ix, *val.Key))
	}

	slog.Info("decompressing all ndjson records")

	f, _ := os.Create("dummy.parquet")
	writer := parquet.NewGenericWriter[syslog.Syslog](f, parquet.Compression(&parquet.Zstd))
	for _, val := range objects {
		var records []syslog.Syslog
		i, _ := source.Open(ctx, *val.Key)
		pipe, _ := iopipe.NewR(i, zio.NewUnzstdStream)
		reader := bufio.NewReader(pipe)
		for {
			bytes, err := reader.ReadBytes('\n')
			if err == io.EOF {
				break
			} else if err != nil {
				break // uh oh
			}
			s, _ := syslog.Parse(bytes)
			dumb, _ := json.MarshalIndent(s, "", "  ")
			fmt.Println(string(dumb))
			records = append(records, s)
		}
		writer.Write(records)
		// io.Copy(os.Stdout, pipe)
		pipe.Close()
	}
	writer.Close()
	f.Close()
}
