package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/seantronsen/openchami-logq/compactor/internal/record/cloudevent"
	"github.com/seantronsen/openchami-logq/compactor/internal/record/syslog"
)

func main() {
	slog.SetDefault(slog.New(
		slog.NewTextHandler(
			os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug},
		)))

	ctx := context.Background()
	source := buildBucket(ctx, getRequiredEnv("S3_BUCKET_READ"))
	sink := buildBucket(ctx, getRequiredEnv("S3_BUCKET_WRITE"))

	var err error
	err = compaction(ctx, source, sink, "logs", syslog.Parse)
	checkErr(err, true)
	err = compaction(ctx, source, sink, "events", cloudevent.Parse)
	checkErr(err, true)
}
