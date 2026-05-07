package provider

import (
	"context"
	"fmt"
	"slices"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v3"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

type RepositoryBranchProtection struct{}

type RepositoryBranchProtectionArgs struct {
	Owner                         string   `pulumi:"owner"`
	Repository                    string   `pulumi:"repository"`
	Name                          string   `pulumi:"name"`
	EnablePush                    bool     `pulumi:"enablePush,optional"`
	EnablePushWhitelist           bool     `pulumi:"enablePushWhitelist,optional"`
	PushWhitelistUsernames        []string `pulumi:"pushWhitelistUsernames,optional"`
	PushWhitelistTeams            []string `pulumi:"pushWhitelistTeams,optional"`
	PushWhitelistDeployKeys       bool     `pulumi:"pushWhitelistDeployKeys,optional"`
	EnableMergeWhitelist          bool     `pulumi:"enableMergeWhitelist,optional"`
	MergeWhitelistUsernames       []string `pulumi:"mergeWhitelistUsernames,optional"`
	MergeWhitelistTeams           []string `pulumi:"mergeWhitelistTeams,optional"`
	EnableStatusCheck             bool     `pulumi:"enableStatusCheck,optional"`
	StatusCheckContexts           []string `pulumi:"statusCheckContexts,optional"`
	RequiredApprovals             int64    `pulumi:"requiredApprovals,optional"`
	EnableApprovalsWhitelist      bool     `pulumi:"enableApprovalsWhitelist,optional"`
	ApprovalsWhitelistUsernames   []string `pulumi:"approvalsWhitelistUsernames,optional"`
	ApprovalsWhitelistTeams       []string `pulumi:"approvalsWhitelistTeams,optional"`
	BlockOnRejectedReviews        bool     `pulumi:"blockOnRejectedReviews,optional"`
	BlockOnOfficialReviewRequests bool     `pulumi:"blockOnOfficialReviewRequests,optional"`
	BlockOnOutdatedBranch         bool     `pulumi:"blockOnOutdatedBranch,optional"`
	DismissStaleApprovals         bool     `pulumi:"dismissStaleApprovals,optional"`
	RequireSignedCommits          bool     `pulumi:"requireSignedCommits,optional"`
	ProtectedFilePatterns         string   `pulumi:"protectedFilePatterns,optional"`
	UnprotectedFilePatterns       string   `pulumi:"unprotectedFilePatterns,optional"`
}

type RepositoryBranchProtectionState struct {
	RepositoryBranchProtectionArgs
	Created string `pulumi:"created"`
	Updated string `pulumi:"updated"`
}

func (b *RepositoryBranchProtection) Annotate(a infer.Annotator) {
	a.Describe(b, "A Forgejo branch protection rule for a repository.")
	a.SetToken("index", "RepositoryBranchProtection")
}

func (a *RepositoryBranchProtectionArgs) Annotate(ann infer.Annotator) {
	ann.Describe(&a.Owner, "Repository owner.")
	ann.Describe(&a.Repository, "Repository name.")
	ann.Describe(&a.Name, "Protected branch name or rule pattern.")
	ann.Describe(&a.EnablePush, "Whether protected branches can be pushed to directly.")
	ann.Describe(&a.EnablePushWhitelist, "Whether direct pushes are limited to the push whitelist.")
	ann.Describe(&a.PushWhitelistUsernames, "Users allowed to push directly.")
	ann.Describe(&a.PushWhitelistTeams, "Teams allowed to push directly.")
	ann.Describe(&a.PushWhitelistDeployKeys, "Whether deploy keys may push directly.")
	ann.Describe(&a.EnableMergeWhitelist, "Whether merging is limited to the merge whitelist.")
	ann.Describe(&a.MergeWhitelistUsernames, "Users allowed to merge.")
	ann.Describe(&a.MergeWhitelistTeams, "Teams allowed to merge.")
	ann.Describe(&a.EnableStatusCheck, "Whether status checks are required before merge.")
	ann.Describe(&a.StatusCheckContexts, "Required status check contexts.")
	ann.Describe(&a.RequiredApprovals, "Number of required approving reviews.")
	ann.Describe(&a.EnableApprovalsWhitelist, "Whether review approvals are limited to the approval whitelist.")
	ann.Describe(&a.ApprovalsWhitelistUsernames, "Users whose approvals count.")
	ann.Describe(&a.ApprovalsWhitelistTeams, "Teams whose approvals count.")
	ann.Describe(&a.BlockOnRejectedReviews, "Whether rejected reviews block merging.")
	ann.Describe(&a.BlockOnOfficialReviewRequests, "Whether official review requests block merging.")
	ann.Describe(&a.BlockOnOutdatedBranch, "Whether outdated branches block merging.")
	ann.Describe(&a.DismissStaleApprovals, "Whether stale approvals are dismissed after new commits.")
	ann.Describe(&a.RequireSignedCommits, "Whether commits must be signed.")
	ann.Describe(&a.ProtectedFilePatterns, "Protected file patterns.")
	ann.Describe(&a.UnprotectedFilePatterns, "Unprotected file patterns.")
}

func (RepositoryBranchProtection) Create(ctx context.Context, req infer.CreateRequest[RepositoryBranchProtectionArgs]) (infer.CreateResponse[RepositoryBranchProtectionState], error) {
	if req.DryRun {
		return infer.CreateResponse[RepositoryBranchProtectionState]{ID: repositoryBranchProtectionID(req.Inputs.Owner, req.Inputs.Repository, req.Inputs.Name), Output: repositoryBranchProtectionStateFromArgs(req.Inputs, "", "")}, nil
	}

	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.CreateResponse[RepositoryBranchProtectionState]{}, err
	}
	protection, _, err := client.CreateBranchProtection(req.Inputs.Owner, req.Inputs.Repository, repositoryBranchProtectionCreateOption(req.Inputs))
	if err != nil {
		return infer.CreateResponse[RepositoryBranchProtectionState]{}, err
	}

	state := repositoryBranchProtectionStateFromAPI(req.Inputs.Owner, req.Inputs.Repository, req.Inputs.Name, protection)
	return infer.CreateResponse[RepositoryBranchProtectionState]{ID: repositoryBranchProtectionID(state.Owner, state.Repository, state.Name), Output: state}, nil
}

