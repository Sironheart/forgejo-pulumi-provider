package provider

import (
	"context"
	"reflect"
	"testing"

	"github.com/pulumi/pulumi-go-provider/infer"
)

func TestRepositoryCreateDryRunBuildsPreviewState(t *testing.T) {
	t.Parallel()

	args := RepositoryArgs{
		Name:          "infra",
		Owner:         "sironheart",
		Description:   "Infrastructure repository",
		Private:       true,
		DefaultBranch: "main",
		Website:       "https://example.test",
		Template:      true,
		Settings:      &RepositorySettingsConfig{Actions: boolPtr(true), ExternalWikiURL: stringPtr("https://wiki.example.test/infra")},
	}
	resp, err := (Repository{}).Create(context.Background(), infer.CreateRequest[RepositoryArgs]{Inputs: args, DryRun: true})
	if err != nil {
		t.Fatalf("create dry-run: %v", err)
	}

	assertEqual(t, resp.ID, "sironheart/infra")
	assertEqual(t, resp.Output.RepositoryArgs, args)
	assertEqual(t, resp.Output.FullName, "sironheart/infra")
}

func TestRepositoryUpdateDryRunPreservesComputedURLs(t *testing.T) {
	t.Parallel()

	inputs := RepositoryArgs{Name: "infra", Owner: "sironheart", Description: "updated"}
	state := RepositoryState{
		RepositoryArgs: RepositoryArgs{Name: "infra", Owner: "sironheart", Description: "old"},
		HTMLURL:        "https://forgejo.test/sironheart/infra",
		SSHURL:         "ssh://git@forgejo.test/sironheart/infra.git",
		CloneURL:       "https://forgejo.test/sironheart/infra.git",
	}
	resp, err := (Repository{}).Update(context.Background(), infer.UpdateRequest[RepositoryArgs, RepositoryState]{Inputs: inputs, State: state, DryRun: true})
	if err != nil {
		t.Fatalf("update dry-run: %v", err)
	}

	assertEqual(t, resp.Output.RepositoryArgs, inputs)
	assertEqual(t, resp.Output.HTMLURL, state.HTMLURL)
	assertEqual(t, resp.Output.SSHURL, state.SSHURL)
	assertEqual(t, resp.Output.CloneURL, state.CloneURL)
}

func TestOrganizationCreateDryRunBuildsPreviewState(t *testing.T) {
	t.Parallel()

	args := OrganizationArgs{Name: "platform", FullName: "Platform", Description: "Team org", Website: "https://example.test", Location: "Remote", Visibility: "private"}
	resp, err := (Organization{}).Create(context.Background(), infer.CreateRequest[OrganizationArgs]{Inputs: args, DryRun: true})
	if err != nil {
		t.Fatalf("create dry-run: %v", err)
	}

	assertEqual(t, resp.ID, "platform")
	assertEqual(t, resp.Output.OrganizationArgs, args)
	assertEqual(t, resp.Output.AvatarURL, "")
}

func TestOrganizationUpdateDryRunPreservesAvatarURL(t *testing.T) {
	t.Parallel()

	inputs := OrganizationArgs{Name: "platform", FullName: "Platform Team"}
	state := OrganizationState{OrganizationArgs: OrganizationArgs{Name: "platform", FullName: "Platform"}, AvatarURL: "https://forgejo.test/avatars/platform"}
	resp, err := (Organization{}).Update(context.Background(), infer.UpdateRequest[OrganizationArgs, OrganizationState]{Inputs: inputs, State: state, DryRun: true})
	if err != nil {
		t.Fatalf("update dry-run: %v", err)
	}

	assertEqual(t, resp.Output.OrganizationArgs, inputs)
	assertEqual(t, resp.Output.AvatarURL, state.AvatarURL)
}

func TestOrganizationTeamCreateDryRunBuildsPreviewState(t *testing.T) {
	t.Parallel()

	args := OrganizationTeamArgs{
		Organization:            "platform",
		Name:                    "maintainers",
		Description:             "Maintainers",
		Permission:              "write",
		CanCreateOrgRepo:        true,
		IncludesAllRepositories: true,
		UnitsMap:                map[string]string{"repo.code": "write"},
	}
	resp, err := (OrganizationTeam{}).Create(context.Background(), infer.CreateRequest[OrganizationTeamArgs]{Inputs: args, DryRun: true})
	if err != nil {
		t.Fatalf("create dry-run: %v", err)
	}

	assertEqual(t, resp.ID, "platform/maintainers")
	assertEqual(t, resp.Output.OrganizationTeamArgs, args)
	assertEqual(t, resp.Output.TeamID, int64(0))
}

