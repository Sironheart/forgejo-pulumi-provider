package provider

import (
	"context"
	"fmt"
	"maps"
	"strconv"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v3"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

type OrganizationTeam struct{}

type OrganizationTeamArgs struct {
	Organization            string            `pulumi:"organization"`
	Name                    string            `pulumi:"name"`
	Description             string            `pulumi:"description,optional"`
	Permission              string            `pulumi:"permission,optional"`
	CanCreateOrgRepo        bool              `pulumi:"canCreateOrgRepo,optional"`
	IncludesAllRepositories bool              `pulumi:"includesAllRepositories,optional"`
	UnitsMap                map[string]string `pulumi:"unitsMap,optional"`
}

type OrganizationTeamState struct {
	OrganizationTeamArgs
	TeamID int64 `pulumi:"teamId"`
}

func (t *OrganizationTeam) Annotate(a infer.Annotator) {
	a.Describe(t, "A Forgejo organization team.")
	a.SetToken("index", "OrganizationTeam")
}

func (a *OrganizationTeamArgs) Annotate(ann infer.Annotator) {
	ann.Describe(&a.Organization, "Organization username/slug.")
	ann.Describe(&a.Name, "Team name.")
	ann.Describe(&a.Description, "Team description.")
	ann.Describe(&a.Permission, "Team repository permission: read, write, or admin.")
	ann.Describe(&a.CanCreateOrgRepo, "Whether team members can create organization repositories.")
	ann.Describe(&a.IncludesAllRepositories, "Whether the team includes all organization repositories.")
	ann.Describe(&a.UnitsMap, "Per-repository-unit permissions, for example repo.code=read or repo.issues=write.")
	ann.SetDefault(&a.Permission, "read")
}

func (OrganizationTeam) Create(ctx context.Context, req infer.CreateRequest[OrganizationTeamArgs]) (infer.CreateResponse[OrganizationTeamState], error) {
	if req.DryRun {
		return infer.CreateResponse[OrganizationTeamState]{ID: organizationTeamPreviewID(req.Inputs), Output: organizationTeamStateFromArgs(req.Inputs, 0)}, nil
	}

	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.CreateResponse[OrganizationTeamState]{}, err
	}

	team, _, err := client.CreateTeam(req.Inputs.Organization, forgejo.CreateTeamOption{
		Name:                    req.Inputs.Name,
		Description:             req.Inputs.Description,
		Permission:              forgejo.AccessMode(req.Inputs.Permission),
		CanCreateOrgRepo:        req.Inputs.CanCreateOrgRepo,
		IncludesAllRepositories: req.Inputs.IncludesAllRepositories,
		UnitsMap:                req.Inputs.UnitsMap,
	})
	if err != nil {
		return infer.CreateResponse[OrganizationTeamState]{}, err
	}

	state := organizationTeamStateFromAPI(team, req.Inputs.Organization)
	return infer.CreateResponse[OrganizationTeamState]{ID: strconv.FormatInt(team.ID, 10), Output: state}, nil
}

func (OrganizationTeam) Read(ctx context.Context, req infer.ReadRequest[OrganizationTeamArgs, OrganizationTeamState]) (infer.ReadResponse[OrganizationTeamArgs, OrganizationTeamState], error) {
	teamID, err := organizationTeamID(req.ID, req.State)
	if err != nil {
		return infer.ReadResponse[OrganizationTeamArgs, OrganizationTeamState]{}, err
	}

	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.ReadResponse[OrganizationTeamArgs, OrganizationTeamState]{}, err
	}

	team, resp, err := client.GetTeam(teamID)
	if isNotFound(resp) {
		return infer.ReadResponse[OrganizationTeamArgs, OrganizationTeamState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[OrganizationTeamArgs, OrganizationTeamState]{}, err
	}

	state := organizationTeamStateFromAPI(team, req.State.Organization)
	return infer.ReadResponse[OrganizationTeamArgs, OrganizationTeamState]{ID: strconv.FormatInt(team.ID, 10), Inputs: state.OrganizationTeamArgs, State: state}, nil
}

