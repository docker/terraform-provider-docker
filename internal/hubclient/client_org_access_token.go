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

const (
	OrgAccessTokenTypeRepo = "TYPE_REPO"
	OrgAccessTokenTypeOrg  = "TYPE_ORG"
)

type OrgAccessToken struct {
	ID          string                   `json:"id"`
	Label       string                   `json:"label"`
	Description string                   `json:"description,omitempty"`
	CreatedBy   string                   `json:"created_by"`
	IsActive    bool                     `json:"is_active"`
	CreatedAt   string                   `json:"created_at"`
	ExpiresAt   string                   `json:"expires_at,omitempty"`
	LastUsedAt  string                   `json:"last_used_at,omitempty"`
	Token       string                   `json:"token,omitempty"`
	Resources   []OrgAccessTokenResource `json:"resources,omitempty"`
}

type OrgAccessTokenPage struct {
	Total    int              `json:"total"`
	Next     string           `json:"next,omitempty"`
	Previous string           `json:"previous,omitempty"`
	Results  []OrgAccessToken `json:"results"`
}

type OrgAccessTokenResource struct {
	Type   string   `json:"type"`
	Path   string   `json:"path"`
	Scopes []string `json:"scopes"`
}

type OrgAccessTokenCreateParams struct {
	Label       string                   `json:"label"`
	Description string                   `json:"description"`
	Resources   []OrgAccessTokenResource `json:"resources"`
	ExpiresAt   string                   `json:"expires_at,omitempty"`
}

type OrgAccessTokenUpdateParams struct {
	Label       string                   `json:"label"`
	Description string                   `json:"description"`
	Resources   []OrgAccessTokenResource `json:"resources"`
}

func (c *Client) CreateOrgAccessToken(ctx context.Context, orgName string, params OrgAccessTokenCreateParams) (OrgAccessToken, error) {
	if orgName == "" {
		return OrgAccessToken{}, fmt.Errorf("orgName is required")
	}

	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(params); err != nil {
		return OrgAccessToken{}, err
	}

	var accessToken OrgAccessToken
	err := c.sendRequest(ctx, "POST", fmt.Sprintf("/orgs/%s/access-tokens", orgName), buf.Bytes(), &accessToken)
	return accessToken, err
}

func (c *Client) ListOrgAccessTokens(ctx context.Context, orgName string) ([]OrgAccessToken, error) {
	if orgName == "" {
		return nil, fmt.Errorf("orgName is required")
	}

	var accessTokens []OrgAccessToken
	nextURL := fmt.Sprintf("/orgs/%s/access-tokens", orgName)

	for nextURL != "" {
		relativeURL := c.convertToRelativeURL(nextURL)

		var page OrgAccessTokenPage
		if err := c.sendRequest(ctx, "GET", relativeURL, nil, &page); err != nil {
			return nil, err
		}

		accessTokens = append(accessTokens, page.Results...)
		nextURL = page.Next
	}

	return accessTokens, nil
}

func (c *Client) GetOrgAccessToken(ctx context.Context, orgName, accessTokenID string) (OrgAccessToken, error) {
	if orgName == "" {
		return OrgAccessToken{}, fmt.Errorf("orgName is required")
	}
	if accessTokenID == "" {
		return OrgAccessToken{}, fmt.Errorf("accessTokenID is required")
	}

	var accessToken OrgAccessToken
	err := c.sendRequest(ctx, "GET", fmt.Sprintf("/orgs/%s/access-tokens/%s", orgName, accessTokenID), nil, &accessToken)
	return accessToken, err
}

func (c *Client) UpdateOrgAccessToken(ctx context.Context, orgName, accessTokenID string, params OrgAccessTokenUpdateParams) (OrgAccessToken, error) {
	if orgName == "" {
		return OrgAccessToken{}, fmt.Errorf("orgName is required")
	}
	if accessTokenID == "" {
		return OrgAccessToken{}, fmt.Errorf("accessTokenID is required")
	}

	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(params); err != nil {
		return OrgAccessToken{}, err
	}

	var accessToken OrgAccessToken
	err := c.sendRequest(ctx, "PATCH", fmt.Sprintf("/orgs/%s/access-tokens/%s", orgName, accessTokenID), buf.Bytes(), &accessToken)
	return accessToken, err
}

func (c *Client) DeleteOrgAccessToken(ctx context.Context, orgName, accessTokenID string) error {
	if orgName == "" {
		return fmt.Errorf("orgName is required")
	}
	if accessTokenID == "" {
		return fmt.Errorf("accessTokenID is required")
	}

	return c.sendRequest(ctx, "DELETE", fmt.Sprintf("/orgs/%s/access-tokens/%s", orgName, accessTokenID), nil, nil)
}
