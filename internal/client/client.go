package client

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func New(baseURL string, token string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		token:      token,
		httpClient: &http.Client{},
	}
}

// Do executes an HTTP request and returns the raw response body, status code, and any error.
func (c *Client) Do(method string, path string, body []byte, query map[string]string) ([]byte, int, error) {
	return c.DoWithPathParams(method, path, nil, body, query)
}

// DoWithPathParams executes an HTTP request, substituting path parameters, and returns the raw response body.
func (c *Client) DoWithPathParams(method string, path string, pathParams map[string]string, body []byte, query map[string]string) ([]byte, int, error) {
	if c.token == "" {
		return nil, 0, fmt.Errorf("no authentication token provided (use --token, PDT_TOKEN env, or pdt config set token)")
	}

	for key, value := range pathParams {
		path = strings.ReplaceAll(path, "{"+key+"}", value)
	}

	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if query != nil {
		q := req.URL.Query()
		for key, value := range query {
			q.Set(key, value)
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response: %w", err)
	}

	return respBody, resp.StatusCode, nil
}