func (OrganizationTeam) Diff(_ context.Context, req infer.DiffRequest[OrganizationTeamArgs, OrganizationTeamState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	addReplaceDiff(diff, "organization", req.State.Organization != req.Inputs.Organization)
	addUpdateDiff(diff, "name", req.State.Name != req.Inputs.Name)
	addUpdateDiff(diff, "description", req.State.Description != req.Inputs.Description)
	addUpdateDiff(diff, "permission", req.State.Permission != req.Inputs.Permission)
	addUpdateDiff(diff, "canCreateOrgRepo", req.State.CanCreateOrgRepo != req.Inputs.CanCreateOrgRepo)
	addUpdateDiff(diff, "includesAllRepositories", req.State.IncludesAllRepositories != req.Inputs.IncludesAllRepositories)
	addUpdateDiff(diff, "unitsMap", !maps.Equal(req.State.UnitsMap, req.Inputs.UnitsMap))
	return p.DiffResponse{HasChanges: len(diff) > 0, DeleteBeforeReplace: true, DetailedDiff: diff}, nil
}

func (OrganizationTeam) Update(ctx context.Context, req infer.UpdateRequest[OrganizationTeamArgs, OrganizationTeamState]) (infer.UpdateResponse[OrganizationTeamState], error) {
	if req.DryRun {
		return infer.UpdateResponse[OrganizationTeamState]{Output: organizationTeamStateFromArgs(req.Inputs, req.State.TeamID)}, nil
	}

	teamID, err := organizationTeamID(req.ID, req.State)
	if err != nil {
		return infer.UpdateResponse[OrganizationTeamState]{}, err
	}
	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.UpdateResponse[OrganizationTeamState]{}, err
	}

	_, err = client.EditTeam(teamID, forgejo.EditTeamOption{
		Name:                    req.Inputs.Name,
		Description:             &req.Inputs.Description,
		Permission:              forgejo.AccessMode(req.Inputs.Permission),
		CanCreateOrgRepo:        &req.Inputs.CanCreateOrgRepo,
		IncludesAllRepositories: &req.Inputs.IncludesAllRepositories,
		UnitsMap:                req.Inputs.UnitsMap,
	})
	if err != nil {
		return infer.UpdateResponse[OrganizationTeamState]{}, err
	}
	team, _, err := client.GetTeam(teamID)
	if err != nil {
		return infer.UpdateResponse[OrganizationTeamState]{}, err
	}

	return infer.UpdateResponse[OrganizationTeamState]{Output: organizationTeamStateFromAPI(team, req.Inputs.Organization)}, nil
}

func (OrganizationTeam) Delete(ctx context.Context, req infer.DeleteRequest[OrganizationTeamState]) (infer.DeleteResponse, error) {
	teamID, err := organizationTeamID(req.ID, req.State)
	if err != nil {
		return infer.DeleteResponse{}, err
	}
	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.DeleteResponse{}, err
	}

	resp, err := client.DeleteTeam(teamID)
	if isNotFound(resp) {
		return infer.DeleteResponse{}, nil
	}
	return infer.DeleteResponse{}, err
}

func organizationTeamStateFromAPI(team *forgejo.Team, fallbackOrg string) OrganizationTeamState {
	org := fallbackOrg
	if team.Organization != nil && team.Organization.UserName != "" {
		org = team.Organization.UserName
	}
	return organizationTeamStateFromArgs(OrganizationTeamArgs{
		Organization:            org,
		Name:                    team.Name,
		Description:             team.Description,
		Permission:              string(team.Permission),
		CanCreateOrgRepo:        team.CanCreateOrgRepo,
		IncludesAllRepositories: team.IncludesAllRepositories,
		UnitsMap:                team.UnitsMap,
	}, team.ID)
}

func organizationTeamStateFromArgs(args OrganizationTeamArgs, teamID int64) OrganizationTeamState {
	return OrganizationTeamState{OrganizationTeamArgs: args, TeamID: teamID}
}

func organizationTeamID(id string, state OrganizationTeamState) (int64, error) {
	if state.TeamID != 0 {
		return state.TeamID, nil
	}
	teamID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("organization team ID must be numeric, got %q", id)
	}
	return teamID, nil
}

func organizationTeamPreviewID(args OrganizationTeamArgs) string {
	return args.Organization + "/" + args.Name
}
