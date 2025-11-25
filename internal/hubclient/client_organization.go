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

type Org struct {
	ID         string `json:"id,omitempty"`
	OrgName    string `json:"orgname"`
	FullName   string `json:"full_name"`
	Location   string `json:"location"`
	Company    string `json:"company"`
	DateJoined string `json:"date_joined"`
}

type OrgSettings struct {
	OrgName     string `json:"id,omitempty"`
	TeamName    string `json:"team_name,omitempty"`
	Permissions string `json:"restricted_images"`
}

type OrgTeam struct {
	ID          int64  `json:"id"`
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MemberCount int    `json:"member_count"`
}

type OrgTeamMember struct {
	ID              string   `json:"id,omitempty"`
	OrgName         string   `json:"org_name,omitempty"`
	TeamName        string   `json:"name"`
	TeamDescription string   `json:"description"`
	UUID            string   `json:"uuid"`
	Username        string   `json:"username"`
	FullName        string   `json:"full_name"`
	Location        string   `json:"location"`
	Company         string   `json:"company"`
	ProfileURL      string   `json:"profile_url"`
	DateJoined      string   `json:"date_joined"`
	GravatarURL     string   `json:"gravatar_url"`
	GravatarEmail   string   `json:"gravatar_email"`
	Type            string   `json:"type"`
	Email           string   `json:"email"`
	Role            string   `json:"role"`
	Groups          []string `json:"groups"`
	IsGuest         bool     `json:"is_guest"`
	PrimaryEmail    string   `json:"primary_email"`
}

type OrgTeamMembersResponse struct {
	Count    int             `json:"count"`
	Next     string          `json:"next"`
	Previous string          `json:"previous"`
	Results  []OrgTeamMember `json:"results"`
}

type OrgTeamMemberRequest struct {
	Member string `json:"member"`
}

type OrgMemberRequest struct {
	Org      string   `json:"org"`
	Team     string   `json:"team"`
	Invitees []string `json:"invitees"`
	Role     string   `json:"role"`
	DryRun   bool     `json:"dry_run"`
}

type OrgInviteResponse struct {
	OrgInvitees []OrgInvitee `json:"invitees"`
}

type OrgInvitee struct {
	Invitee string    `json:"invitee"`
	Status  string    `json:"status"`
	Invite  OrgInvite `json:"invite"`
}

type OrgInvite struct {
	ID              string `json:"id"`
	InviterUsername string `json:"inviter_username"`
	Invitee         string `json:"invitee"`
	Team            string `json:"team"`
	Org             string `json:"org"`
	Role            string `json:"role"`
	CreatedAt       string `json:"created_at"`
}

type OrgInvitesListResponse struct {
	Data []OrgInvite `json:"data"`
}

type OrgSettingImageAccessManagement struct {
	RestrictedImages ImageAccessManagementRestrictedImages `json:"restricted_images"`
}

type ImageAccessManagementRestrictedImages struct {
	Enabled                 bool `json:"enabled"`
	AllowOfficialImages     bool `json:"allow_official_images"`
	AllowVerifiedPublishers bool `json:"allow_verified_publishers"`
}

type OrgSettingRegistryAccessManagement struct {
	Enabled            bool                                       `json:"enabled"`
	StandardRegistries []RegistryAccessManagementStandardRegistry `json:"standard_registries"`
	CustomRegistries   []RegistryAccessManagementCustomRegistry   `json:"custom_registries"`
}

// OrgMemberList response
type OrgMemberListResponse struct {
	Count    int         `json:"count"`    // The total number of items that match with the search.
	Previous string      `json:"previous"` // The URL or link for the previous page of items.
	Next     string      `json:"next"`     // The URL or link for the next page of items.
	Results  []OrgMember `json:"results"`  // List of accounts.
}

// Weirdly, org role params are lowercase even though org role return values
// are uppercase.
type OrgRoleParam string

const (
	OrgRoleParamOwner  OrgRoleParam = "owner"
	OrgRoleParamEditor OrgRoleParam = "editor"
	OrgRoleParamMember OrgRoleParam = "member"
)

// OrgMember represents each organization member
type OrgMember struct {
	Email         string   `json:"email"`          // User's email address
	Role          string   `json:"role"`           // Enum: "Owner", "Member", "Invitee" - User's role in the Organization
	Groups        []string `json:"groups"`         // Groups (Teams) that the user is member of
	IsGuest       bool     `json:"is_guest"`       // If the organization has verified domains
	ID            string   `json:"id"`             // The UUID trimmed
	Company       string   `json:"company"`        // Company name
	DateJoined    string   `json:"date_joined"`    // Date joined
	FullName      string   `json:"full_name"`      // Full name
	GravatarEmail string   `json:"gravatar_email"` // Gravatar email
	GravatarURL   string   `json:"gravatar_url"`   // Gravatar URL
	Location      string   `json:"location"`       // Location
	ProfileURL    string   `json:"profile_url"`    // Profile URL
	Type          string   `json:"type"`           // Enum: "User", "Org"
	Username      string   `json:"username"`       // Username
}

