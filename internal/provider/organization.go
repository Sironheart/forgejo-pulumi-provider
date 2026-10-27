package provider

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v3"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

type Organization struct{}

type OrganizationArgs struct {
	Name                      string  `pulumi:"name"`
	FullName                  string  `pulumi:"fullName,optional"`
	Description               string  `pulumi:"description,optional"`
	Website                   string  `pulumi:"website,optional"`
	Location                  string  `pulumi:"location,optional"`
	Visibility                string  `pulumi:"visibility,optional"`
	Email                     *string `pulumi:"email,optional"`
	RepoAdminChangeTeamAccess *bool   `pulumi:"repoAdminChangeTeamAccess,optional"`
}

type OrganizationState struct {
	OrganizationArgs
	AvatarURL string `pulumi:"avatarUrl"`
}

type organizationCreateOption struct {
	forgejo.CreateOrgOption
	Email                     *string `json:"email,omitempty"`
	RepoAdminChangeTeamAccess *bool   `json:"repo_admin_change_team_access,omitempty"`
}

type organizationEditOption struct {
	forgejo.EditOrgOption
	Email                     *string `json:"email,omitempty"`
	RepoAdminChangeTeamAccess *bool   `json:"repo_admin_change_team_access,omitempty"`
}

type organizationAPIResponse struct {
	forgejo.Organization
	Name                      string `json:"name"`
	Email                     string `json:"email"`
	RepoAdminChangeTeamAccess bool   `json:"repo_admin_change_team_access"`
}

func (o *Organization) Annotate(a infer.Annotator) {
	a.Describe(o, "A Forgejo organization.")
	a.SetToken("index", "Organization")
}

func (a *OrganizationArgs) Annotate(ann infer.Annotator) {
	ann.Describe(&a.Name, "Organization username/slug.")
	ann.Describe(&a.FullName, "Organization display name.")
	ann.Describe(&a.Description, "Organization description.")
	ann.Describe(&a.Website, "Organization website URL.")
	ann.Describe(&a.Location, "Organization location.")
	ann.Describe(&a.Visibility, "Organization visibility, for example public, limited, or private. Leave empty to use Forgejo's default.")
	ann.Describe(&a.Email, "Organization email address. Omit to leave unmanaged; set an empty string to clear it.")
	ann.Describe(&a.RepoAdminChangeTeamAccess, "Whether repository administrators can change team access. Omit to leave unmanaged.")
}

func (Organization) Create(ctx context.Context, req infer.CreateRequest[OrganizationArgs]) (infer.CreateResponse[OrganizationState], error) {
	if req.DryRun {
		return infer.CreateResponse[OrganizationState]{ID: req.Inputs.Name, Output: organizationStateFromArgs(req.Inputs, "")}, nil
	}

	client, err := forgejoAPIClientFromConfig(ctx)
	if err != nil {
		return infer.CreateResponse[OrganizationState]{}, err
	}

	opt := organizationCreateOption{
		CreateOrgOption: forgejo.CreateOrgOption{
			Name:        req.Inputs.Name,
			FullName:    req.Inputs.FullName,
			Description: req.Inputs.Description,
			Website:     req.Inputs.Website,
			Location:    req.Inputs.Location,
			Visibility:  forgejo.VisibleType(req.Inputs.Visibility),
		},
		Email:                     req.Inputs.Email,
		RepoAdminChangeTeamAccess: req.Inputs.RepoAdminChangeTeamAccess,
	}
	if err := opt.Validate(); err != nil {
		return infer.CreateResponse[OrganizationState]{}, err
	}
	org := new(organizationAPIResponse)
	_, err = client.do(http.MethodPost, "/orgs", opt, org)
	if err != nil {
		return infer.CreateResponse[OrganizationState]{}, err
	}

	state := organizationStateFromAPI(req.Inputs, org)
	return infer.CreateResponse[OrganizationState]{ID: state.Name, Output: state}, nil
}

