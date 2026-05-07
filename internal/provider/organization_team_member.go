package provider

import (
	"context"
	"fmt"
	"strings"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v3"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

type OrganizationTeamMember struct{}

type OrganizationTeamMemberArgs struct {
	Organization string `pulumi:"organization,optional"`
	Team         string `pulumi:"team,optional"`
	TeamID       int64  `pulumi:"teamId,optional"`
	Username     string `pulumi:"username"`
}

type OrganizationTeamMemberState struct {
	OrganizationTeamMemberArgs
	UserID int64 `pulumi:"userId"`
}

func (m *OrganizationTeamMember) Annotate(a infer.Annotator) {
	a.Describe(m, "A Forgejo organization team membership.")
	a.SetToken("index", "OrganizationTeamMember")
}

func (a *OrganizationTeamMemberArgs) Annotate(ann infer.Annotator) {
	ann.Describe(&a.Organization, "Organization username/slug. Required when teamId is not set.")
	ann.Describe(&a.Team, "Team name. Required when teamId is not set.")
	ann.Describe(&a.TeamID, "Numeric Forgejo team ID. If omitted, organization and team are used to look up the team.")
	ann.Describe(&a.Username, "Username to add to the team.")
}

func (OrganizationTeamMember) Create(ctx context.Context, req infer.CreateRequest[OrganizationTeamMemberArgs]) (infer.CreateResponse[OrganizationTeamMemberState], error) {
	if err := validateOrganizationTeamMemberArgs(req.Inputs); err != nil {
		return infer.CreateResponse[OrganizationTeamMemberState]{}, err
	}
	if req.DryRun {
		return infer.CreateResponse[OrganizationTeamMemberState]{ID: organizationTeamMemberPreviewID(req.Inputs), Output: organizationTeamMemberStateFromArgs(req.Inputs, 0)}, nil
	}

	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.CreateResponse[OrganizationTeamMemberState]{}, err
	}
	teamID, _, found, err := organizationTeamMemberTeamID(client, req.Inputs)
	if err != nil {
		return infer.CreateResponse[OrganizationTeamMemberState]{}, err
	}
	if !found {
		return infer.CreateResponse[OrganizationTeamMemberState]{}, fmt.Errorf("organization team %q not found in organization %q", req.Inputs.Team, req.Inputs.Organization)
	}

	_, err = client.AddTeamMember(teamID, req.Inputs.Username)
	if err != nil {
		return infer.CreateResponse[OrganizationTeamMemberState]{}, err
	}
	user, _, err := client.GetTeamMember(teamID, req.Inputs.Username)
	if err != nil {
		return infer.CreateResponse[OrganizationTeamMemberState]{}, err
	}

	state := organizationTeamMemberStateFromAPI(req.Inputs, teamID, user)
	return infer.CreateResponse[OrganizationTeamMemberState]{ID: organizationTeamMemberID(state.TeamID, state.Username), Output: state}, nil
}

