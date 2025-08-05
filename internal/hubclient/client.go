/*
   Copyright 2024 Docker Terraform Provider authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package hubclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

// TokenProvider provides a valid authentication token for API requests
type TokenProvider interface {
	// EnsureToken returns a valid token, refreshing if necessary
	EnsureToken(ctx context.Context) (token string, err error)
	// Username returns the username associated with this provider (for display purposes)
	Username() string
}

type Client struct {
	BaseURL        string
	HTTPClient     *http.Client
	tokenProvider  TokenProvider
	maxPageResults int64
}

type Config struct {
	BaseURL        string
	TokenProvider  TokenProvider
	Transport      http.RoundTripper
	MaxPageResults int64
}

func NewClient(config Config) *Client {
	baseClient := &http.Client{
		Timeout:   time.Minute,
		Transport: config.Transport,
	}
	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient = baseClient

	return &Client{
		BaseURL:        config.BaseURL,
		HTTPClient:     retryClient.StandardClient(),
		tokenProvider:  config.TokenProvider,
		maxPageResults: config.MaxPageResults,
	}
}

func (c *Client) sendRequest(ctx context.Context, method string, url string, body []byte, result interface{}) error {
	token, err := c.tokenProvider.EnsureToken(ctx)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s%s", c.BaseURL, url)
	req, err := http.NewRequest(method, path, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	req = req.WithContext(ctx)

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		bodyBytes, readErr := io.ReadAll(res.Body)
		if readErr != nil {
			return readErr
		}
		return fmt.Errorf("server response %s: %s", path, string(bodyBytes))
	}

	if result != nil {
		if err = json.NewDecoder(res.Body).Decode(result); err != nil {
			return err
		}
	}

	return nil
}

// convertToRelativeURL converts full URLs to relative paths for sendRequest
func (c *Client) convertToRelativeURL(url string) string {
	if strings.HasPrefix(url, "https://") {
		return strings.TrimPrefix(url, c.BaseURL)
	}
	return url
}

func (c *Client) paginate(ctx context.Context, initialURL string, processPage func(string) (interface{}, error)) error {
	nextURL := initialURL
	pagesFetched := 0

	for nextURL != "" {
		relativeURL := c.convertToRelativeURL(nextURL)

		next, err := processPage(relativeURL)
		if err != nil {
			return err
		}

		pagesFetched++

		// If maxPageResults is set to 0, we will fetch all pages
		if c.maxPageResults > 0 && int64(pagesFetched) >= c.maxPageResults {
			break
		}

		nextURL = ""
		if next != nil {
			if nextStr, ok := next.(string); ok {
				nextURL = nextStr
			}
		}
	}

	return nil
}

func (c *Client) Username() string {
	return c.tokenProvider.Username()
}

func (c *Client) MaxPageResults() int64 {
	return c.maxPageResults
}
