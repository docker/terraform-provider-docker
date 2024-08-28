package hubclient

import (
	"context"
	"encoding/json"
	"fmt"
)

type Repository struct {
	Name              string      `json:"name"`
	Namespace         string      `json:"namespace"`
	RepositoryType    string      `json:"repository_type,omitempty"`
	IsPrivate         bool        `json:"is_private"`
	Status            int         `json:"status"`
	StatusDescription string      `json:"status_description"`
	Description       string      `json:"description"`
	StarCount         int64       `json:"star_count"`
	PullCount         int64       `json:"pull_count"`
	LastUpdated       string      `json:"last_updated"`
	DateRegistered    string      `json:"date_registered"`
	Affiliation       string      `json:"affiliation"`
	MediaTypes        []string    `json:"media_types,omitempty"`
	ContentTypes      []string    `json:"content_types,omitempty"`
	User              string      `json:"use"`
	IsAutomated       bool        `json:"is_automated"`
	CollaboratorCount int64       `json:"collaborator_count"`
	HubUser           string      `json:"hub_user"`
	HasStarred        bool        `json:"has_starred"`
	FullDescription   string      `json:"full_description"`
	Permissions       Permissions `json:"permissions"`
}

type Permissions struct {
	Read  bool `json:"read"`
	Write bool `json:"write"`
	Admin bool `json:"admin"`
}

type Repositories struct {
	Count    int          `json:"count"`
	Next     interface{}  `json:"next,omitempty"`
	Previous interface{}  `json:"previous,omitempty"`
	Results  []Repository `json:"results"`
}

type TeamRepoPermissionLevel string

const (
	TeamRepoPermissionLevelRead  = "read"
	TeamRepoPermissionLevelWrite = "write"
	TeamRepoPermissionLevelAdmin = "admin"
)

type TeamRepoPermission struct {
	TeamID     int64  `json:"group_id"`
	TeamName   string `json:"group_name"`
	Permission string `json:"permission"`
}

type CreateRepostoryRequest struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	FullDescription string `json:"full_description"`
	Registry        string `json:"registry"`
	IsPrivate       bool   `json:"is_private"`
}

func (c *Client) CreateRepository(ctx context.Context, namespace string, req CreateRepostoryRequest) (Repository, error) {
	repository := Repository{}
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return repository, err
	}

	url := fmt.Sprintf("/namespaces/%s/repositories", namespace)
	err = c.sendRequest(ctx, "POST", url, reqJSON, &repository)
	return repository, err
}

type UpdateRepositoryRequest struct {
	Description     string `json:"description"`
	FullDescription string `json:"full_description"`
	Status          int    `json:"status"`
}

func (c *Client) UpdateRepository(ctx context.Context, id string, req UpdateRepositoryRequest) (Repository, error) {
	repository := Repository{}
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return repository, err
	}

	err = c.sendRequest(ctx, "PATCH", fmt.Sprintf("/repositories/%s/", id), reqJSON, &repository)
	return repository, err
}

type SetRepositoryPrivacyRequest struct {
	IsPrivate bool `json:"is_private"`
}

func (c *Client) SetRepositoryPrivacy(ctx context.Context, id string, isPrivate bool) error {
	req := SetRepositoryPrivacyRequest{
		IsPrivate: isPrivate,
	}
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("error marshalling request: %w", err)
	}

	return c.sendRequest(ctx, "POST", fmt.Sprintf("/repositories/%s/privacy", id), reqJSON, nil)
}

func (c *Client) GetRepository(ctx context.Context, id string) (Repository, error) {
	repository := Repository{}
	err := c.sendRequest(ctx, "GET", fmt.Sprintf("/repositories/%s/", id), nil, &repository)
	return repository, err
}

func (c *Client) DeleteRepository(ctx context.Context, id string) error {
	return c.sendRequest(ctx, "DELETE", fmt.Sprintf("/repositories/%s/", id), nil, nil)
}

func (c *Client) GetRepositories(ctx context.Context, namespace string, maxResults int) (Repositories, error) {
	repositories := Repositories{} // Initialize an empty Repositories struct
	err := c.sendRequest(ctx, "GET", fmt.Sprintf("/repositories/%s/?page_size=%d", namespace, maxResults), nil, &repositories)
	return repositories, err
}

func (c *Client) CreatePermissionForTeamAndRepo(ctx context.Context, repository string, teamID int64, permission string) (TeamRepoPermission, error) {
	created := TeamRepoPermission{}
	body, err := json.Marshal(&TeamRepoPermission{
		Permission: permission,
		TeamID:     teamID,
	})
	if err != nil {
		return created, err
	}
	err = c.sendRequest(ctx, "POST", fmt.Sprintf("/repositories/%s/groups/", repository), body, &created)
	if err != nil {
		return created, err
	}
	return created, err
}

func (c *Client) GetPermissionForTeamAndRepo(ctx context.Context, repository string, teamID int64) (TeamRepoPermission, error) {
	perm := TeamRepoPermission{}
	err := c.sendRequest(ctx, "GET", fmt.Sprintf("/repositories/%s/groups/%v/", repository, teamID), nil, &perm)
	return perm, err
}

func (c *Client) UpdatePermissionForTeamAndRepo(ctx context.Context, repository string, teamID int64, permission string) (TeamRepoPermission, error) {
	updated := TeamRepoPermission{}
	body, err := json.Marshal(&TeamRepoPermission{
		Permission: permission,
	})
	if err != nil {
		return updated, err
	}
	err = c.sendRequest(ctx, "PATCH", fmt.Sprintf("/repositories/%s/groups/%v/", repository, teamID), body, &updated)
	if err != nil {
		return updated, err
	}
	return updated, err
}

func (c *Client) DeletePermissionForTeamAndRepo(ctx context.Context, repository string, teamID int64) error {
	return c.sendRequest(ctx, "DELETE", fmt.Sprintf("/repositories/%s/groups/%v/", repository, teamID), nil, nil)
}
