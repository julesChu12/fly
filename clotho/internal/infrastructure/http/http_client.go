package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// HTTPClient represents a simple HTTP client wrapper
type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewHTTPClient creates a new HTTP client
func NewHTTPClient(baseURL string, timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Get performs a GET request
func (c *HTTPClient) Get(ctx context.Context, path string, params interface{}, result interface{}) error {
	return c.doRequest(ctx, http.MethodGet, path, params, nil, result)
}

// Post performs a POST request
func (c *HTTPClient) Post(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.doRequest(ctx, http.MethodPost, path, nil, body, result)
}

// Put performs a PUT request
func (c *HTTPClient) Put(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.doRequest(ctx, http.MethodPut, path, nil, body, result)
}

// Delete performs a DELETE request
func (c *HTTPClient) Delete(ctx context.Context, path string, result interface{}) error {
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil, result)
}

// doRequest performs the actual HTTP request
func (c *HTTPClient) doRequest(ctx context.Context, method, path string, params, body, result interface{}) error {
	// Build URL
	fullURL, err := c.buildURL(path, params)
	if err != nil {
		return fmt.Errorf("failed to build URL: %w", err)
	}

	// Prepare request body
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	// Perform request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	// Decode response
	if result != nil {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// buildURL builds the full URL with query parameters
func (c *HTTPClient) buildURL(path string, params interface{}) (string, error) {
	baseURL := c.baseURL
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	// Parse base URL
	parsedURL, err := url.Parse(baseURL + path)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	// Add query parameters
	if params != nil {
		query := parsedURL.Query()

		// Handle map[string]string
		if paramsMap, ok := params.(map[string]string); ok {
			for key, value := range paramsMap {
				query.Set(key, value)
			}
		} else {
			// For other types, marshal to JSON and add as query params
			jsonBytes, err := json.Marshal(params)
			if err != nil {
				return "", fmt.Errorf("failed to marshal params: %w", err)
			}

			var paramsMap map[string]interface{}
			if err := json.Unmarshal(jsonBytes, &paramsMap); err != nil {
				return "", fmt.Errorf("failed to unmarshal params: %w", err)
			}

			for key, value := range paramsMap {
				if strValue, ok := value.(string); ok {
					query.Set(key, strValue)
				} else if strValue, ok := value.(fmt.Stringer); ok {
					query.Set(key, strValue.String())
				}
			}
		}

		parsedURL.RawQuery = query.Encode()
	}

	return parsedURL.String(), nil
}