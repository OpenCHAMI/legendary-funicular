// Package config
package config

import "fmt"

type Config struct {
	S3SSL           bool    `json:"s3_ssl"             secret:"false"`
	S3Region        *string `json:"s3_region"          secret:"false"`
	S3Endpoint      *string `json:"s3_endpoint"        secret:"false"`
	S3BucketNDJSON  *string `json:"s3_bucket_ndjson"   secret:"false"`
	S3BucketParquet *string `json:"s3_bucket_parquet"  secret:"false"`
	S3KeyAccess     *string `json:"s3_key_access"      secret:"true"`
	S3KeySecret     *string `json:"s3_key_secret"      secret:"true"`
}

func ref[T any](x T) *T {
	return &x
}

func req(p *string, name string) error {
	if p == nil || *p == "" {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}

func validate(c *Config) error {
	if c == nil {
		return fmt.Errorf("config is nil")
	}

	if err := req(c.S3Region, "s3_region"); err != nil {
		return err
	}
	if err := req(c.S3Endpoint, "s3_endpoint"); err != nil {
		return err
	}
	if err := req(c.S3BucketNDJSON, "s3_bucket_ndjson"); err != nil {
		return err
	}
	if err := req(c.S3BucketParquet, "s3_bucket_parquet"); err != nil {
		return err
	}
	if err := req(c.S3KeyAccess, "s3_key_access"); err != nil {
		return err
	}
	if err := req(c.S3KeySecret, "s3_key_secret"); err != nil {
		return err
	}

	return nil
}

type Option func(*Config)

func New(opts ...Option) (*Config, error) {
	cfg := &Config{}

	for _, opt := range opts {
		opt(cfg)
	}

	if err := validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil

}

func WithDefaults() Option {
	return func(c *Config) {
		c.S3SSL = false
		c.S3Region = ref("us-east-1")
		c.S3Endpoint = ref("localhost:7070")
		c.S3BucketNDJSON = ref("openchami-logs-raw")
		c.S3BucketParquet = ref("openchami-logs-daily")
	}
}

func WithS3KeyAccess(k string) Option {
	return func(c *Config) {
		c.S3KeyAccess = ref(k)
	}
}

func WithS3KeySecret(k string) Option {
	return func(c *Config) {
		c.S3KeySecret = ref(k)
	}
}

func WithS3SSL(v bool) Option {
	return func(c *Config) {
		c.S3SSL = v
	}
}

func WithS3Region(v string) Option {
	return func(c *Config) {
		c.S3Region = ref(v)
	}
}

func WithS3Endpoint(v string) Option {
	return func(c *Config) {
		c.S3Endpoint = ref(v)
	}
}

func WithS3BucketNDJSON(v string) Option {
	return func(c *Config) {
		c.S3BucketNDJSON = ref(v)
	}
}

func WithS3BucketParquet(v string) Option {
	return func(c *Config) {
		c.S3BucketParquet = ref(v)
	}
}
