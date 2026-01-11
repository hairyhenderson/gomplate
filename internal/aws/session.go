package aws

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/hairyhenderson/gomplate/v4/env"
)

var (
	co            ClientOptions
	coInit        sync.Once
	sdkConfig     aws.Config
	sdkConfigInit sync.Once
)

// ClientOptions -
type ClientOptions struct {
	Timeout time.Duration
}

// GetClientOptions - Centralised reading of AWS_TIMEOUT
// ... but cannot use in vault/auth.go as different strconv.Atoi error handling
func GetClientOptions() ClientOptions {
	coInit.Do(func() {
		timeout := env.Getenv("AWS_TIMEOUT")
		if timeout == "" {
			timeout = "500"
		}

		t, err := strconv.Atoi(timeout)
		if err != nil {
			panic(fmt.Errorf("invalid AWS_TIMEOUT value '%s' - must be an integer: %w", timeout, err))
		}

		co.Timeout = time.Duration(t) * time.Millisecond
	})
	return co
}

// SDKConfig -
func SDKConfig(region ...string) aws.Config {
	sdkConfigInit.Do(func() {
		options := GetClientOptions()
		timeout := options.Timeout
		if timeout == 0 {
			timeout = 500 * time.Millisecond
		}

		opts := []func(*config.LoadOptions) error{
			config.WithHTTPClient(&http.Client{Timeout: timeout}),
		}

		if env.Getenv("AWS_ANON") == "true" {
			opts = append(opts, config.WithCredentialsProvider(aws.AnonymousCredentials{}))
		}

		if len(region) > 0 && region[0] != "" {
			opts = append(opts, config.WithRegion(region[0]))
		}

		cfg, err := config.LoadDefaultConfig(context.Background(), opts...)
		if err != nil {
			panic(fmt.Errorf("failed to load config: %w", err))
		}

		sdkConfig = cfg
	})

	return sdkConfig
}