const (
	StandardRegistryDocker = "DockerHub"
)

type RegistryAccessManagementStandardRegistry struct {
	ID      string `json:"id"`
	Allowed bool   `json:"allowed"`
}

type RegistryAccessManagementCustomRegistry struct {
	Address      string `json:"address"`
	FriendlyName string `json:"friendly_name"`
	Allowed      bool   `json:"allowed"`
}

func (c *Client) GetOrg(ctx context.Context, orgName string) (Org, error) {
	org := Org{}
	err := c.sendRequest(ctx, "GET", fmt.Sprintf("/orgs/%s/", orgName), nil, &org)
	return org, err
}

func (c *Client) ListOrgMembers(ctx context.Context, orgName string) ([]OrgMember, error) {
	var members []OrgMember
	initialURL := fmt.Sprintf("/orgs/%s/members", orgName)
	err := c.paginate(ctx, initialURL, func(url string) (interface{}, error) {
		var page OrgMemberListResponse
		if err := c.sendRequest(ctx, "GET", url, nil, &page); err != nil {
			return nil, err
		}

		members = append(members, page.Results...)
		return page.Next, nil
	})
	if err != nil {
		return nil, err
	}
	return members, nil
}

func (c *Client) GetOrgSettings(ctx context.Context, orgName string) (Org, error) {
	org := Org{}
	err := c.sendRequest(ctx, "GET", fmt.Sprintf("/orgs/%s/settings/", orgName), nil, &org)
	return org, err
}

func (c *Client) GetOrgTeam(ctx context.Context, orgName string, teamName string) (OrgTeam, error) {
	orgTeam := OrgTeam{}
	err := c.sendRequest(ctx, "GET", fmt.Sprintf("/orgs/%s/groups/%s/", orgName, teamName), nil, &orgTeam)
	return orgTeam, err
}

func (c *Client) CreateOrgTeam(ctx context.Context, orgName string, createOrgTeam OrgTeam) (OrgTeam, error) {
	orgTeam := OrgTeam{}
	createOrgTeamJSON, err := json.Marshal(createOrgTeam)
	if err != nil {
		return orgTeam, err
	}
	err = c.sendRequest(ctx, "POST", fmt.Sprintf("/orgs/%s/groups/", orgName), createOrgTeamJSON, &orgTeam)
	if err != nil {
		return orgTeam, err
	}
	return orgTeam, err
}

func (c *Client) UpdateOrgTeam(ctx context.Context, orgName string, teamName string, updateOrgTeam OrgTeam) (OrgTeam, error) {
	orgTeam := OrgTeam{}
	updateOrgTeamJSON, err := json.Marshal(updateOrgTeam)
	if err != nil {
		return orgTeam, err
	}
	err = c.sendRequest(ctx, "PATCH", fmt.Sprintf("/orgs/%s/groups/%s/", orgName, teamName), updateOrgTeamJSON, &orgTeam)
	return orgTeam, err
}

func (c *Client) DeleteOrgTeam(ctx context.Context, orgName string, teamName string) error {
	return c.sendRequest(ctx, "DELETE", fmt.Sprintf("/orgs/%s/groups/%s/", orgName, teamName), nil, nil)
}

func (c *Client) DeleteOrgTeamMember(ctx context.Context, orgName string, teamName string, userName string) error {
	return c.sendRequest(ctx, "DELETE", fmt.Sprintf("/orgs/%s/groups/%s/members/%s", orgName, teamName, userName), nil, nil)
}

func (c *Client) AddOrgTeamMember(ctx context.Context, orgName string, teamName string, userName string) error {
	memberRequest := OrgTeamMemberRequest{
		Member: userName,
	}
	memberRequestJSON, err := json.Marshal(memberRequest)
	if err != nil {
		return err
	}
	return c.sendRequest(ctx, "POST", fmt.Sprintf("/orgs/%s/groups/%s/members/", orgName, teamName), memberRequestJSON, nil)
}

func (c *Client) ListOrgTeamMembers(ctx context.Context, orgName string, teamName string) ([]OrgTeamMember, error) {
	var members []OrgTeamMember
	initialURL := fmt.Sprintf("/orgs/%s/groups/%s/members/", orgName, teamName)
	err := c.paginate(ctx, initialURL, func(url string) (interface{}, error) {
		var page OrgTeamMembersResponse
		if err := c.sendRequest(ctx, "GET", url, nil, &page); err != nil {
			return nil, err
		}

		members = append(members, page.Results...)
		return page.Next, nil
	})
	if err != nil {
		return nil, err
	}
	return members, nil
}