func (Organization) Read(ctx context.Context, req infer.ReadRequest[OrganizationArgs, OrganizationState]) (infer.ReadResponse[OrganizationArgs, OrganizationState], error) {
	name := req.State.Name
	if name == "" {
		name = req.ID
	}
	if name == "" {
		return infer.ReadResponse[OrganizationArgs, OrganizationState]{}, errors.New("organization name is required")
	}

	client, err := forgejoAPIClientFromConfig(ctx)
	if err != nil {
		return infer.ReadResponse[OrganizationArgs, OrganizationState]{}, err
	}

	org := new(organizationAPIResponse)
	resp, err := client.do(http.MethodGet, "/orgs/"+url.PathEscape(name), nil, org)
	if isNotFound(resp) {
		return infer.ReadResponse[OrganizationArgs, OrganizationState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[OrganizationArgs, OrganizationState]{}, err
	}

	state := organizationStateFromAPI(req.State.OrganizationArgs, org)
	return infer.ReadResponse[OrganizationArgs, OrganizationState]{ID: state.Name, Inputs: state.OrganizationArgs, State: state}, nil
}

func (Organization) Diff(_ context.Context, req infer.DiffRequest[OrganizationArgs, OrganizationState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	if req.State.Name != req.Inputs.Name {
		diff["name"] = p.PropertyDiff{Kind: p.UpdateReplace, InputDiff: true}
	}
	addUpdateDiff(diff, "fullName", req.State.FullName != req.Inputs.FullName)
	addUpdateDiff(diff, "description", req.State.Description != req.Inputs.Description)
	addUpdateDiff(diff, "website", req.State.Website != req.Inputs.Website)
	addUpdateDiff(diff, "location", req.State.Location != req.Inputs.Location)
	addUpdateDiff(diff, "visibility", req.Inputs.Visibility != "" && req.State.Visibility != req.Inputs.Visibility)
	addUpdateDiff(diff, "email", req.Inputs.Email != nil && !equalStringPtr(req.State.Email, req.Inputs.Email))
	addUpdateDiff(diff, "repoAdminChangeTeamAccess", req.Inputs.RepoAdminChangeTeamAccess != nil && !equalBoolPtr(req.State.RepoAdminChangeTeamAccess, req.Inputs.RepoAdminChangeTeamAccess))
	return p.DiffResponse{HasChanges: len(diff) > 0, DeleteBeforeReplace: true, DetailedDiff: diff}, nil
}

func (Organization) Update(ctx context.Context, req infer.UpdateRequest[OrganizationArgs, OrganizationState]) (infer.UpdateResponse[OrganizationState], error) {
	if req.DryRun {
		return infer.UpdateResponse[OrganizationState]{Output: organizationStateFromArgs(req.Inputs, req.State.AvatarURL)}, nil
	}

	client, err := forgejoAPIClientFromConfig(ctx)
	if err != nil {
		return infer.UpdateResponse[OrganizationState]{}, err
	}
	name := req.State.Name
	if name == "" {
		name = req.ID
	}
	if name == "" {
		return infer.UpdateResponse[OrganizationState]{}, errors.New("organization name is required")
	}
	visibility := req.Inputs.Visibility
	if visibility == "" {
		visibility = req.State.Visibility
	}

	opt := organizationEditOption{
		EditOrgOption: forgejo.EditOrgOption{
			FullName:    req.Inputs.FullName,
			Description: req.Inputs.Description,
			Website:     req.Inputs.Website,
			Location:    req.Inputs.Location,
			Visibility:  forgejo.VisibleType(visibility),
		},
		Email:                     req.Inputs.Email,
		RepoAdminChangeTeamAccess: req.Inputs.RepoAdminChangeTeamAccess,
	}
	if err := opt.Validate(); err != nil {
		return infer.UpdateResponse[OrganizationState]{}, err
	}
	path := "/orgs/" + url.PathEscape(name)
	_, err = client.do(http.MethodPatch, path, opt, nil)
	if err != nil {
		return infer.UpdateResponse[OrganizationState]{}, err
	}
	org := new(organizationAPIResponse)
	_, err = client.do(http.MethodGet, path, nil, org)
	if err != nil {
		return infer.UpdateResponse[OrganizationState]{}, err
	}

	return infer.UpdateResponse[OrganizationState]{Output: organizationStateFromAPI(req.Inputs, org)}, nil
}

func (Organization) Delete(ctx context.Context, req infer.DeleteRequest[OrganizationState]) (infer.DeleteResponse, error) {
	name := req.State.Name
	if name == "" {
		name = req.ID
	}

	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.DeleteResponse{}, err
	}

	resp, err := client.DeleteOrg(name)
	if isNotFound(resp) {
		return infer.DeleteResponse{}, nil
	}
	return infer.DeleteResponse{}, err
}

func organizationStateFromAPI(args OrganizationArgs, org *organizationAPIResponse) OrganizationState {
	name := org.Name
	if name == "" {
		name = org.UserName
	}
	state := OrganizationState{
		OrganizationArgs: OrganizationArgs{
			Name:        name,
			FullName:    org.FullName,
			Description: org.Description,
			Website:     org.Website,
			Location:    org.Location,
			Visibility:  org.Visibility,
		},
		AvatarURL: org.AvatarURL,
	}
	if args.Email != nil {
		state.Email = &org.Email
	}
	if args.RepoAdminChangeTeamAccess != nil {
		state.RepoAdminChangeTeamAccess = &org.RepoAdminChangeTeamAccess
	}
	return state
}

func organizationStateFromArgs(args OrganizationArgs, avatarURL string) OrganizationState {
	return OrganizationState{OrganizationArgs: args, AvatarURL: avatarURL}
}