func (RepositoryBranchProtection) Read(ctx context.Context, req infer.ReadRequest[RepositoryBranchProtectionArgs, RepositoryBranchProtectionState]) (infer.ReadResponse[RepositoryBranchProtectionArgs, RepositoryBranchProtectionState], error) {
	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.ReadResponse[RepositoryBranchProtectionArgs, RepositoryBranchProtectionState]{}, err
	}

	protection, resp, err := client.GetBranchProtection(req.State.Owner, req.State.Repository, req.State.Name)
	if isNotFound(resp) {
		return infer.ReadResponse[RepositoryBranchProtectionArgs, RepositoryBranchProtectionState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[RepositoryBranchProtectionArgs, RepositoryBranchProtectionState]{}, err
	}

	state := repositoryBranchProtectionStateFromAPI(req.State.Owner, req.State.Repository, req.State.Name, protection)
	return infer.ReadResponse[RepositoryBranchProtectionArgs, RepositoryBranchProtectionState]{ID: repositoryBranchProtectionID(state.Owner, state.Repository, state.Name), Inputs: state.RepositoryBranchProtectionArgs, State: state}, nil
}

func (RepositoryBranchProtection) Diff(_ context.Context, req infer.DiffRequest[RepositoryBranchProtectionArgs, RepositoryBranchProtectionState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	addReplaceDiff(diff, "owner", req.State.Owner != req.Inputs.Owner)
	addReplaceDiff(diff, "repository", req.State.Repository != req.Inputs.Repository)
	addReplaceDiff(diff, "name", req.State.Name != req.Inputs.Name)
	addUpdateDiff(diff, "enablePush", req.State.EnablePush != req.Inputs.EnablePush)
	addUpdateDiff(diff, "enablePushWhitelist", req.State.EnablePushWhitelist != req.Inputs.EnablePushWhitelist)
	addUpdateDiff(diff, "pushWhitelistUsernames", !slices.Equal(req.State.PushWhitelistUsernames, req.Inputs.PushWhitelistUsernames))
	addUpdateDiff(diff, "pushWhitelistTeams", !slices.Equal(req.State.PushWhitelistTeams, req.Inputs.PushWhitelistTeams))
	addUpdateDiff(diff, "pushWhitelistDeployKeys", req.State.PushWhitelistDeployKeys != req.Inputs.PushWhitelistDeployKeys)
	addUpdateDiff(diff, "enableMergeWhitelist", req.State.EnableMergeWhitelist != req.Inputs.EnableMergeWhitelist)
	addUpdateDiff(diff, "mergeWhitelistUsernames", !slices.Equal(req.State.MergeWhitelistUsernames, req.Inputs.MergeWhitelistUsernames))
	addUpdateDiff(diff, "mergeWhitelistTeams", !slices.Equal(req.State.MergeWhitelistTeams, req.Inputs.MergeWhitelistTeams))
	addUpdateDiff(diff, "enableStatusCheck", req.State.EnableStatusCheck != req.Inputs.EnableStatusCheck)
	addUpdateDiff(diff, "statusCheckContexts", !slices.Equal(req.State.StatusCheckContexts, req.Inputs.StatusCheckContexts))
	addUpdateDiff(diff, "requiredApprovals", req.State.RequiredApprovals != req.Inputs.RequiredApprovals)
	addUpdateDiff(diff, "enableApprovalsWhitelist", req.State.EnableApprovalsWhitelist != req.Inputs.EnableApprovalsWhitelist)
	addUpdateDiff(diff, "approvalsWhitelistUsernames", !slices.Equal(req.State.ApprovalsWhitelistUsernames, req.Inputs.ApprovalsWhitelistUsernames))
	addUpdateDiff(diff, "approvalsWhitelistTeams", !slices.Equal(req.State.ApprovalsWhitelistTeams, req.Inputs.ApprovalsWhitelistTeams))
	addUpdateDiff(diff, "blockOnRejectedReviews", req.State.BlockOnRejectedReviews != req.Inputs.BlockOnRejectedReviews)
	addUpdateDiff(diff, "blockOnOfficialReviewRequests", req.State.BlockOnOfficialReviewRequests != req.Inputs.BlockOnOfficialReviewRequests)
	addUpdateDiff(diff, "blockOnOutdatedBranch", req.State.BlockOnOutdatedBranch != req.Inputs.BlockOnOutdatedBranch)
	addUpdateDiff(diff, "dismissStaleApprovals", req.State.DismissStaleApprovals != req.Inputs.DismissStaleApprovals)
	addUpdateDiff(diff, "requireSignedCommits", req.State.RequireSignedCommits != req.Inputs.RequireSignedCommits)
	addUpdateDiff(diff, "protectedFilePatterns", req.State.ProtectedFilePatterns != req.Inputs.ProtectedFilePatterns)
	addUpdateDiff(diff, "unprotectedFilePatterns", req.State.UnprotectedFilePatterns != req.Inputs.UnprotectedFilePatterns)
	return p.DiffResponse{HasChanges: len(diff) > 0, DeleteBeforeReplace: true, DetailedDiff: diff}, nil
}

func (RepositoryBranchProtection) Update(ctx context.Context, req infer.UpdateRequest[RepositoryBranchProtectionArgs, RepositoryBranchProtectionState]) (infer.UpdateResponse[RepositoryBranchProtectionState], error) {
	if req.DryRun {
		return infer.UpdateResponse[RepositoryBranchProtectionState]{Output: repositoryBranchProtectionStateFromArgs(req.Inputs, req.State.Created, req.State.Updated)}, nil
	}

	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.UpdateResponse[RepositoryBranchProtectionState]{}, err
	}
	protection, _, err := client.EditBranchProtection(req.State.Owner, req.State.Repository, req.State.Name, repositoryBranchProtectionEditOption(req.Inputs))
	if err != nil {
		return infer.UpdateResponse[RepositoryBranchProtectionState]{}, err
	}

	return infer.UpdateResponse[RepositoryBranchProtectionState]{Output: repositoryBranchProtectionStateFromAPI(req.Inputs.Owner, req.Inputs.Repository, req.Inputs.Name, protection)}, nil
}