func (OrganizationTeamMember) Read(ctx context.Context, req infer.ReadRequest[OrganizationTeamMemberArgs, OrganizationTeamMemberState]) (infer.ReadResponse[OrganizationTeamMemberArgs, OrganizationTeamMemberState], error) {
	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.ReadResponse[OrganizationTeamMemberArgs, OrganizationTeamMemberState]{}, err
	}

	teamID, resp, found, err := organizationTeamMemberTeamID(client, req.State.OrganizationTeamMemberArgs)
	if isNotFound(resp) {
		return infer.ReadResponse[OrganizationTeamMemberArgs, OrganizationTeamMemberState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[OrganizationTeamMemberArgs, OrganizationTeamMemberState]{}, err
	}
	if !found {
		return infer.ReadResponse[OrganizationTeamMemberArgs, OrganizationTeamMemberState]{}, nil
	}

	user, resp, err := client.GetTeamMember(teamID, req.State.Username)
	if isNotFound(resp) {
		return infer.ReadResponse[OrganizationTeamMemberArgs, OrganizationTeamMemberState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[OrganizationTeamMemberArgs, OrganizationTeamMemberState]{}, err
	}

	state := organizationTeamMemberStateFromAPI(req.State.OrganizationTeamMemberArgs, teamID, user)
	return infer.ReadResponse[OrganizationTeamMemberArgs, OrganizationTeamMemberState]{ID: organizationTeamMemberID(state.TeamID, state.Username), Inputs: state.OrganizationTeamMemberArgs, State: state}, nil
}

func (OrganizationTeamMember) Diff(_ context.Context, req infer.DiffRequest[OrganizationTeamMemberArgs, OrganizationTeamMemberState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	if req.Inputs.TeamID > 0 {
		addReplaceDiff(diff, "teamId", req.State.TeamID != req.Inputs.TeamID)
	} else {
		addReplaceDiff(diff, "organization", req.State.Organization != req.Inputs.Organization)
		addReplaceDiff(diff, "team", req.State.Team != req.Inputs.Team)
	}
	addReplaceDiff(diff, "username", req.State.Username != req.Inputs.Username)
	return p.DiffResponse{HasChanges: len(diff) > 0, DeleteBeforeReplace: true, DetailedDiff: diff}, nil
}

func (OrganizationTeamMember) Delete(ctx context.Context, req infer.DeleteRequest[OrganizationTeamMemberState]) (infer.DeleteResponse, error) {
	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.DeleteResponse{}, err
	}

	teamID, resp, found, err := organizationTeamMemberTeamID(client, req.State.OrganizationTeamMemberArgs)
	if isNotFound(resp) {
		return infer.DeleteResponse{}, nil
	}
	if err != nil {
		return infer.DeleteResponse{}, err
	}
	if !found {
		return infer.DeleteResponse{}, nil
	}

	resp, err = client.RemoveTeamMember(teamID, req.State.Username)
	if isNotFound(resp) {
		return infer.DeleteResponse{}, nil
	}
	return infer.DeleteResponse{}, err
}

func validateOrganizationTeamMemberArgs(args OrganizationTeamMemberArgs) error {
	if strings.TrimSpace(args.Username) == "" {
		return fmt.Errorf("username is required")
	}
	if args.TeamID <= 0 && (strings.TrimSpace(args.Organization) == "" || strings.TrimSpace(args.Team) == "") {
		return fmt.Errorf("teamId or both organization and team are required")
	}
	return nil
}

func organizationTeamMemberTeamID(client *forgejo.Client, args OrganizationTeamMemberArgs) (int64, *forgejo.Response, bool, error) {
	if args.TeamID > 0 {
		return args.TeamID, nil, true, nil
	}
	if err := validateOrganizationTeamMemberArgs(args); err != nil {
		return 0, nil, false, err
	}

	teams, resp, err := client.ListOrgTeams(args.Organization, forgejo.ListTeamsOptions{ListOptions: forgejo.ListOptions{Page: -1}})
	if err != nil {
		return 0, resp, false, err
	}
	for _, team := range teams {
		if team.Name == args.Team {
			return team.ID, resp, true, nil
		}
	}
	return 0, resp, false, nil
}

func organizationTeamMemberStateFromAPI(args OrganizationTeamMemberArgs, teamID int64, user *forgejo.User) OrganizationTeamMemberState {
	args.TeamID = teamID
	userID := int64(0)
	if user != nil {
		userID = user.ID
		if user.UserName != "" {
			args.Username = user.UserName
		}
	}
	return organizationTeamMemberStateFromArgs(args, userID)
}

func organizationTeamMemberStateFromArgs(args OrganizationTeamMemberArgs, userID int64) OrganizationTeamMemberState {
	return OrganizationTeamMemberState{OrganizationTeamMemberArgs: args, UserID: userID}
}

func organizationTeamMemberID(teamID int64, username string) string {
	return fmt.Sprintf("%d/%s", teamID, username)
}

func organizationTeamMemberPreviewID(args OrganizationTeamMemberArgs) string {
	if args.TeamID > 0 {
		return organizationTeamMemberID(args.TeamID, args.Username)
	}
	return fmt.Sprintf("%s/%s/%s", args.Organization, args.Team, args.Username)
}
