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
	"context"
	"encoding/json"
	"fmt"
)

const ImmutableTagRulesSeparator = ","

type ImmutableTagsSettings struct {
	Enabled bool     `json:"enabled"`
	Rules   []string `json:"rules"`
}

type Repository struct {
	Name                  string                `json:"name"`
	Namespace             string                `json:"namespace"`
	RepositoryType        string                `json:"repository_type,omitempty"`
	IsPrivate             bool                  `json:"is_private"`
	Status                int                   `json:"status"`
	StatusDescription     string                `json:"status_description"`
	Description           string                `json:"description"`
	StarCount             int64                 `json:"star_count"`
	PullCount             int64                 `json:"pull_count"`
	LastUpdated           string                `json:"last_updated"`
	DateRegistered        string                `json:"date_registered"`
	Affiliation           string                `json:"affiliation"`
	MediaTypes            []string              `json:"media_types,omitempty"`
	ContentTypes          []string              `json:"content_types,omitempty"`
	User                  string                `json:"use"`
	IsAutomated           bool                  `json:"is_automated"`
	CollaboratorCount     int64                 `json:"collaborator_count"`
	HubUser               string                `json:"hub_user"`
	HasStarred            bool                  `json:"has_starred"`
	FullDescription       string                `json:"full_description"`
	Permissions           Permissions           `json:"permissions"`
	ImmutableTagsSettings ImmutableTagsSettings `json:"immutable_tags_settings"`
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

type Tag struct {
	Name            string     `json:"name"`
	FullSize        int64      `json:"full_size"`
	ID              int64      `json:"id"`
	Repository      int64      `json:"repository"`
	Creator         int64      `json:"creator"`
	LastUpdated     string     `json:"last_updated"`
	LastUpdater     int64      `json:"last_updater"`
	LastUpdaterName string     `json:"last_updater_username"`
	ImageID         string     `json:"image_id"`
	V2              bool       `json:"v2"`
	TagStatus       string     `json:"tag_status"`
	TagLastPulled   string     `json:"tag_last_pulled"`
	TagLastPushed   string     `json:"tag_last_pushed"`
	MediaType       string     `json:"media_type"`
	ContentType     string     `json:"content_type"`
	Digest          string     `json:"digest"`
	Images          []TagImage `json:"images"`
}

type TagImage struct {
	Architecture string `json:"architecture"`
	Features     string `json:"features"`
	Variant      string `json:"variant"`
	Digest       string `json:"digest"`
	OS           string `json:"os"`
	OSFeatures   string `json:"os_features"`
	OSVersion    string `json:"os_version"`
	Size         int64  `json:"size"`
	Status       string `json:"status"`
	LastPulled   string `json:"last_pulled"`
	LastPushed   string `json:"last_pushed"`
}

type Tags struct {
	Count    int         `json:"count"`
	Next     interface{} `json:"next,omitempty"`
	Previous interface{} `json:"previous,omitempty"`
	Results  []Tag       `json:"results"`
}

type CreateRepositoryRequest struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	FullDescription string `json:"full_description"`
	Registry        string `json:"registry"`
	IsPrivate       bool   `json:"is_private"`
}

func (c *Client) CreateRepository(ctx context.Context, namespace string, req CreateRepositoryRequest) (Repository, error) {
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
	Description        string `json:"description"`
	FullDescription    string `json:"full_description"`
	ImmutableTags      bool   `json:"immutable_tags"`
	ImmutableTagsRules string `json:"immutable_tags_rules,omitempty"`
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

func (c *Client) GetRepositoryTags(ctx context.Context, namespace, name string) (*Tags, error) {
	var allTags []Tag
	initialURL := fmt.Sprintf("/namespaces/%s/repositories/%s/tags", namespace, name)

	err := c.paginate(ctx, initialURL, func(url string) (interface{}, error) {
		var page Tags
		if err := c.sendRequest(ctx, "GET", url, nil, &page); err != nil {
			return nil, err
		}

		allTags = append(allTags, page.Results...)
		return page.Next, nil
	})
	if err != nil {
		return nil, err
	}

	return &Tags{
		Count:   len(allTags),
		Results: allTags,
	}, nil
}

func (c *Client) GetRepositoryTag(ctx context.Context, namespace string, repository string, tag string) (Tag, error) {
	tagInfo := Tag{}
	url := fmt.Sprintf("/repositories/%s/%s/tags/%s", namespace, repository, tag)
	err := c.sendRequest(ctx, "GET", url, nil, &tagInfo)
	return tagInfo, err
}

func (c *Client) GetRepositories(ctx context.Context, namespace string) (Repositories, error) {
	var allRepos []Repository
	initialURL := fmt.Sprintf("/repositories/%s/", namespace)

	err := c.paginate(ctx, initialURL, func(url string) (interface{}, error) {
		var page Repositories
		if err := c.sendRequest(ctx, "GET", url, nil, &page); err != nil {
			return nil, err
		}

		allRepos = append(allRepos, page.Results...)
		return page.Next, nil
	})
	if err != nil {
		return Repositories{}, err
	}

	return Repositories{
		Count:   len(allRepos),
		Results: allRepos,
	}, nil
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