func (RepositoryBranchProtection) Delete(ctx context.Context, req infer.DeleteRequest[RepositoryBranchProtectionState]) (infer.DeleteResponse, error) {
	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.DeleteResponse{}, err
	}

	resp, err := client.DeleteBranchProtection(req.State.Owner, req.State.Repository, req.State.Name)
	if isNotFound(resp) {
		return infer.DeleteResponse{}, nil
	}
	return infer.DeleteResponse{}, err
}

func repositoryBranchProtectionCreateOption(args RepositoryBranchProtectionArgs) forgejo.CreateBranchProtectionOption {
	return forgejo.CreateBranchProtectionOption{
		BranchName:                    args.Name,
		RuleName:                      args.Name,
		EnablePush:                    args.EnablePush,
		EnablePushWhitelist:           args.EnablePushWhitelist,
		PushWhitelistUsernames:        args.PushWhitelistUsernames,
		PushWhitelistTeams:            args.PushWhitelistTeams,
		PushWhitelistDeployKeys:       args.PushWhitelistDeployKeys,
		EnableMergeWhitelist:          args.EnableMergeWhitelist,
		MergeWhitelistUsernames:       args.MergeWhitelistUsernames,
		MergeWhitelistTeams:           args.MergeWhitelistTeams,
		EnableStatusCheck:             args.EnableStatusCheck,
		StatusCheckContexts:           args.StatusCheckContexts,
		RequiredApprovals:             args.RequiredApprovals,
		EnableApprovalsWhitelist:      args.EnableApprovalsWhitelist,
		ApprovalsWhitelistUsernames:   args.ApprovalsWhitelistUsernames,
		ApprovalsWhitelistTeams:       args.ApprovalsWhitelistTeams,
		BlockOnRejectedReviews:        args.BlockOnRejectedReviews,
		BlockOnOfficialReviewRequests: args.BlockOnOfficialReviewRequests,
		BlockOnOutdatedBranch:         args.BlockOnOutdatedBranch,
		DismissStaleApprovals:         args.DismissStaleApprovals,
		RequireSignedCommits:          args.RequireSignedCommits,
		ProtectedFilePatterns:         args.ProtectedFilePatterns,
		UnprotectedFilePatterns:       args.UnprotectedFilePatterns,
	}
}

