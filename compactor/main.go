package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"

	"github.com/parquet-go/parquet-go"
	"github.com/seantronsen/openchami-logq/compactor/internal/pipeline"
	"github.com/seantronsen/openchami-logq/compactor/internal/syslog"
	"github.com/seantronsen/openchami-logq/compactor/internal/util"
	"github.com/seantronsen/openchami-logq/compactor/internal/zio"
)

func main() {
	ctx := context.Background()
	source := util.BuildBucket(ctx, util.GetRequiredEnv("S3_BUCKET_READ"))
	sink := util.BuildBucket(ctx, util.GetRequiredEnv("S3_BUCKET_WRITE"))

	target := "events"
	objects, err := source.List(ctx, target)
	util.CheckErr(err)

	slog.Info("decompressing all ndjson records")
	oname := util.BuildObjectName(target)

	var wg sync.WaitGroup
	defer wg.Wait()
	pr, pw := io.Pipe()
	wg.Go(func() {
		n, err := sink.SpooledPut(ctx, oname, pr)
		util.CheckErr(err)
		slog.Info(fmt.Sprintf("wrote %d bytes to spool", n))
		pr.Close()
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
