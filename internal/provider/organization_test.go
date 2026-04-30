package provider

import (
	"context"
	"testing"

	"github.com/pulumi/pulumi-go-provider/infer"
)

func TestOrganizationDiffIgnoresComputedVisibility(t *testing.T) {
	t.Parallel()

	resp, err := (Organization{}).Diff(context.Background(), infer.DiffRequest[OrganizationArgs, OrganizationState]{
		State:  OrganizationState{OrganizationArgs: OrganizationArgs{Name: "team", Visibility: "public"}},
		Inputs: OrganizationArgs{Name: "team"},
	})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	if resp.HasChanges {
		t.Fatalf("expected no diff, got %#v", resp.DetailedDiff)
	}
}
