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
	ID          int    `json:"id"`
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

type OrgMembersResponse struct {
	Count    int             `json:"count"`
	Next     interface{}     `json:"next"`
	Previous interface{}     `json:"previous"`
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

func (c *Client) ListOrgTeamMembers(ctx context.Context, orgName string, teamName string) (OrgMembersResponse, error) {
	membersResponse := OrgMembersResponse{}
	err := c.sendRequest(ctx, "GET", fmt.Sprintf("/orgs/%s/groups/%s/members/", orgName, teamName), nil, &membersResponse)
	return membersResponse, err
}

// // TODO: This is returning a 503 for some reason... moving on as to no stay blocked
// func (c *Client) ListOrgTeamMembers(ctx context.Context, orgName string, teamName string) ([]string, error) {
// 	var members []string
// 	err := c.sendRequest(ctx, "GET", fmt.Sprintf("/orgs/%s/teams/%s/members", orgName, teamName), nil, &members)
// 	return members, err
// }

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

func (c *Client) InviteOrgMember(ctx context.Context, orgName, teamName, role string, invitees []string, dryRun bool) (OrgInviteResponse, error) {
	inviteRequest := OrgMemberRequest{
		Org:      orgName,
		Team:     teamName,
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
