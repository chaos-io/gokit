package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	jsoniter "github.com/json-iterator/go"
)

type Client struct {
	client *http.Client
}

func NewClient() IClient {
	return &Client{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) DoHTTPRequest(ctx context.Context, requestParam *RequestParam) error {
	if requestParam == nil {
		return fmt.Errorf("request param is nil")
	}

	// set timeout
	if requestParam.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, requestParam.Timeout)
		defer cancel()
	}

	// parse request
	var body io.Reader
	if requestParam.Body != nil {
		// use in direct when body is io.Reader
		if reader, ok := requestParam.Body.(io.Reader); ok {
			body = reader
		} else {
			bodyBytes, err := jsoniter.Marshal(requestParam.Body)
			if err != nil {
				return fmt.Errorf("failed to marshal request body: %w", err)
			}
			body = bytes.NewReader(bodyBytes)
		}
	}

	req, err := http.NewRequestWithContext(ctx, requestParam.Method, requestParam.RequestURI, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// set header
	if requestParam.Header != nil {
		for key, value := range requestParam.Header {
			req.Header.Set(key, value)
		}
	}

	// set Content-Type when body is not nil and Content-Type is empty
	if requestParam.Body != nil && req.Header.Get("Content-Type") == "" {
		if _, ok := requestParam.Body.(io.Reader); ok {
			req.Header.Set("Content-Type", "text/plain")
		} else {
			req.Header.Set("Content-Type", "application/json")
		}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// check HTTP status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// parse response
	if requestParam.Response != nil {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		if err := jsoniter.ConfigFastest.Unmarshal(bodyBytes, requestParam.Response); err != nil {
			return fmt.Errorf("failed to unmarshal response body: %w", err)
		}
	}

	return nil
}