func repositoryBranchProtectionEditOption(args RepositoryBranchProtectionArgs) forgejo.EditBranchProtectionOption {
	return forgejo.EditBranchProtectionOption{
		EnablePush:                    &args.EnablePush,
		EnablePushWhitelist:           &args.EnablePushWhitelist,
		PushWhitelistUsernames:        args.PushWhitelistUsernames,
		PushWhitelistTeams:            args.PushWhitelistTeams,
		PushWhitelistDeployKeys:       &args.PushWhitelistDeployKeys,
		EnableMergeWhitelist:          &args.EnableMergeWhitelist,
		MergeWhitelistUsernames:       args.MergeWhitelistUsernames,
		MergeWhitelistTeams:           args.MergeWhitelistTeams,
		EnableStatusCheck:             &args.EnableStatusCheck,
		StatusCheckContexts:           args.StatusCheckContexts,
		RequiredApprovals:             &args.RequiredApprovals,
		EnableApprovalsWhitelist:      &args.EnableApprovalsWhitelist,
		ApprovalsWhitelistUsernames:   args.ApprovalsWhitelistUsernames,
		ApprovalsWhitelistTeams:       args.ApprovalsWhitelistTeams,
		BlockOnRejectedReviews:        &args.BlockOnRejectedReviews,
		BlockOnOfficialReviewRequests: &args.BlockOnOfficialReviewRequests,
		BlockOnOutdatedBranch:         &args.BlockOnOutdatedBranch,
		DismissStaleApprovals:         &args.DismissStaleApprovals,
		RequireSignedCommits:          &args.RequireSignedCommits,
		ProtectedFilePatterns:         &args.ProtectedFilePatterns,
		UnprotectedFilePatterns:       &args.UnprotectedFilePatterns,
	}
}

func repositoryBranchProtectionStateFromAPI(owner, repo, fallbackName string, protection *forgejo.BranchProtection) RepositoryBranchProtectionState {
	if protection == nil {
		return repositoryBranchProtectionStateFromArgs(RepositoryBranchProtectionArgs{Owner: owner, Repository: repo, Name: fallbackName}, "", "")
	}
	name := protection.RuleName
	if name == "" {
		name = protection.BranchName
	}
	if name == "" {
		name = fallbackName
	}

	return repositoryBranchProtectionStateFromArgs(RepositoryBranchProtectionArgs{
		Owner:                         owner,
		Repository:                    repo,
		Name:                          name,
		EnablePush:                    protection.EnablePush,
		EnablePushWhitelist:           protection.EnablePushWhitelist,
		PushWhitelistUsernames:        protection.PushWhitelistUsernames,
		PushWhitelistTeams:            protection.PushWhitelistTeams,
		PushWhitelistDeployKeys:       protection.PushWhitelistDeployKeys,
		EnableMergeWhitelist:          protection.EnableMergeWhitelist,
		MergeWhitelistUsernames:       protection.MergeWhitelistUsernames,
		MergeWhitelistTeams:           protection.MergeWhitelistTeams,
		EnableStatusCheck:             protection.EnableStatusCheck,
		StatusCheckContexts:           protection.StatusCheckContexts,
		RequiredApprovals:             protection.RequiredApprovals,
		EnableApprovalsWhitelist:      protection.EnableApprovalsWhitelist,
		ApprovalsWhitelistUsernames:   protection.ApprovalsWhitelistUsernames,
		ApprovalsWhitelistTeams:       protection.ApprovalsWhitelistTeams,
		BlockOnRejectedReviews:        protection.BlockOnRejectedReviews,
		BlockOnOfficialReviewRequests: protection.BlockOnOfficialReviewRequests,
		BlockOnOutdatedBranch:         protection.BlockOnOutdatedBranch,
		DismissStaleApprovals:         protection.DismissStaleApprovals,
		RequireSignedCommits:          protection.RequireSignedCommits,
		ProtectedFilePatterns:         protection.ProtectedFilePatterns,
		UnprotectedFilePatterns:       protection.UnprotectedFilePatterns,
	}, formatForgejoTime(protection.Created), formatForgejoTime(protection.Updated))
}

func repositoryBranchProtectionStateFromArgs(args RepositoryBranchProtectionArgs, created, updated string) RepositoryBranchProtectionState {
	return RepositoryBranchProtectionState{RepositoryBranchProtectionArgs: args, Created: created, Updated: updated}
}

func repositoryBranchProtectionID(owner, repo, name string) string {
	return fmt.Sprintf("%s/%s/%s", owner, repo, name)
}
