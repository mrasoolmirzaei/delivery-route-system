package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/avast/retry-go/v4"

	"github.com/sirupsen/logrus"
)

const (
	defaultTimeout    = 3 * time.Second
	defaultMaxRetries = 10
	defaultRetryDelay = 100 * time.Millisecond
)

type HTTPClient struct {
	client     *http.Client
	log        logrus.FieldLogger
	maxRetries uint
	retryDelay time.Duration
}

type Config struct {
	Log         logrus.FieldLogger
	RetryConfig *RetryConfig
	Timeout     time.Duration
}

type RetryConfig struct {
	MaxRetries uint
	BaseDelay  time.Duration
}

func NewHTTPClient(cfg *Config) *HTTPClient {
	timeout := defaultTimeout
	if cfg.Timeout > 0 {
		timeout = cfg.Timeout
	}

	client := &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
		log:        cfg.Log,
		maxRetries: defaultMaxRetries,
		retryDelay: defaultRetryDelay,
	}

	if cfg.RetryConfig != nil {
		client.maxRetries = cfg.RetryConfig.MaxRetries
		client.retryDelay = cfg.RetryConfig.BaseDelay
	}

	return client
}

// shouldRetryOnStatus checks if we should retry based on status code
func (c *HTTPClient) shouldRetryOnStatus(statusCode int) bool {
	return statusCode >= http.StatusInternalServerError
}

func (c *HTTPClient) Get(ctx context.Context, url string, response any) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		c.log.WithError(err).Errorf("failed to create request to get %s", url)
		return err
	}

	var resp *http.Response
	err = retry.Do(
		func() error {
			resp, err = c.client.Do(req)
			if err != nil {
				return err
			}
			if c.shouldRetryOnStatus(resp.StatusCode) {
				return fmt.Errorf("retryable status code: %d", resp.StatusCode)
			}
			return nil
		},
		retry.Attempts(c.maxRetries),
		retry.Delay(c.retryDelay),
		retry.DelayType(retry.BackOffDelay),
		retry.LastErrorOnly(true),
		retry.Context(ctx),
		retry.OnRetry(func(n uint, err error) {
			c.log.WithError(err).Errorf("failed to get %s : %s", url, resp.Status)
		}),
	)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.log.Errorf("failed to get %s : %s", url, resp.Status)
		return err
	}

	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		c.log.WithError(err).Errorf("failed to decode %s", url)
		return err
	}

	return nil
}
