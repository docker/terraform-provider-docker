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
	"sync"
	"time"

	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/hashicorp/go-retryablehttp"
)

type Auth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Token struct {
	Token string `json:"token"`
}

type Client struct {
	BaseURL     string
	auth        Auth
	HTTPClient  *http.Client
	token       string
	tokenExpiry time.Time
	mu          sync.Mutex
}

type roundTripper struct {
	userAgent string
}

func (rt *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", rt.userAgent)
	return http.DefaultTransport.RoundTrip(req)
}

type Config struct {
	BaseURL          string
	Username         string
	Password         string
	UserAgentVersion string
}

func NewClient(config Config) *Client {
	version := config.UserAgentVersion
	if version == "" {
		version = "dev"
	}

	baseClient := &http.Client{
		Timeout: time.Minute,
		Transport: &roundTripper{
			userAgent: fmt.Sprintf("terraform-provider-docker/%s", version),
		},
	}
	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient = baseClient

	return &Client{
		BaseURL: config.BaseURL,
		auth: Auth{
			Username: config.Username,
			Password: config.Password,
		},
		HTTPClient: retryClient.StandardClient(),
	}
}

// parseTokenExpiration parses the JWT token to get the exact expiration time using go-jose.
func parseTokenExpiration(tokenString string) (time.Time, error) {
	token, err := jwt.ParseSigned(tokenString)
	if err != nil {
		return time.Time{}, err
	}

	claims := jwt.Claims{}
	if err := token.UnsafeClaimsWithoutVerification(&claims); err != nil {
		return time.Time{}, err
	}

	if claims.Expiry != nil {
		return claims.Expiry.Time(), nil
	}

	return time.Time{}, fmt.Errorf("could not find expiration in token")
}

func (c *Client) ensureValidToken(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != "" && time.Now().Before(c.tokenExpiry) {
		return nil
	}

	authJSON, err := json.Marshal(c.auth)
	if err != nil {
		return fmt.Errorf("decode auth settings: %v", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/users/login/", c.BaseURL), bytes.NewBuffer(authJSON))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctx)

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("check credentials: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("check credentials status: %s", res.Status)
	}

	token := Token{}
	if err = json.NewDecoder(res.Body).Decode(&token); err != nil {
		return fmt.Errorf("check credentials decode response: %v", err)
	}

	// Parse the exact expiration time from the token
	expirationTime, err := parseTokenExpiration(token.Token)
	if err != nil {
		return fmt.Errorf("check credentials expiry: %v", err)
	}

	// Store the new token and its exact expiration time
	c.token = token.Token
	c.tokenExpiry = expirationTime

	return nil
}

func (c *Client) sendRequest(ctx context.Context, method string, url string, body []byte, result interface{}) error {
	if err := c.ensureValidToken(ctx); err != nil {
		return err
	}

	path := fmt.Sprintf("%s%s", c.BaseURL, url)
	req, err := http.NewRequest(method, path, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

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

func (c *Client) Username() string {
	return c.auth.Username
}
