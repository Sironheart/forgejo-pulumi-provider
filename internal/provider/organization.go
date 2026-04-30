package provider

import (
	"context"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v3"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

type Organization struct{}

type OrganizationArgs struct {
	Name        string `pulumi:"name"`
	FullName    string `pulumi:"fullName,optional"`
	Description string `pulumi:"description,optional"`
	Website     string `pulumi:"website,optional"`
	Location    string `pulumi:"location,optional"`
	Visibility  string `pulumi:"visibility,optional"`
}

type OrganizationState struct {
	OrganizationArgs
	AvatarURL string `pulumi:"avatarUrl"`
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
}

func (Organization) Create(ctx context.Context, req infer.CreateRequest[OrganizationArgs]) (infer.CreateResponse[OrganizationState], error) {
	if req.DryRun {
		return infer.CreateResponse[OrganizationState]{ID: req.Inputs.Name, Output: organizationStateFromArgs(req.Inputs, "")}, nil
	}

	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.CreateResponse[OrganizationState]{}, err
	}

	org, _, err := client.CreateOrg(forgejo.CreateOrgOption{
		Name:        req.Inputs.Name,
		FullName:    req.Inputs.FullName,
		Description: req.Inputs.Description,
		Website:     req.Inputs.Website,
		Location:    req.Inputs.Location,
		Visibility:  forgejo.VisibleType(req.Inputs.Visibility),
	})
	if err != nil {
		return infer.CreateResponse[OrganizationState]{}, err
	}

	state := organizationStateFromAPI(org)
	return infer.CreateResponse[OrganizationState]{ID: state.Name, Output: state}, nil
}

func (Organization) Read(ctx context.Context, req infer.ReadRequest[OrganizationArgs, OrganizationState]) (infer.ReadResponse[OrganizationArgs, OrganizationState], error) {
	name := req.State.Name
	if name == "" {
		name = req.ID
	}

	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.ReadResponse[OrganizationArgs, OrganizationState]{}, err
	}

	org, resp, err := client.GetOrg(name)
	if isNotFound(resp) {
		return infer.ReadResponse[OrganizationArgs, OrganizationState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[OrganizationArgs, OrganizationState]{}, err
	}

	state := organizationStateFromAPI(org)
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
	return p.DiffResponse{HasChanges: len(diff) > 0, DeleteBeforeReplace: true, DetailedDiff: diff}, nil
}

func (Organization) Update(ctx context.Context, req infer.UpdateRequest[OrganizationArgs, OrganizationState]) (infer.UpdateResponse[OrganizationState], error) {
	if req.DryRun {
		return infer.UpdateResponse[OrganizationState]{Output: organizationStateFromArgs(req.Inputs, req.State.AvatarURL)}, nil
	}

	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.UpdateResponse[OrganizationState]{}, err
	}
	name := req.State.Name
	if name == "" {
		name = req.ID
	}
	visibility := req.Inputs.Visibility
	if visibility == "" {
		visibility = req.State.Visibility
	}

	resp, err := client.EditOrg(name, forgejo.EditOrgOption{
		FullName:    req.Inputs.FullName,
		Description: req.Inputs.Description,
		Website:     req.Inputs.Website,
		Location:    req.Inputs.Location,
		Visibility:  forgejo.VisibleType(visibility),
	})
	if isNotFound(resp) {
		return infer.UpdateResponse[OrganizationState]{}, nil
	}
	if err != nil {
		return infer.UpdateResponse[OrganizationState]{}, err
	}
	org, _, err := client.GetOrg(name)
	if err != nil {
		return infer.UpdateResponse[OrganizationState]{}, err
	}

	return infer.UpdateResponse[OrganizationState]{Output: organizationStateFromAPI(org)}, nil
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

func organizationStateFromAPI(org *forgejo.Organization) OrganizationState {
	return OrganizationState{
		OrganizationArgs: OrganizationArgs{
			Name:        org.UserName,
			FullName:    org.FullName,
			Description: org.Description,
			Website:     org.Website,
			Location:    org.Location,
			Visibility:  org.Visibility,
		},
		AvatarURL: org.AvatarURL,
	}
}

func organizationStateFromArgs(args OrganizationArgs, avatarURL string) OrganizationState {
	return OrganizationState{OrganizationArgs: args, AvatarURL: avatarURL}
}
