package provider

import (
	"context"
	"testing"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

func TestPublicKeyDiffReplacesKeyMaterial(t *testing.T) {
	t.Parallel()

	resp, err := (PublicKey{}).Diff(context.Background(), infer.DiffRequest[PublicKeyArgs, PublicKeyState]{
		State:  PublicKeyState{PublicKeyArgs: PublicKeyArgs{Title: "laptop", Key: "ssh-ed25519 old", ReadOnly: true}},
		Inputs: PublicKeyArgs{Title: "laptop", Key: "ssh-ed25519 new", ReadOnly: true},
	})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	assertDiffKind(t, resp, "key", p.UpdateReplace)
}

func TestOrganizationTeamDiffUpdatesUnitsMap(t *testing.T) {
	t.Parallel()

	resp, err := (OrganizationTeam{}).Diff(context.Background(), infer.DiffRequest[OrganizationTeamArgs, OrganizationTeamState]{
		State: OrganizationTeamState{OrganizationTeamArgs: OrganizationTeamArgs{
			Organization: "infra",
			Name:         "maintainers",
			Permission:   "read",
			UnitsMap:     map[string]string{"repo.code": "read"},
		}},
		Inputs: OrganizationTeamArgs{
			Organization: "infra",
			Name:         "maintainers",
			Permission:   "read",
			UnitsMap:     map[string]string{"repo.code": "write"},
		},
	})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	assertDiffKind(t, resp, "unitsMap", p.Update)
}

func TestRepositoryDiffUpdatesNestedSettings(t *testing.T) {
	t.Parallel()

	resp, err := (Repository{}).Diff(context.Background(), infer.DiffRequest[RepositoryArgs, RepositoryState]{
		State: RepositoryState{RepositoryArgs: RepositoryArgs{
			Name:     "infra",
			Owner:    "sironheart",
			Settings: &RepositorySettingsConfig{Actions: boolPtr(false)},
		}},
		Inputs: RepositoryArgs{
			Name:     "infra",
			Owner:    "sironheart",
			Settings: &RepositorySettingsConfig{Actions: boolPtr(true)},
		},
	})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	assertDiffKind(t, resp, "settings", p.Update)
}

func TestOrganizationTeamMemberDiffReplacesUsername(t *testing.T) {
	t.Parallel()

	resp, err := (OrganizationTeamMember{}).Diff(context.Background(), infer.DiffRequest[OrganizationTeamMemberArgs, OrganizationTeamMemberState]{
		State:  OrganizationTeamMemberState{OrganizationTeamMemberArgs: OrganizationTeamMemberArgs{Organization: "platform", Team: "maintainers", TeamID: 42, Username: "alice"}},
		Inputs: OrganizationTeamMemberArgs{Organization: "platform", Team: "maintainers", Username: "bob"},
	})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	assertDiffKind(t, resp, "username", p.UpdateReplace)
}

func TestOrganizationTeamMemberDiffIgnoresResolvedTeamID(t *testing.T) {
	t.Parallel()

	resp, err := (OrganizationTeamMember{}).Diff(context.Background(), infer.DiffRequest[OrganizationTeamMemberArgs, OrganizationTeamMemberState]{
		State:  OrganizationTeamMemberState{OrganizationTeamMemberArgs: OrganizationTeamMemberArgs{Organization: "platform", Team: "maintainers", TeamID: 42, Username: "alice"}},
		Inputs: OrganizationTeamMemberArgs{Organization: "platform", Team: "maintainers", Username: "alice"},
	})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	if resp.HasChanges {
		t.Fatalf("expected no changes, got %#v", resp.DetailedDiff)
	}
}

func TestRepositoryTagProtectionDiffUpdatesWhitelist(t *testing.T) {
	t.Parallel()

	resp, err := (RepositoryTagProtection{}).Diff(context.Background(), infer.DiffRequest[RepositoryTagProtectionArgs, RepositoryTagProtectionState]{
		State: RepositoryTagProtectionState{RepositoryTagProtectionArgs: RepositoryTagProtectionArgs{
			Owner:              "sironheart",
			Repository:         "infra",
			NamePattern:        "v*",
			WhitelistUsernames: []string{"alice"},
		}},
		Inputs: RepositoryTagProtectionArgs{
			Owner:              "sironheart",
			Repository:         "infra",
			NamePattern:        "v*",
			WhitelistUsernames: []string{"alice", "bob"},
		},
	})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	assertDiffKind(t, resp, "whitelistUsernames", p.Update)
}

func TestRepositoryActionVariableDiffUpdatesValue(t *testing.T) {
	t.Parallel()

	resp, err := (RepositoryActionVariable{}).Diff(context.Background(), infer.DiffRequest[RepositoryActionVariableArgs, RepositoryActionVariableState]{
		State: RepositoryActionVariableState{RepositoryActionVariableArgs: RepositoryActionVariableArgs{
			ActionVariableArgs: ActionVariableArgs{Name: "ENVIRONMENT", Value: "dev"},
			Owner:              "sironheart",
			Repository:         "infra",
		}},
		Inputs: RepositoryActionVariableArgs{
			ActionVariableArgs: ActionVariableArgs{Name: "ENVIRONMENT", Value: "prod"},
			Owner:              "sironheart",
			Repository:         "infra",
		},
	})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	assertDiffKind(t, resp, "value", p.Update)
}

func TestRepositoryPushMirrorDiffReplacesBranchFilter(t *testing.T) {
	t.Parallel()

	resp, err := (RepositoryPushMirror{}).Diff(context.Background(), infer.DiffRequest[RepositoryPushMirrorArgs, RepositoryPushMirrorState]{
		State: RepositoryPushMirrorState{RepositoryPushMirrorArgs: RepositoryPushMirrorArgs{
			Owner:         "sironheart",
			Repository:    "infra",
			RemoteAddress: "https://example.test/sironheart/infra.git",
			BranchFilter:  "main",
		}},
		Inputs: RepositoryPushMirrorArgs{
			Owner:         "sironheart",
			Repository:    "infra",
			RemoteAddress: "https://example.test/sironheart/infra.git",
			BranchFilter:  "release/*",
		},
	})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	assertDiffKind(t, resp, "branchFilter", p.UpdateReplace)
}

func TestRepositoryActionSecretDiffUpdatesValue(t *testing.T) {
	t.Parallel()

	resp, err := (RepositoryActionSecret{}).Diff(context.Background(), infer.DiffRequest[RepositoryActionSecretArgs, RepositoryActionSecretState]{
		State: RepositoryActionSecretState{RepositoryActionSecretArgs: RepositoryActionSecretArgs{
			ActionSecretArgs: ActionSecretArgs{Name: "DEPLOY_TOKEN", Value: "old"},
			Owner:            "sironheart",
			Repository:       "infra",
		}},
		Inputs: RepositoryActionSecretArgs{
			ActionSecretArgs: ActionSecretArgs{Name: "DEPLOY_TOKEN", Value: "new"},
			Owner:            "sironheart",
			Repository:       "infra",
		},
	})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	assertDiffKind(t, resp, "value", p.Update)
}

func TestOrganizationActionVariableDiffUpdatesValue(t *testing.T) {
	t.Parallel()

	resp, err := (OrganizationActionVariable{}).Diff(context.Background(), infer.DiffRequest[OrganizationActionVariableArgs, OrganizationActionVariableState]{
		State: OrganizationActionVariableState{OrganizationActionVariableArgs: OrganizationActionVariableArgs{
			ActionVariableArgs: ActionVariableArgs{Name: "ENVIRONMENT", Value: "dev"},
			Organization:       "platform",
		}},
		Inputs: OrganizationActionVariableArgs{
			ActionVariableArgs: ActionVariableArgs{Name: "ENVIRONMENT", Value: "prod"},
			Organization:       "platform",
		},
	})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	assertDiffKind(t, resp, "value", p.Update)
}

func TestRepositoryBranchProtectionDiffUpdatesStatusChecks(t *testing.T) {
	t.Parallel()

	resp, err := (RepositoryBranchProtection{}).Diff(context.Background(), infer.DiffRequest[RepositoryBranchProtectionArgs, RepositoryBranchProtectionState]{
		State: RepositoryBranchProtectionState{RepositoryBranchProtectionArgs: RepositoryBranchProtectionArgs{
			Owner:               "sironheart",
			Repository:          "infra",
			Name:                "main",
			EnableStatusCheck:   true,
			StatusCheckContexts: []string{"test"},
		}},
		Inputs: RepositoryBranchProtectionArgs{
			Owner:               "sironheart",
			Repository:          "infra",
			Name:                "main",
			EnableStatusCheck:   true,
			StatusCheckContexts: []string{"test", "lint"},
		},
	})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	assertDiffKind(t, resp, "statusCheckContexts", p.Update)
}

func TestRepositorySettingsDiffUpdatesExternalTrackerURL(t *testing.T) {
	t.Parallel()

	resp, err := (RepositorySettings{}).Diff(context.Background(), infer.DiffRequest[RepositorySettingsArgs, RepositorySettingsState]{
		State: RepositorySettingsState{RepositorySettingsArgs: RepositorySettingsArgs{
			Owner:      "sironheart",
			Repository: "infra",
			RepositorySettingsConfig: RepositorySettingsConfig{
				ExternalTrackerURL: stringPtr("https://old.example.test"),
			},
		}},
		Inputs: RepositorySettingsArgs{
			Owner:      "sironheart",
			Repository: "infra",
			RepositorySettingsConfig: RepositorySettingsConfig{
				ExternalTrackerURL: stringPtr("https://new.example.test"),
			},
		},
	})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	assertDiffKind(t, resp, "externalTrackerUrl", p.Update)
}

func TestRepositorySettingsDiffUpdatesDefaultDeleteBranchAfterMerge(t *testing.T) {
	t.Parallel()

	resp, err := (RepositorySettings{}).Diff(context.Background(), infer.DiffRequest[RepositorySettingsArgs, RepositorySettingsState]{
		State: RepositorySettingsState{RepositorySettingsArgs: RepositorySettingsArgs{
			Owner:      "sironheart",
			Repository: "infra",
			RepositorySettingsConfig: RepositorySettingsConfig{
				DefaultDeleteBranchAfterMerge: boolPtr(false),
			},
		}},
		Inputs: RepositorySettingsArgs{
			Owner:      "sironheart",
			Repository: "infra",
			RepositorySettingsConfig: RepositorySettingsConfig{
				DefaultDeleteBranchAfterMerge: boolPtr(true),
			},
		},
	})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	assertDiffKind(t, resp, "defaultDeleteBranchAfterMerge", p.Update)
}

func TestRepositorySettingsEditOptionEnablesPullRequestsForDefaultBranchDeletion(t *testing.T) {
	t.Parallel()

	opt := repositorySettingsEditOption(RepositorySettingsArgs{RepositorySettingsConfig: RepositorySettingsConfig{
		DefaultDeleteBranchAfterMerge: boolPtr(true),
	}}, nil)

	assertBoolPtr(t, opt.DefaultDeleteBranchAfterMerge, true)
	assertBoolPtr(t, opt.HasPullRequests, true)
}

func TestRepositorySettingsEditOptionKeepsExplicitPullRequests(t *testing.T) {
	t.Parallel()

	opt := repositorySettingsEditOption(RepositorySettingsArgs{RepositorySettingsConfig: RepositorySettingsConfig{
		PullRequests:                  boolPtr(false),
		DefaultDeleteBranchAfterMerge: boolPtr(true),
	}}, nil)

	assertBoolPtr(t, opt.DefaultDeleteBranchAfterMerge, true)
	assertBoolPtr(t, opt.HasPullRequests, false)
}

func TestRepositorySettingsDiffUpdatesArchived(t *testing.T) {
	t.Parallel()

	resp, err := (RepositorySettings{}).Diff(context.Background(), infer.DiffRequest[RepositorySettingsArgs, RepositorySettingsState]{
		State: RepositorySettingsState{RepositorySettingsArgs: RepositorySettingsArgs{
			Owner:      "sironheart",
			Repository: "infra",
			RepositorySettingsConfig: RepositorySettingsConfig{
				Archived: boolPtr(false),
			},
		}},
		Inputs: RepositorySettingsArgs{
			Owner:      "sironheart",
			Repository: "infra",
			RepositorySettingsConfig: RepositorySettingsConfig{
				Archived: boolPtr(true),
			},
		},
	})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	assertDiffKind(t, resp, "archived", p.Update)
}

func TestRepositorySettingsEditOptionPropagatesArchived(t *testing.T) {
	t.Parallel()

	opt := repositorySettingsEditOption(RepositorySettingsArgs{RepositorySettingsConfig: RepositorySettingsConfig{
		Archived: boolPtr(true),
	}}, nil)

	assertBoolPtr(t, opt.Archived, true)
}

func assertDiffKind(t *testing.T, resp infer.DiffResponse, property string, kind p.DiffKind) {
	t.Helper()
	if !resp.HasChanges {
		t.Fatalf("expected diff for %s", property)
	}
	diff, ok := resp.DetailedDiff[property]
	if !ok {
		t.Fatalf("expected diff for %s, got %#v", property, resp.DetailedDiff)
	}
	if diff.Kind != kind {
		t.Fatalf("expected %s diff kind %v, got %v", property, kind, diff.Kind)
	}
}

func assertBoolPtr(t *testing.T, got *bool, want bool) {
	t.Helper()
	if got == nil || *got != want {
		t.Fatalf("expected bool pointer %v, got %v", want, derefBoolPtr(got))
	}
}

func derefBoolPtr(p *bool) any {
	if p == nil {
		return nil
	}
	return *p
}
