package provider

import (
	"context"
	"fmt"
	"slices"
	"strconv"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v3"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

type RepositoryTagProtection struct{}

type RepositoryTagProtectionArgs struct {
	Owner              string   `pulumi:"owner"`
	Repository         string   `pulumi:"repository"`
	NamePattern        string   `pulumi:"namePattern"`
	WhitelistUsernames []string `pulumi:"whitelistUsernames,optional"`
	WhitelistTeams     []string `pulumi:"whitelistTeams,optional"`
}

type RepositoryTagProtectionState struct {
	RepositoryTagProtectionArgs
	ProtectionID int64 `pulumi:"protectionId"`
}

func (t *RepositoryTagProtection) Annotate(a infer.Annotator) {
	a.Describe(t, "A Forgejo tag protection rule for a repository.")
	a.SetToken("index", "RepositoryTagProtection")
}

func (a *RepositoryTagProtectionArgs) Annotate(ann infer.Annotator) {
	ann.Describe(&a.Owner, "Repository owner.")
	ann.Describe(&a.Repository, "Repository name.")
	ann.Describe(&a.NamePattern, "Protected tag name pattern, for example v*.")
	ann.Describe(&a.WhitelistUsernames, "Users allowed to create matching tags.")
	ann.Describe(&a.WhitelistTeams, "Teams allowed to create matching tags.")
}

func (RepositoryTagProtection) Create(ctx context.Context, req infer.CreateRequest[RepositoryTagProtectionArgs]) (infer.CreateResponse[RepositoryTagProtectionState], error) {
	if req.DryRun {
		return infer.CreateResponse[RepositoryTagProtectionState]{ID: repositoryTagProtectionPreviewID(req.Inputs), Output: repositoryTagProtectionStateFromArgs(req.Inputs, 0)}, nil
	}

	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.CreateResponse[RepositoryTagProtectionState]{}, err
	}

	protection, _, err := client.CreateTagProtection(req.Inputs.Owner, req.Inputs.Repository, forgejo.CreateTagProtectionOption{
		NamePattern:        req.Inputs.NamePattern,
		WhitelistUsernames: req.Inputs.WhitelistUsernames,
		WhitelistTeams:     req.Inputs.WhitelistTeams,
	})
	if err != nil {
		return infer.CreateResponse[RepositoryTagProtectionState]{}, err
	}

	state := repositoryTagProtectionStateFromAPI(req.Inputs.Owner, req.Inputs.Repository, protection)
	return infer.CreateResponse[RepositoryTagProtectionState]{ID: strconv.FormatInt(protection.ID, 10), Output: state}, nil
}

func (RepositoryTagProtection) Read(ctx context.Context, req infer.ReadRequest[RepositoryTagProtectionArgs, RepositoryTagProtectionState]) (infer.ReadResponse[RepositoryTagProtectionArgs, RepositoryTagProtectionState], error) {
	protectionID, err := repositoryTagProtectionID(req.ID, req.State)
	if err != nil {
		return infer.ReadResponse[RepositoryTagProtectionArgs, RepositoryTagProtectionState]{}, err
	}

	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.ReadResponse[RepositoryTagProtectionArgs, RepositoryTagProtectionState]{}, err
	}

	protection, resp, err := client.GetTagProtection(req.State.Owner, req.State.Repository, protectionID)
	if isNotFound(resp) {
		return infer.ReadResponse[RepositoryTagProtectionArgs, RepositoryTagProtectionState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[RepositoryTagProtectionArgs, RepositoryTagProtectionState]{}, err
	}

	state := repositoryTagProtectionStateFromAPI(req.State.Owner, req.State.Repository, protection)
	return infer.ReadResponse[RepositoryTagProtectionArgs, RepositoryTagProtectionState]{ID: strconv.FormatInt(protection.ID, 10), Inputs: state.RepositoryTagProtectionArgs, State: state}, nil
}

