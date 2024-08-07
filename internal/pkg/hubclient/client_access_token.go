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

type AccessTokenListParams struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type AccessTokenPage struct {
	Count   int           `json:"count"`
	Results []AccessToken `json:"results"`
}

type AccessTokenDeleteResponse struct {
	Detail  string `json:"detail"`
	Message string `json:"message"`
}

func (c *Client) GetAccessToken(ctx context.Context, accessTokenID string) (AccessToken, error) {
	if !isValidUUID(accessTokenID) {
		return AccessToken{}, fmt.Errorf("accessTokenID is required")
	}
	accessToken := AccessToken{}
	err := c.sendRequest(ctx, "GET", fmt.Sprintf("/access-tokens/%s", accessTokenID), nil, &accessToken)
	return accessToken, err
}

func (c *Client) GetAccessTokens(ctx context.Context, params AccessTokenListParams) (AccessTokenPage, error) {
	accessTokenPage := AccessTokenPage{}
	err := c.sendRequest(ctx, "GET",
		fmt.Sprintf("/access-tokens?page=%d&page_size=%d", params.Page, params.PageSize), nil, &accessTokenPage)
	return accessTokenPage, err
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

func (c *Client) DeleteAccessToken(ctx context.Context, accessTokenID string) (AccessTokenDeleteResponse, error) {
	if !isValidUUID(accessTokenID) {
		return AccessTokenDeleteResponse{}, fmt.Errorf("accessTokenID is required")
	}

	accessTokenDeleteResponse := AccessTokenDeleteResponse{}
	err := c.sendRequest(ctx, "DELETE", fmt.Sprintf("/access-tokens/%s", accessTokenID), nil, &accessTokenDeleteResponse)
	return accessTokenDeleteResponse, err
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