func TestOrganizationTeamUpdateDryRunPreservesTeamID(t *testing.T) {
	t.Parallel()

	inputs := OrganizationTeamArgs{Organization: "platform", Name: "maintainers", Permission: "admin"}
	state := OrganizationTeamState{OrganizationTeamArgs: OrganizationTeamArgs{Organization: "platform", Name: "maintainers", Permission: "write"}, TeamID: 42}
	resp, err := (OrganizationTeam{}).Update(context.Background(), infer.UpdateRequest[OrganizationTeamArgs, OrganizationTeamState]{Inputs: inputs, State: state, DryRun: true})
	if err != nil {
		t.Fatalf("update dry-run: %v", err)
	}

	assertEqual(t, resp.Output.OrganizationTeamArgs, inputs)
	assertEqual(t, resp.Output.TeamID, int64(42))
}

func TestOrganizationTeamMemberCreateDryRunBuildsPreviewState(t *testing.T) {
	t.Parallel()

	args := OrganizationTeamMemberArgs{Organization: "platform", Team: "maintainers", Username: "alice"}
	resp, err := (OrganizationTeamMember{}).Create(context.Background(), infer.CreateRequest[OrganizationTeamMemberArgs]{Inputs: args, DryRun: true})
	if err != nil {
		t.Fatalf("create dry-run: %v", err)
	}

	assertEqual(t, resp.ID, "platform/maintainers/alice")
	assertEqual(t, resp.Output.OrganizationTeamMemberArgs, args)
	assertEqual(t, resp.Output.UserID, int64(0))
}

func TestOrganizationTeamMemberCreateDryRunUsesTeamIDPreview(t *testing.T) {
	t.Parallel()

	args := OrganizationTeamMemberArgs{TeamID: 42, Username: "alice"}
	resp, err := (OrganizationTeamMember{}).Create(context.Background(), infer.CreateRequest[OrganizationTeamMemberArgs]{Inputs: args, DryRun: true})
	if err != nil {
		t.Fatalf("create dry-run: %v", err)
	}

	assertEqual(t, resp.ID, "42/alice")
	assertEqual(t, resp.Output.OrganizationTeamMemberArgs, args)
	assertEqual(t, resp.Output.UserID, int64(0))
}

func TestDeployKeyCreateDryRunBuildsPreviewState(t *testing.T) {
	t.Parallel()

	args := DeployKeyArgs{Owner: "sironheart", Repository: "infra", Title: "ci", Key: "ssh-ed25519 AAAA", ReadOnly: true}
	resp, err := (DeployKey{}).Create(context.Background(), infer.CreateRequest[DeployKeyArgs]{Inputs: args, DryRun: true})
	if err != nil {
		t.Fatalf("create dry-run: %v", err)
	}

	assertEqual(t, resp.ID, "sironheart/infra/ci")
	assertEqual(t, resp.Output.DeployKeyArgs, args)
	assertEqual(t, resp.Output.KeyID, int64(0))
	assertEqual(t, resp.Output.URL, "")
	assertEqual(t, resp.Output.Fingerprint, "")
}

func TestPublicKeyCreateDryRunBuildsPreviewState(t *testing.T) {
	t.Parallel()

	args := PublicKeyArgs{Title: "laptop", Key: "ssh-ed25519 AAAA", ReadOnly: true}
	resp, err := (PublicKey{}).Create(context.Background(), infer.CreateRequest[PublicKeyArgs]{Inputs: args, DryRun: true})
	if err != nil {
		t.Fatalf("create dry-run: %v", err)
	}

	assertEqual(t, resp.ID, "laptop")
	assertEqual(t, resp.Output.PublicKeyArgs, args)
	assertEqual(t, resp.Output.KeyID, int64(0))
	assertEqual(t, resp.Output.Owner, "")
}

func TestRepositoryActionVariableCreateDryRunBuildsPreviewState(t *testing.T) {
	t.Parallel()

	args := RepositoryActionVariableArgs{ActionVariableArgs: ActionVariableArgs{Name: "ENVIRONMENT", Value: "prod"}, Owner: "sironheart", Repository: "infra"}
	resp, err := (RepositoryActionVariable{}).Create(context.Background(), infer.CreateRequest[RepositoryActionVariableArgs]{Inputs: args, DryRun: true})
	if err != nil {
		t.Fatalf("create dry-run: %v", err)
	}

	assertEqual(t, resp.ID, "sironheart/infra/ENVIRONMENT")
	assertEqual(t, resp.Output.RepositoryActionVariableArgs, args)
	assertEqual(t, resp.Output.OwnerID, int64(0))
	assertEqual(t, resp.Output.RepoID, int64(0))
}

