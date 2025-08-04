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

package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// LoginTokenProvider uses static username/password to obtain tokens via login API
// The name of this struct was specifically chosen
// to troll iam-team :)
type LoginTokenProvider struct {
	username    string
	password    string
	baseURL     string
	httpClient  *http.Client
	token       string
	tokenExpiry time.Time
	mu          sync.Mutex
}

// NewLoginTokenProvider creates a token provider that uses username/password
func NewLoginTokenProvider(username, password, baseURL string) *LoginTokenProvider {
	return &LoginTokenProvider{
		username:   username,
		password:   password,
		baseURL:    baseURL,
		httpClient: http.DefaultClient,
	}
}

// EnsureToken returns a cached token if valid, otherwise authenticates with
// username/password to get a new token from the Docker Hub API.
func (p *LoginTokenProvider) EnsureToken(ctx context.Context) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Return cached token if still valid
	if p.token != "" && time.Now().Before(p.tokenExpiry) {
		return p.token, nil
	}

	// Request new token
	auth := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: p.username,
		Password: p.password,
	}

	authJSON, err := json.Marshal(auth)
	if err != nil {
		return "", fmt.Errorf("marshal auth: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/users/login", p.baseURL), bytes.NewBuffer(authJSON))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("login request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("login failed: %s", res.Status)
	}

	var tokenResponse struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(res.Body).Decode(&tokenResponse); err != nil {
		return "", fmt.Errorf("decode token response: %v", err)
	}

	// Parse token expiry
	claims, err := getClaims(tokenResponse.Token)
	if err != nil {
		return "", fmt.Errorf("parse token claims: %v", err)
	}
	if claims.Expiry == nil {
		return "", fmt.Errorf("token does not contain expiry")
	}

	// Cache the token
	p.token = tokenResponse.Token
	p.tokenExpiry = claims.Expiry.Time()

	return p.token, nil
}

func (p *LoginTokenProvider) Username() string {
	return p.username
}

// AccessTokenProvider uses access tokens directly from the credential store
type AccessTokenProvider struct {
	configKey      string
	cachedUsername string
	configStore    *ConfigStore
	mu             sync.Mutex
}

// NewAccessTokenProvider creates a token provider that uses access tokens from the credential store
func NewAccessTokenProvider(configStore *ConfigStore, configKey string) *AccessTokenProvider {
	return &AccessTokenProvider{
		configKey:   configKey,
		configStore: configStore,
	}
}

func (p *AccessTokenProvider) EnsureToken(ctx context.Context) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Always get fresh access token from store (no caching for access tokens)
	username, accessToken, err := p.configStore.GetCredentialStoreAccessTokens(p.configKey)
	if err != nil {
		return "", fmt.Errorf("get access token from store: %v", err)
	}

	// Cache username for display purposes
	p.cachedUsername = username

	return accessToken, nil
}

func (p *AccessTokenProvider) Username() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.cachedUsername
}

// Helper methods for creating token providers

// NewAccessTokenProviderFromStore creates an AccessTokenProvider from a ConfigStore
// Returns error if no valid access token is available
func NewAccessTokenProviderFromStore(configStore *ConfigStore, configKey string) (*AccessTokenProvider, error) {
	// Test if we can get a valid access token
	_, _, err := configStore.GetCredentialStoreAccessTokens(configKey)
	if err != nil {
		return nil, fmt.Errorf("no valid access token available: %v", err)
	}

	return NewAccessTokenProvider(configStore, configKey), nil
}

// NewLoginTokenProviderFromStore creates a LoginTokenProvider from pull credentials in the ConfigStore
func NewLoginTokenProviderFromStore(configStore *ConfigStore, configKey, baseURL string) (*LoginTokenProvider, error) {
	username, password, err := configStore.GetCredentialStorePullTokens(configKey)
	if err != nil {
		return nil, fmt.Errorf("no pull credentials available: %v", err)
	}

	if username == "" {
		return nil, fmt.Errorf("empty username found in store")
	}

	if password == "" {
		return nil, fmt.Errorf("empty password found in store")
	}

	return NewLoginTokenProvider(username, password, baseURL), nil
}
