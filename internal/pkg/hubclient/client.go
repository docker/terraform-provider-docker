// API client for hub.docker.com
package hubclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Auth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Token struct {
	Token string `json:"token"`
}

type Client struct {
	BaseURL    string
	auth       Auth
	HTTPClient *http.Client
	userAgent  string
}

type Config struct {
	Host             string
	Username         string
	Password         string
	UserAgentVersion string
}

// Create the API client, providing the authentication.
func NewClient(config Config) *Client {
	version := config.UserAgentVersion
	if version == "" {
		version = "dev"
	}

	return &Client{
		BaseURL: config.Host,
		auth: Auth{
			Username: config.Username,
			Password: config.Password,
		},
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
		userAgent: fmt.Sprintf("terraform-provider-docker/%s", version),
	}
}

func (c *Client) sendRequest(ctx context.Context, method string, url string, body []byte, result interface{}) error {
	authJSON, err := json.Marshal(c.auth)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/users/login/", c.BaseURL), bytes.NewBuffer(authJSON))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")

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
		return errors.New(string(bodyBytes))
	}
	token := Token{}
	if err = json.NewDecoder(res.Body).Decode(&token); err != nil {
		return err
	}

	req, err = http.NewRequest(method, fmt.Sprintf("%s%s", c.BaseURL, url), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.Token))

	req = req.WithContext(ctx)

	res, err = c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		bodyBytes, readErr := io.ReadAll(res.Body)
		if readErr != nil {
			return readErr
		}
		return errors.New(string(bodyBytes))
	}

	if result != nil {
		if err = json.NewDecoder(res.Body).Decode(result); err != nil {
			return err
		}
	}

	return nil
}
