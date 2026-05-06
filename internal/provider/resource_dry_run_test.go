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
		Issues:        true,
		Wiki:          true,
		Projects:      true,
		Template:      true,
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

	inputs := RepositoryArgs{Name: "infra", Owner: "sironheart", Description: "updated", Issues: true, Wiki: true, Projects: true}
	state := RepositoryState{
		RepositoryArgs: RepositoryArgs{Name: "infra", Owner: "sironheart", Description: "old", Issues: true, Wiki: true, Projects: true},
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

	args := RepositoryActionVariableArgs{Owner: "sironheart", Repository: "infra", Name: "ENVIRONMENT", Value: "prod"}
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

	inputs := RepositoryActionVariableArgs{Owner: "sironheart", Repository: "infra", Name: "ENVIRONMENT", Value: "prod"}
	state := RepositoryActionVariableState{RepositoryActionVariableArgs: RepositoryActionVariableArgs{Owner: "sironheart", Repository: "infra", Name: "ENVIRONMENT", Value: "dev"}, OwnerID: 10, RepoID: 20}
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

func assertEqual[T any](t *testing.T, got, want T) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %#v, got %#v", want, got)
	}
}