func TestRepositoryActionVariableUpdateDryRunPreservesIDs(t *testing.T) {
	t.Parallel()

	inputs := RepositoryActionVariableArgs{ActionVariableArgs: ActionVariableArgs{Name: "ENVIRONMENT", Value: "prod"}, Owner: "sironheart", Repository: "infra"}
	state := RepositoryActionVariableState{RepositoryActionVariableArgs: RepositoryActionVariableArgs{ActionVariableArgs: ActionVariableArgs{Name: "ENVIRONMENT", Value: "dev"}, Owner: "sironheart", Repository: "infra"}, OwnerID: 10, RepoID: 20}
	resp, err := (RepositoryActionVariable{}).Update(context.Background(), infer.UpdateRequest[RepositoryActionVariableArgs, RepositoryActionVariableState]{Inputs: inputs, State: state, DryRun: true})
	if err != nil {
		t.Fatalf("update dry-run: %v", err)
	}

	assertEqual(t, resp.Output.RepositoryActionVariableArgs, inputs)
	assertEqual(t, resp.Output.OwnerID, int64(10))
	assertEqual(t, resp.Output.RepoID, int64(20))
}

func TestRepositoryTagProtectionCreateDryRunBuildsPreviewState(t *testing.T) {
	t.Parallel()

	args := RepositoryTagProtectionArgs{Owner: "sironheart", Repository: "infra", NamePattern: "v*", WhitelistUsernames: []string{"alice"}, WhitelistTeams: []string{"release"}}
	resp, err := (RepositoryTagProtection{}).Create(context.Background(), infer.CreateRequest[RepositoryTagProtectionArgs]{Inputs: args, DryRun: true})
	if err != nil {
		t.Fatalf("create dry-run: %v", err)
	}

	assertEqual(t, resp.ID, "sironheart/infra/v*")
	assertEqual(t, resp.Output.RepositoryTagProtectionArgs, args)
	assertEqual(t, resp.Output.ProtectionID, int64(0))
}

func TestRepositoryTagProtectionUpdateDryRunPreservesProtectionID(t *testing.T) {
	t.Parallel()

	inputs := RepositoryTagProtectionArgs{Owner: "sironheart", Repository: "infra", NamePattern: "v*", WhitelistUsernames: []string{"alice", "bob"}}
	state := RepositoryTagProtectionState{RepositoryTagProtectionArgs: RepositoryTagProtectionArgs{Owner: "sironheart", Repository: "infra", NamePattern: "v*", WhitelistUsernames: []string{"alice"}}, ProtectionID: 30}
	resp, err := (RepositoryTagProtection{}).Update(context.Background(), infer.UpdateRequest[RepositoryTagProtectionArgs, RepositoryTagProtectionState]{Inputs: inputs, State: state, DryRun: true})
	if err != nil {
		t.Fatalf("update dry-run: %v", err)
	}

	assertEqual(t, resp.Output.RepositoryTagProtectionArgs, inputs)
	assertEqual(t, resp.Output.ProtectionID, int64(30))
}

func TestRepositoryPushMirrorCreateDryRunBuildsPreviewState(t *testing.T) {
	t.Parallel()

	args := RepositoryPushMirrorArgs{
		Owner:          "sironheart",
		Repository:     "infra",
		RemoteAddress:  "https://example.test/sironheart/infra.git",
		RemoteUsername: "mirror",
		RemotePassword: "secret",
		Interval:       "8h30m0s",
		BranchFilter:   "main",
		SyncOnCommit:   true,
		UseSSH:         true,
	}
	resp, err := (RepositoryPushMirror{}).Create(context.Background(), infer.CreateRequest[RepositoryPushMirrorArgs]{Inputs: args, DryRun: true})
	if err != nil {
		t.Fatalf("create dry-run: %v", err)
	}

	assertEqual(t, resp.ID, "sironheart/infra/https://example.test/sironheart/infra.git")
	assertEqual(t, resp.Output.RepositoryPushMirrorArgs, args)
	assertEqual(t, resp.Output.RemoteName, "")
	assertEqual(t, resp.Output.PublicKey, "")
	assertEqual(t, resp.Output.LastError, "")
}

