package httpclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGet_Success(t *testing.T) {
	type Response struct {
		Message string `json:"message"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response{Message: "success"})
	}))
	defer server.Close()

	client := NewHTTPClient(&Config{
		Log: logrus.New(),
		RetryConfig: &RetryConfig{
			MaxRetries: 1,
			BaseDelay:  10 * time.Millisecond,
		},
	})

	var response Response
	err := client.Get(context.Background(), server.URL, &response)

	require.NoError(t, err)
	assert.Equal(t, "success", response.Message)
}

func TestGet_RetryableError(t *testing.T) {
	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount < 2 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		}
	}))
	defer server.Close()

	client := NewHTTPClient(&Config{
		Log: logrus.New(),
		RetryConfig: &RetryConfig{
			MaxRetries: 3,
			BaseDelay:  50 * time.Millisecond,
		},
	})

	var response map[string]interface{}
	err := client.Get(context.Background(), server.URL, &response)

	require.NoError(t, err)
	assert.Equal(t, 2, requestCount)
}

func TestGet_MaxRetriesExceeded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewHTTPClient(&Config{
		Log: logrus.New(),
		RetryConfig: &RetryConfig{
			MaxRetries: 2,
			BaseDelay:  50 * time.Millisecond,
		},
	})

	var response map[string]interface{}
	err := client.Get(context.Background(), server.URL, &response)

	require.Error(t, err)
}

func TestGet_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json{"))
	}))
	defer server.Close()

	client := NewHTTPClient(&Config{
		Log: logrus.New(),
		RetryConfig: &RetryConfig{
			MaxRetries: 1,
			BaseDelay:  10 * time.Millisecond,
		},
	})

	var response map[string]interface{}
	err := client.Get(context.Background(), server.URL, &response)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid character")
}
