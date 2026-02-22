package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"

	"github.com/seantronsen/openchami-logq/compactor/internal/aws/s3/bucket"
	"github.com/seantronsen/openchami-logq/compactor/internal/record"
)

func compaction[T record.Record](
	ctx context.Context,
	source, sink bucket.Bucket,
	prefix string,
	parser record.Parser[T],
) error {
	var err error
	var keySink []string
	keySrc, err := source.List(ctx, prefix)
	if err != nil {
		return err
	}
	if len(keySrc) == 0 {
		slog.Debug(fmt.Sprintf("nothing to do for prefix '%s'", prefix))
		return nil
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

		key := buildObjectKey(prefix)
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