func (c *Client) GetOrgSettingImageAccessManagement(ctx context.Context, orgName string) (OrgSettingImageAccessManagement, error) {
	var settings OrgSettingImageAccessManagement
	err := c.sendRequest(ctx, "GET", fmt.Sprintf("/orgs/%s/settings/", orgName), nil, &settings)
	return settings, err
}

func (c *Client) SetOrgSettingImageAccessManagement(ctx context.Context, orgName string, iamSettings OrgSettingImageAccessManagement) (OrgSettingImageAccessManagement, error) {
	reqBody, err := json.Marshal(iamSettings)
	if err != nil {
		return OrgSettingImageAccessManagement{}, err
	}
	err = c.sendRequest(ctx, "PUT", fmt.Sprintf("/orgs/%s/settings", orgName), reqBody, nil)
	if err != nil {
		return OrgSettingImageAccessManagement{}, err
	}
	return c.GetOrgSettingImageAccessManagement(ctx, orgName)
}

func (c *Client) GetOrgSettingRegistryAccessManagement(ctx context.Context, orgName string) (OrgSettingRegistryAccessManagement, error) {
	var settings OrgSettingRegistryAccessManagement
	err := c.sendRequest(ctx, "GET", fmt.Sprintf("/orgs/%s/settings/registry-access-management", orgName), nil, &settings)
	return settings, err
}

func (c *Client) SetOrgSettingRegistryAccessManagement(ctx context.Context, orgName string, reamSettings OrgSettingRegistryAccessManagement) (OrgSettingRegistryAccessManagement, error) {
	reqBody, err := json.Marshal(reamSettings)
	if err != nil {
		return OrgSettingRegistryAccessManagement{}, err
	}
	err = c.sendRequest(ctx, "PUT", fmt.Sprintf("/orgs/%s/settings/registry-access-management", orgName), reqBody, nil)
	if err != nil {
		return OrgSettingRegistryAccessManagement{}, err
	}
	return c.GetOrgSettingRegistryAccessManagement(ctx, orgName)
}

func (c *Client) ListOrgInvites(ctx context.Context, orgName string) ([]OrgInvite, error) {
	var invites OrgInvitesListResponse
	err := c.sendRequest(ctx, "GET", fmt.Sprintf("/orgs/%s/invites", orgName), nil, &invites)
	return invites.Data, err
}

func (c *Client) InviteOrgMember(ctx context.Context, orgName, role string, invitees []string, dryRun bool) (OrgInviteResponse, error) {
	inviteRequest := OrgMemberRequest{
		Org:      orgName,
		Invitees: invitees,
		Role:     role,
		DryRun:   dryRun,
	}
	reqBody, err := json.Marshal(inviteRequest)
	if err != nil {
		return OrgInviteResponse{}, err
	}

	var inviteResponse OrgInviteResponse
	err = c.sendRequest(ctx, "POST", "/invites/bulk", reqBody, &inviteResponse)
	return inviteResponse, err
}

func (c *Client) DeleteOrgInvite(ctx context.Context, inviteID string) error {
	url := fmt.Sprintf("/invites/%s", inviteID)
	return c.sendRequest(ctx, "DELETE", url, nil, nil)
}

func (c *Client) DeleteOrgMember(ctx context.Context, orgName string, userName string) error {
	url := fmt.Sprintf("/orgs/%s/members/%s/", orgName, userName)
	return c.sendRequest(ctx, "DELETE", url, nil, nil)
}

func (c *Client) UpdateOrgMember(ctx context.Context, orgName string, userName string, role OrgRoleParam) error {
	url := fmt.Sprintf("/orgs/%s/members/%s/", orgName, userName)
	body := map[string]string{"role": string(role)}
	reqBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	return c.sendRequest(ctx, "PUT", url, reqBody, nil)
}

// func (c *Client) GetOrgInvitedMember(ctx context.Context, inviteID string) (OrgMembersResponse, error) {
// 	url := fmt.Sprintf("/invites", inviteID)
// 	var membersResponse OrgMembersResponse
// 	err := c.sendRequest(ctx, "GET", url, nil, &membersResponse)
// 	return membersResponse, err
// }

// func (c *Client) GetOrgMembers(ctx context.Context, orgName string) (OrgMembersResponse, error) {
// 	url := fmt.Sprintf("/orgs/%s/members", orgName)
// 	var membersResponse OrgMembersResponse
// 	err := c.sendRequest(ctx, "GET", url, nil, &membersResponse)
// 	return membersResponse, err
// }
