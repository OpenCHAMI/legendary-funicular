package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/parquet-go/parquet-go"
	"github.com/seantronsen/openchami-logq/compactor/internal/aws/s3/bucket"
	"github.com/seantronsen/openchami-logq/compactor/internal/pipeline"
	"github.com/seantronsen/openchami-logq/compactor/internal/syslog"
	"github.com/seantronsen/openchami-logq/compactor/internal/util"
	"github.com/seantronsen/openchami-logq/compactor/internal/zio"
)

func buildBucket(ctx context.Context, bucketName string) bucket.Bucket {
	access := util.GetRequiredEnv("S3_ACCESS_KEY")
	secret := util.GetRequiredEnv("S3_SECRET_KEY")
	region := util.GetRequiredEnv("S3_REGION")
	endpoint := util.GetRequiredEnv("S3_ENDPOINT")
	cfg := util.BuildAwsCfg(ctx, access, secret, region)
	return bucket.New(cfg, region, endpoint, bucketName)
}

func main() {
	ctx := context.Background()
	source := buildBucket(ctx, util.GetRequiredEnv("S3_BUCKET_READ"))
	sink := buildBucket(ctx, util.GetRequiredEnv("S3_BUCKET_WRITE"))

	// objects, err := source.List(ctx, "")
	objects, err := source.List(ctx, "events")
	if err != nil {
		slog.Error(fmt.Sprintf("failed to list objects for bucket '%s'\n%s", source.Name, err))
		os.Exit(1)
	}

	slog.Info("decompressing all ndjson records")
	oname := uuid.NewString() + ".parquet"

	var wg sync.WaitGroup
	defer wg.Wait()
	pr, pw := io.Pipe()
	wg.Go(func() {
		spool, err := os.CreateTemp("", "spool-*")
		if err != nil {
			slog.Error(fmt.Sprint(err))
			os.Exit(1)
		}
		defer spool.Close()
		defer os.Remove(spool.Name())

		n, err := io.Copy(spool, pr)
		if err != nil {
			slog.Error(fmt.Sprint(err))
			os.Exit(1)
		} else {
			slog.Info(fmt.Sprintf("wrote %d bytes to spool", n))
			pr.Close()
			spool.Seek(0, io.SeekStart)
		}

		err = sink.Put(ctx, oname, spool)
		if err != nil {
			slog.Error(fmt.Sprint(err))
			os.Exit(1)
		}
	})
	wg.Go(func() {
		writer := parquet.NewGenericWriter[syslog.Syslog](pw, parquet.Compression(&parquet.Zstd))
		for _, val := range objects {
			var records []syslog.Syslog
			i, _ := source.Open(ctx, *val.Key)
			pipe, _ := pipeline.NewR(i, zio.NewUnzstdStream)
			reader := bufio.NewReader(pipe)
			for {
				bytes, err := reader.ReadBytes('\n')
				if err == io.EOF {
					break
				} else if err != nil {
					slog.Warn(fmt.Sprint(err))
					break // uh oh
				}
				s, _ := syslog.Parse(bytes)
				records = append(records, s)
			}
			writer.Write(records)
			pipe.Close()
			slog.Info(fmt.Sprintf("successful compaction of: %s", *val.Key))
		}
		writer.Close()
		pw.Close()
	})
}