func TestRepositoryActionSecretCreateDryRunBuildsPreviewState(t *testing.T) {
	t.Parallel()

	args := RepositoryActionSecretArgs{ActionSecretArgs: ActionSecretArgs{Name: "DEPLOY_TOKEN", Value: "secret"}, Owner: "sironheart", Repository: "infra"}
	resp, err := (RepositoryActionSecret{}).Create(context.Background(), infer.CreateRequest[RepositoryActionSecretArgs]{Inputs: args, DryRun: true})
	if err != nil {
		t.Fatalf("create dry-run: %v", err)
	}

	assertEqual(t, resp.ID, "sironheart/infra/DEPLOY_TOKEN")
	assertEqual(t, resp.Output.RepositoryActionSecretArgs, args)
	assertEqual(t, resp.Output.Created, "")
}

func TestOrganizationActionSecretCreateDryRunBuildsPreviewState(t *testing.T) {
	t.Parallel()

	args := OrganizationActionSecretArgs{ActionSecretArgs: ActionSecretArgs{Name: "DEPLOY_TOKEN", Value: "secret"}, Organization: "platform"}
	resp, err := (OrganizationActionSecret{}).Create(context.Background(), infer.CreateRequest[OrganizationActionSecretArgs]{Inputs: args, DryRun: true})
	if err != nil {
		t.Fatalf("create dry-run: %v", err)
	}

	assertEqual(t, resp.ID, "platform/DEPLOY_TOKEN")
	assertEqual(t, resp.Output.OrganizationActionSecretArgs, args)
	assertEqual(t, resp.Output.Created, "")
}

func TestOrganizationActionVariableCreateDryRunBuildsPreviewState(t *testing.T) {
	t.Parallel()

	args := OrganizationActionVariableArgs{ActionVariableArgs: ActionVariableArgs{Name: "ENVIRONMENT", Value: "prod"}, Organization: "platform"}
	resp, err := (OrganizationActionVariable{}).Create(context.Background(), infer.CreateRequest[OrganizationActionVariableArgs]{Inputs: args, DryRun: true})
	if err != nil {
		t.Fatalf("create dry-run: %v", err)
	}

	assertEqual(t, resp.ID, "platform/ENVIRONMENT")
	assertEqual(t, resp.Output.OrganizationActionVariableArgs, args)
	assertEqual(t, resp.Output.OwnerID, int64(0))
}

func TestRepositoryBranchProtectionCreateDryRunBuildsPreviewState(t *testing.T) {
	t.Parallel()

	args := RepositoryBranchProtectionArgs{
		Owner:                   "sironheart",
		Repository:              "infra",
		Name:                    "main",
		EnableStatusCheck:       true,
		StatusCheckContexts:     []string{"test"},
		RequiredApprovals:       2,
		BlockOnRejectedReviews:  true,
		RequireSignedCommits:    true,
		ProtectedFilePatterns:   "*.go",
		UnprotectedFilePatterns: "docs/**",
	}
	resp, err := (RepositoryBranchProtection{}).Create(context.Background(), infer.CreateRequest[RepositoryBranchProtectionArgs]{Inputs: args, DryRun: true})
	if err != nil {
		t.Fatalf("create dry-run: %v", err)
	}

	assertEqual(t, resp.ID, "sironheart/infra/main")
	assertEqual(t, resp.Output.RepositoryBranchProtectionArgs, args)
	assertEqual(t, resp.Output.Created, "")
	assertEqual(t, resp.Output.Updated, "")
}

func TestRepositorySettingsCreateDryRunBuildsPreviewState(t *testing.T) {
	t.Parallel()

	args := RepositorySettingsArgs{
		Owner:      "sironheart",
		Repository: "infra",
		RepositorySettingsConfig: RepositorySettingsConfig{
			Issues:                                 boolPtr(true),
			PullRequests:                           boolPtr(true),
			DefaultDeleteBranchAfterMerge:          boolPtr(true),
			Wiki:                                   boolPtr(true),
			Actions:                                boolPtr(true),
			ExternalWikiURL:                        stringPtr("https://wiki.example.test/infra"),
			ExternalTrackerURL:                     stringPtr("https://issues.example.test"),
			ExternalTrackerFormat:                  stringPtr("https://issues.example.test/{index}"),
			InternalTrackerEnableIssueDependencies: boolPtr(true),
		},
	}
	resp, err := (RepositorySettings{}).Create(context.Background(), infer.CreateRequest[RepositorySettingsArgs]{Inputs: args, DryRun: true})
	if err != nil {
		t.Fatalf("create dry-run: %v", err)
	}

	assertEqual(t, resp.ID, "sironheart/infra")
	assertEqual(t, resp.Output.RepositorySettingsArgs, args)
}

func assertEqual[T any](t *testing.T, got, want T) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %#v, got %#v", want, got)
	}
}
