package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"

	"github.com/parquet-go/parquet-go"
	"github.com/seantronsen/openchami-logq/compactor/internal/aws/s3/bucket"
	"github.com/seantronsen/openchami-logq/compactor/internal/cloudevent"
	"github.com/seantronsen/openchami-logq/compactor/internal/pipeline"
	"github.com/seantronsen/openchami-logq/compactor/internal/syslog"
	"github.com/seantronsen/openchami-logq/compactor/internal/util"
	"github.com/seantronsen/openchami-logq/compactor/internal/zio"
)

// return the objects which were successfully compacted
// return listing should be the full object path including the prefix (but not th ebucket name)

type Record interface {
	syslog.Syslog | cloudevent.CloudEvent
}
type Parser[T Record] func(b []byte) (T, error)

func recordsFromStream[T Record](stream io.Reader, parser Parser[T]) ([]T, error) {
	var records []T
	reader := bufio.NewReader(stream)
	for {
		bytes, err := reader.ReadBytes('\n')
		if err == io.EOF || len(bytes) == 0 {
			break
		}
		if err != nil {
			return records, err
		}
		s, err := parser(bytes)
		if err != nil {
			return records, err
		}
		records = append(records, s)
	}

	return records, nil
}

func cvtS3Objects2Parquet[T Record](
	ctx context.Context,
	source bucket.Bucket,
	keys []string,
	parser Parser[T],
	out io.Writer,
) ([]string, error) {

	var objectsSink []string
	writer := parquet.NewGenericWriter[T](out, parquet.Compression(&parquet.Zstd))
	defer writer.Close()

	// attempt to add all records from input objects to the outgoing parquet stream
	// if failure occurs for a single object, continue without adding those records and log the error
	// e.g., leave that object for later after the subsequent bug fix
	for _, key := range keys {
		i, err := source.Open(ctx, key)
		if err != nil {
			return objectsSink, err
		}

		pipe, err := pipeline.NewR(i, zio.NewUnzstdStream)
		if err != nil {
			return objectsSink, err
		}

		records, err := recordsFromStream(pipe, parser)
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

func compaction[T Record](
	ctx context.Context,
	source, sink bucket.Bucket,
	prefix string,
	parser Parser[T],
) error {
	var err error
	var keySink []string
	keySrc, err := source.List(ctx, prefix)
	if err != nil || len(keySrc) == 0 {
		return err
	}

	var wg sync.WaitGroup
	pipeRead, pipeWrite := io.Pipe()
	consumerErrCh := make(chan error, 1)
	producerErrCh := make(chan error, 1)
	keyCh := make(chan string)

	// S3 Reader
	wg.Go(func() {
		defer pipeWrite.Close()
		defer close(keyCh)
		defer close(producerErrCh)

		objectsSink, err := cvtS3Objects2Parquet(ctx, source, keySrc, parser, pipeWrite)
		if err != nil {
			producerErrCh <- err
			return
		}
		for _, key := range objectsSink {
			keyCh <- key
		}
	})

	// S3 Writer
	wg.Go(func() {
		defer pipeRead.Close()
		defer close(consumerErrCh)

		key := util.BuildObjectName(prefix)
		n, err := sink.SpooledPut(ctx, key, pipeRead)
		if err != nil {
			consumerErrCh <- err
			return
		}
		slog.Info(fmt.Sprintf("wrote %d bytes to s3://%s/%s", n, sink.Name, key))
	})

	for key := range keyCh {
		keySink = append(keySink, key)
	}
	if err := <-producerErrCh; err != nil {
		return err
	}
	if err := <-consumerErrCh; err != nil {
		return err
	}
	wg.Wait()

	for _, key := range keySink {
		if e := source.Delete(ctx, key); e != nil {
			err = e
		} else {
			slog.Debug(fmt.Sprintf("deleted %s", key))
		}
	}
	return err
}

func main() {
	slog.SetDefault(slog.New(
		slog.NewTextHandler(
			os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug},
		)))

	ctx := context.Background()
	source := util.BuildBucket(ctx, util.GetRequiredEnv("S3_BUCKET_READ"))
	sink := util.BuildBucket(ctx, util.GetRequiredEnv("S3_BUCKET_WRITE"))

	var err error
	err = compaction(ctx, source, sink, "logs", syslog.Parse)
	util.CheckErr(err, true)
	err = compaction(ctx, source, sink, "events", cloudevent.Parse)
	util.CheckErr(err, true)
}
