package boxhttp

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Global HTTP client to reuse connections
var client = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
	Timeout: 30 * time.Second, // Default timeout, can be overridden in requests
}

// setHeaders sets custom headers to the request
func setHeaders(req *http.Request, headers map[string]string) {
	for k, v := range headers {
		req.Header.Set(k, v)
	}
}

// handleResponse handles reading the body and closing it
func handleResponse(resp *http.Response) ([]byte, int, error) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}
	return body, resp.StatusCode, nil
}

// PostJSON sends a POST request with a JSON payload
func PostJSON(url string, js []byte, headers map[string]string) ([]byte, int, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(js))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	setHeaders(req, headers)

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	return handleResponse(resp)
}

// PostForm sends a POST request with form data
func PostForm(url string, form url.Values, headers map[string]string) ([]byte, int, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	setHeaders(req, headers)

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	return handleResponse(resp)
}

// Get sends a GET request with a custom timeout
func Get(url string, timeout time.Duration, headers map[string]string) ([]byte, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, 0, err
	}
	setHeaders(req, headers)

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	return handleResponse(resp)
}
