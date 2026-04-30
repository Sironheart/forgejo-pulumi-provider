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
			Owner:      "sironheart",
			Repository: "infra",
			Name:       "ENVIRONMENT",
			Value:      "dev",
		}},
		Inputs: RepositoryActionVariableArgs{
			Owner:      "sironheart",
			Repository: "infra",
			Name:       "ENVIRONMENT",
			Value:      "prod",
		},
	})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	assertDiffKind(t, resp, "value", p.Update)
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
