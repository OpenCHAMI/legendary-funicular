// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package cmd

import (
	"github.com/OpenCHAMI/legendary-funicular/query/internal/config"
	"github.com/OpenCHAMI/legendary-funicular/query/internal/utils"
)

func BuildCfg() *config.Config {
	var opts []config.Option
	opts = append(opts, config.WithDefaults())
	opts = append(opts, config.WithS3KeyAccess(utils.GetEnvFatal("S3_ACCESS_KEY")))
	opts = append(opts, config.WithS3KeySecret(utils.GetEnvFatal("S3_SECRET_KEY")))

	if value, err := utils.GetEnv("S3_REGION"); err != nil && value != "" {
		opts = append(opts, config.WithS3Region(value))
	}
	if value, err := utils.GetEnv("S3_ENDPOINT"); err != nil && value != "" {
		opts = append(opts, config.WithS3Endpoint(value))
	}
	if value, err := utils.GetEnv("S3_SSL"); err != nil && value != "" {
		opts = append(opts, config.WithS3SSL(value == "true"))
	}
	if value, err := utils.GetEnv("S3_BUCKET_NDJSON"); err != nil && value != "" {
		opts = append(opts, config.WithS3BucketNDJSON(value))
	}
	if value, err := utils.GetEnv("S3_BUCKET_PARQUET"); err != nil && value != "" {
		opts = append(opts, config.WithS3BucketParquet(value))
	}

	cfg, err := config.New(opts...)
	utils.CheckFatal(err)

	return cfg
}