func (RepositoryTagProtection) Diff(_ context.Context, req infer.DiffRequest[RepositoryTagProtectionArgs, RepositoryTagProtectionState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	addReplaceDiff(diff, "owner", req.State.Owner != req.Inputs.Owner)
	addReplaceDiff(diff, "repository", req.State.Repository != req.Inputs.Repository)
	addUpdateDiff(diff, "namePattern", req.State.NamePattern != req.Inputs.NamePattern)
	addUpdateDiff(diff, "whitelistUsernames", !slices.Equal(req.State.WhitelistUsernames, req.Inputs.WhitelistUsernames))
	addUpdateDiff(diff, "whitelistTeams", !slices.Equal(req.State.WhitelistTeams, req.Inputs.WhitelistTeams))
	return p.DiffResponse{HasChanges: len(diff) > 0, DeleteBeforeReplace: true, DetailedDiff: diff}, nil
}

func (RepositoryTagProtection) Update(ctx context.Context, req infer.UpdateRequest[RepositoryTagProtectionArgs, RepositoryTagProtectionState]) (infer.UpdateResponse[RepositoryTagProtectionState], error) {
	if req.DryRun {
		return infer.UpdateResponse[RepositoryTagProtectionState]{Output: repositoryTagProtectionStateFromArgs(req.Inputs, req.State.ProtectionID)}, nil
	}

	protectionID, err := repositoryTagProtectionID(req.ID, req.State)
	if err != nil {
		return infer.UpdateResponse[RepositoryTagProtectionState]{}, err
	}
	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.UpdateResponse[RepositoryTagProtectionState]{}, err
	}

	protection, _, err := client.EditTagProtection(req.State.Owner, req.State.Repository, protectionID, forgejo.EditTagProtectionOption{
		NamePattern:        &req.Inputs.NamePattern,
		WhitelistUsernames: req.Inputs.WhitelistUsernames,
		WhitelistTeams:     req.Inputs.WhitelistTeams,
	})
	if err != nil {
		return infer.UpdateResponse[RepositoryTagProtectionState]{}, err
	}

	return infer.UpdateResponse[RepositoryTagProtectionState]{Output: repositoryTagProtectionStateFromAPI(req.Inputs.Owner, req.Inputs.Repository, protection)}, nil
}

func (RepositoryTagProtection) Delete(ctx context.Context, req infer.DeleteRequest[RepositoryTagProtectionState]) (infer.DeleteResponse, error) {
	protectionID, err := repositoryTagProtectionID(req.ID, req.State)
	if err != nil {
		return infer.DeleteResponse{}, err
	}
	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.DeleteResponse{}, err
	}

	resp, err := client.DeleteTagProtection(req.State.Owner, req.State.Repository, protectionID)
	if isNotFound(resp) {
		return infer.DeleteResponse{}, nil
	}
	return infer.DeleteResponse{}, err
}

func repositoryTagProtectionStateFromAPI(owner, repo string, protection *forgejo.TagProtection) RepositoryTagProtectionState {
	return repositoryTagProtectionStateFromArgs(RepositoryTagProtectionArgs{
		Owner:              owner,
		Repository:         repo,
		NamePattern:        protection.NamePattern,
		WhitelistUsernames: protection.WhitelistUsernames,
		WhitelistTeams:     protection.WhitelistTeams,
	}, protection.ID)
}

func repositoryTagProtectionStateFromArgs(args RepositoryTagProtectionArgs, protectionID int64) RepositoryTagProtectionState {
	return RepositoryTagProtectionState{RepositoryTagProtectionArgs: args, ProtectionID: protectionID}
}

func repositoryTagProtectionID(id string, state RepositoryTagProtectionState) (int64, error) {
	if state.ProtectionID != 0 {
		return state.ProtectionID, nil
	}
	protectionID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("repository tag protection ID must be numeric, got %q", id)
	}
	return protectionID, nil
}

func repositoryTagProtectionPreviewID(args RepositoryTagProtectionArgs) string {
	return fmt.Sprintf("%s/%s/%s", args.Owner, args.Repository, args.NamePattern)
}
