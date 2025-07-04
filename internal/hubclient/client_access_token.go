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
)

type AccessToken struct {
	UUID        string   `json:"uuid"`
	ClientID    string   `json:"client_id"`
	CreatorIP   string   `json:"creator_ip"`
	CreatorUA   string   `json:"creator_ua"`
	CreatedAt   string   `json:"created_at"`
	LastUsed    string   `json:"last_used"`
	GeneratedBy string   `json:"generated_by"`
	IsActive    bool     `json:"is_active"`
	Token       string   `json:"token"`
	TokenLabel  string   `json:"token_label"`
	Scopes      []string `json:"scopes"`
}

type AccessTokenCreateParams struct {
	TokenLabel string   `json:"token_label"`
	Scopes     []string `json:"scopes"`
}

type AccessTokenUpdateParams struct {
	TokenLabel string `json:"token_label"`
	IsActive   bool   `json:"is_active"`
}

type AccessTokenPage struct {
	Count    int           `json:"count"`
	Next     interface{}   `json:"next,omitempty"`
	Previous interface{}   `json:"previous,omitempty"`
	Results  []AccessToken `json:"results"`
}

func (c *Client) GetAccessToken(ctx context.Context, accessTokenID string) (AccessToken, error) {
	if !isValidUUID(accessTokenID) {
		return AccessToken{}, fmt.Errorf("accessTokenID is required")
	}
	accessToken := AccessToken{}
	err := c.sendRequest(ctx, "GET", fmt.Sprintf("/access-tokens/%s", accessTokenID), nil, &accessToken)
	return accessToken, err
}

func (c *Client) GetAccessTokens(ctx context.Context) (AccessTokenPage, error) {
	var allTokens []AccessToken
	initialURL := "/access-tokens"

	err := c.paginate(ctx, initialURL, func(url string) (interface{}, error) {
		var page AccessTokenPage
		if err := c.sendRequest(ctx, "GET", url, nil, &page); err != nil {
			return nil, err
		}

		allTokens = append(allTokens, page.Results...)
		return page.Next, nil
	})
	if err != nil {
		return AccessTokenPage{}, err
	}

	return AccessTokenPage{
		Count:   len(allTokens),
		Results: allTokens,
	}, nil
}

func (c *Client) UpdateAccessToken(ctx context.Context, accessTokenID string, accessTokenUpdate AccessTokenUpdateParams) (AccessToken, error) {
	if !isValidUUID(accessTokenID) {
		return AccessToken{}, fmt.Errorf("accessTokenID is required")
	}
	accessTokenUpdated := AccessToken{}
	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(accessTokenUpdate); err != nil {
		return AccessToken{}, err
	}

	err := c.sendRequest(ctx, "PATCH", fmt.Sprintf("/access-tokens/%s", accessTokenID), buf.Bytes(), &accessTokenUpdated)
	return accessTokenUpdated, err
}

func (c *Client) DeleteAccessToken(ctx context.Context, accessTokenID string) error {
	if !isValidUUID(accessTokenID) {
		return fmt.Errorf("accessTokenID is required")
	}

	return c.sendRequest(ctx, "DELETE", fmt.Sprintf("/access-tokens/%s", accessTokenID), nil, nil)
}

func (c *Client) CreateAccessToken(ctx context.Context, accessToken AccessTokenCreateParams) (AccessToken, error) {
	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(accessToken); err != nil {
		return AccessToken{}, err
	}

	accessTokenCreated := AccessToken{}
	err := c.sendRequest(ctx, "POST", "/access-tokens", buf.Bytes(), &accessTokenCreated)
	return accessTokenCreated, err
}

// TODO(nicks): better validation.
func isValidUUID(uuid string) bool {
	return uuid != ""
}
