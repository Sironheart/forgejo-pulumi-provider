package provider

import (
	"context"
	"testing"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
)

func TestRepositoryDiffIgnoresComputedOwnerAndDefaultBranch(t *testing.T) {
	t.Parallel()

	resp, err := (Repository{}).Diff(context.Background(), infer.DiffRequest[RepositoryArgs, RepositoryState]{
		State: RepositoryState{RepositoryArgs: RepositoryArgs{
			Name:          "infra",
			Owner:         "alice",
			DefaultBranch: "main",
		}},
		Inputs: RepositoryArgs{
			Name: "infra",
		},
	})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	if resp.HasChanges {
		t.Fatalf("expected no diff, got %#v", resp.DetailedDiff)
	}
}

func TestRepositoryDiffMigratesLegacyUnitState(t *testing.T) {
	t.Parallel()

	resp, err := Provider().Diff(context.Background(), p.DiffRequest{
		ID:  "sironheart/infra",
		Urn: resource.URN("urn:pulumi:dev::git-manager::forgejo:index:Repository::infra"),
		State: property.NewMap(map[string]property.Value{
			"name":     property.New("infra"),
			"owner":    property.New("sironheart"),
			"issues":   property.New(true),
			"wiki":     property.New(true),
			"projects": property.New(false),
			"fullName": property.New("sironheart/infra"),
			"htmlUrl":  property.New("https://forgejo.test/sironheart/infra"),
			"sshUrl":   property.New("ssh://git@forgejo.test/sironheart/infra.git"),
			"cloneUrl": property.New("https://forgejo.test/sironheart/infra.git"),
		}),
		Inputs: property.NewMap(map[string]property.Value{
			"name":  property.New("infra"),
			"owner": property.New("sironheart"),
			"settings": property.New(map[string]property.Value{
				"issues":   property.New(true),
				"wiki":     property.New(true),
				"projects": property.New(false),
			}),
		}),
	})
	if err != nil {
		t.Fatalf("diff legacy state: %v", err)
	}
	if resp.HasChanges {
		t.Fatalf("expected no diff, got %#v", resp.DetailedDiff)
	}
}
