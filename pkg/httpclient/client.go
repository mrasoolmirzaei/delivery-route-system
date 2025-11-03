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
	defaultTimeout         = 3 * time.Second
	defaultMaxRetries      = 10
	defaultRetryDelay      = 100 * time.Millisecond
	defaultMaxIdleConns    = 100
	defaultMaxConnsPerHost = 10
	defaultIdleConnTimeout = 90 * time.Second
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

	transport := &http.Transport{
		MaxIdleConns:       defaultMaxIdleConns,
		MaxConnsPerHost:    defaultMaxConnsPerHost,
		IdleConnTimeout:    defaultIdleConnTimeout,
		DisableKeepAlives:  false,
		DisableCompression: false,
	}

	client := &HTTPClient{
		client: &http.Client{
			Timeout:   timeout,
			Transport: transport,
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
			var doErr error
			resp, doErr = c.client.Do(req)
			if doErr != nil {
				return doErr
			}
			if resp == nil {
				return fmt.Errorf("response is nil")
			}

			if c.shouldRetryOnStatus(resp.StatusCode) {
				resp.Body.Close()
				return fmt.Errorf("retryable status code: %d", resp.StatusCode)
			}
			return nil
		},
		retry.Attempts(c.maxRetries),
		retry.Delay(c.retryDelay),
		retry.DelayType(retry.BackOffDelay),
		retry.LastErrorOnly(true),
		retry.Context(ctx),
		retry.OnRetry(func(n uint, retryErr error) {
			if resp != nil {
				c.log.WithError(retryErr).Warnf("retry attempt %d for %s (status: %d)", n+1, url, resp.StatusCode)
				resp.Body.Close()
				resp = nil
			} else {
				c.log.WithError(retryErr).Warnf("retry attempt %d for %s", n+1, url)
			}
		}),
	)

	if err != nil {
		return fmt.Errorf("failed to get %s after retries: %w", url, err)
	}

	if resp == nil {
		return fmt.Errorf("response is nil after retries for %s", url)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.log.Errorf("failed to get %s : %s", url, resp.Status)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		c.log.WithError(err).Errorf("failed to decode response from %s", url)
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
