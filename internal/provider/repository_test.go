package provider

import (
	"context"
	"testing"

	"github.com/pulumi/pulumi-go-provider/infer"
)

func TestRepositoryDiffIgnoresComputedOwnerAndDefaultBranch(t *testing.T) {
	t.Parallel()

	resp, err := (Repository{}).Diff(context.Background(), infer.DiffRequest[RepositoryArgs, RepositoryState]{
		State: RepositoryState{RepositoryArgs: RepositoryArgs{
			Name:          "infra",
			Owner:         "alice",
			DefaultBranch: "main",
			Issues:        true,
			Wiki:          true,
			Projects:      true,
		}},
		Inputs: RepositoryArgs{
			Name:     "infra",
			Issues:   true,
			Wiki:     true,
			Projects: true,
		},
	})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	if resp.HasChanges {
		t.Fatalf("expected no diff, got %#v", resp.DetailedDiff)
	}
}
