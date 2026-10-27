package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

func TestAncillaryOrganizationAPI(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		email  *string
		access *bool
	}{
		{name: "omitted"},
		{name: "clear", email: stringPtr(""), access: boolPtr(false)},
		{name: "set", email: stringPtr("team@example.test"), access: boolPtr(true)},
		{name: "email-only", email: stringPtr("team@example.test")},
		{name: "access-only", access: boolPtr(false)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args := OrganizationArgs{
				Name: "platform", FullName: "Platform", Description: "Team org", Website: "https://example.test", Location: "Remote", Visibility: "private",
				Email: tc.email, RepoAdminChangeTeamAccess: tc.access,
			}
			remote := map[string]any{
				"name": "platform", "username": "legacy", "full_name": "Platform", "description": "Team org", "website": "https://example.test", "location": "Remote", "visibility": "private",
				"avatar_url": "https://example.test/avatar", "email": "unmanaged@example.test", "repo_admin_change_team_access": true,
			}
			if tc.email != nil {
				remote["email"] = *tc.email
			}
			if tc.access != nil {
				remote["repo_admin_change_team_access"] = *tc.access
			}
			if tc.name == "omitted" {
				delete(remote, "name")
				remote["username"] = "platform"
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "token test-token" {
					t.Error("missing API authentication")
				}
				path := "/forgejo/api/v1/orgs/platform"
				if r.Method == http.MethodPost {
					path = "/forgejo/api/v1/orgs"
				}
				if r.URL.Path != path {
					t.Errorf("path = %q, want %q", r.URL.Path, path)
				}
				if r.Method != http.MethodGet {
					var body map[string]any
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
					}
					want := map[string]any{"full_name": "Platform", "description": "Team org", "website": "https://example.test", "location": "Remote", "visibility": "private"}
					if r.Method == http.MethodPost {
						want["username"] = "platform"
					} else if r.Method != http.MethodPatch {
						t.Errorf("unexpected method %s", r.Method)
					}
					if tc.email != nil {
						want["email"] = *tc.email
					}
					if tc.access != nil {
						want["repo_admin_change_team_access"] = *tc.access
					}
					if !reflect.DeepEqual(body, want) {
						t.Errorf("%s payload = %#v, want %#v", r.Method, body, want)
					}
				}
				ancillaryJSON(t, w, remote)
			}))
			defer server.Close()

			preview, err := (Organization{}).Create(context.Background(), infer.CreateRequest[OrganizationArgs]{Inputs: args, DryRun: true})
			if err != nil {
				t.Fatal(err)
			}
			assertEqual(t, preview.Output.OrganizationArgs, args)
			withAncillaryConfig(t, server.URL+"/forgejo/api/v1/", func(ctx context.Context) {
				created, err := (Organization{}).Create(ctx, infer.CreateRequest[OrganizationArgs]{Inputs: args})
				if err != nil {
					t.Fatal(err)
				}
				assertEqual(t, created.ID, "platform")
				assertEqual(t, created.Output.OrganizationArgs, args)
				stale := created.Output
				if tc.email != nil {
					stale.Email = stringPtr("old@example.test")
				}
				if tc.access != nil {
					stale.RepoAdminChangeTeamAccess = boolPtr(!*tc.access)
				}
				read, err := (Organization{}).Read(ctx, infer.ReadRequest[OrganizationArgs, OrganizationState]{ID: created.ID, Inputs: args, State: stale})
				if err != nil {
					t.Fatal(err)
				}
				assertEqual(t, read.ID, created.ID)
				assertEqual(t, read.Inputs, args)
				assertEqual(t, read.State, created.Output)
				stale.Email = stringPtr("previously-managed@example.test")
				stale.RepoAdminChangeTeamAccess = boolPtr(true)
				for _, dryRun := range []bool{true, false} {
					updated, err := (Organization{}).Update(ctx, infer.UpdateRequest[OrganizationArgs, OrganizationState]{ID: created.ID, Inputs: args, State: stale, DryRun: dryRun})
					if err != nil {
						t.Fatal(err)
					}
					assertEqual(t, updated.Output, created.Output)
				}
			})
		})
	}
}

func TestAncillaryOrganizationOptionalDiff(t *testing.T) {
	t.Parallel()
	state := OrganizationState{OrganizationArgs: OrganizationArgs{Name: "platform", Email: stringPtr("old@example.test"), RepoAdminChangeTeamAccess: boolPtr(true)}}
	for _, args := range []OrganizationArgs{
		{Name: "platform"},
		{Name: "platform", Email: stringPtr("old@example.test"), RepoAdminChangeTeamAccess: boolPtr(true)},
	} {
		diff, err := (Organization{}).Diff(context.Background(), infer.DiffRequest[OrganizationArgs, OrganizationState]{State: state, Inputs: args})
		if err != nil || diff.HasChanges {
			t.Fatalf("unexpected optional diff: %#v, %v", diff.DetailedDiff, err)
		}
	}
	diff, err := (Organization{}).Diff(context.Background(), infer.DiffRequest[OrganizationArgs, OrganizationState]{
		State: state, Inputs: OrganizationArgs{Name: "platform", Email: stringPtr(""), RepoAdminChangeTeamAccess: boolPtr(false)},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertDiffKind(t, diff, "email", p.Update)
	assertDiffKind(t, diff, "repoAdminChangeTeamAccess", p.Update)
}

func TestAncillaryOrganizationUpdateNotFound(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if (r.Method != http.MethodPatch && r.Method != http.MethodGet) || r.URL.Path != "/api/v1/orgs/platform" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
		}
		w.WriteHeader(http.StatusNotFound)
		ancillaryJSON(t, w, map[string]string{"message": "organization missing"})
	}))
	defer server.Close()

	withAncillaryConfig(t, server.URL, func(ctx context.Context) {
		_, err := (Organization{}).Update(ctx, infer.UpdateRequest[OrganizationArgs, OrganizationState]{
			ID: "platform", Inputs: OrganizationArgs{Name: "platform"},
		})
		if err == nil || !strings.Contains(err.Error(), "organization missing") {
			t.Fatalf("expected update 404 error, got %v", err)
		}
		read, err := (Organization{}).Read(ctx, infer.ReadRequest[OrganizationArgs, OrganizationState]{ID: "platform"})
		if err != nil || read.ID != "" {
			t.Fatalf("expected missing organization on refresh, got %#v, %v", read, err)
		}
	})
}

func TestAncillarySecretLifecycle(t *testing.T) {
	t.Parallel()
	for _, scope := range []string{"repository", "organization"} {
		t.Run(scope, func(t *testing.T) {
			path := "/api/v1/orgs/platform/actions/secrets"
			if scope == "repository" {
				path = "/api/v1/repos/platform/infra/actions/secrets"
			}
			const created = "2026-10-01T12:00:00Z"
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					if r.URL.Path != path || r.URL.Query().Get("page") != "0" || r.URL.Query().Get("limit") != "0" {
						t.Errorf("unexpected all-secrets lookup: %s", r.URL)
					}
					ancillaryJSON(t, w, []map[string]string{{"name": "OTHER"}, {"name": "DEPLOY_TOKEN", "created_at": created}})
				case http.MethodPut:
					var body map[string]string
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
					}
					wantValue := "old-value"
					if body["name"] == "deploy_token" {
						wantValue = "new-value"
					}
					if r.URL.Path != path+"/"+body["name"] || body["data"] != wantValue {
						t.Error("incorrect secret write payload or path")
					}
					w.WriteHeader(http.StatusNoContent)
				default:
					t.Errorf("unexpected method: %s", r.Method)
					w.WriteHeader(http.StatusMethodNotAllowed)
				}
			}))
			defer server.Close()
			withAncillaryConfig(t, server.URL, func(ctx context.Context) {
				secret := ActionSecretArgs{Name: "Deploy_Token", Value: "old-value"}
				if scope == "repository" {
					args := RepositoryActionSecretArgs{Owner: "platform", Repository: "infra", ActionSecretArgs: secret}
					for _, dryRun := range []bool{true, false} {
						resp, err := (RepositoryActionSecret{}).Create(ctx, infer.CreateRequest[RepositoryActionSecretArgs]{Inputs: args, DryRun: dryRun})
						if err != nil {
							t.Fatal(err)
						}
						assertEqual(t, resp.ID, "platform/infra/DEPLOY_TOKEN")
						assertEqual(t, resp.Output.Name, "DEPLOY_TOKEN")
						assertEqual(t, resp.Output.Value, secret.Value)
						if !dryRun {
							assertEqual(t, resp.Output.Created, created)
						}
					}
					state := RepositoryActionSecretState{RepositoryActionSecretArgs: args, Created: created}
					args.Name, args.Value = "deploy_token", "new-value"
					for _, dryRun := range []bool{true, false} {
						resp, err := (RepositoryActionSecret{}).Update(ctx, infer.UpdateRequest[RepositoryActionSecretArgs, RepositoryActionSecretState]{Inputs: args, State: state, DryRun: dryRun})
						if err != nil {
							t.Fatal(err)
						}
						assertEqual(t, resp.Output.Name, "DEPLOY_TOKEN")
						assertEqual(t, resp.Output.Value, args.Value)
						assertEqual(t, resp.Output.Created, created)
					}
					read, err := (RepositoryActionSecret{}).Read(ctx, infer.ReadRequest[RepositoryActionSecretArgs, RepositoryActionSecretState]{State: state})
					if err != nil {
						t.Fatal(err)
					}
					assertEqual(t, read.ID, "platform/infra/DEPLOY_TOKEN")
					assertEqual(t, read.State.Name, "DEPLOY_TOKEN")
					assertEqual(t, read.Inputs.Value, secret.Value)
					assertEqual(t, read.State.Value, secret.Value)
				} else {
					args := OrganizationActionSecretArgs{Organization: "platform", ActionSecretArgs: secret}
					for _, dryRun := range []bool{true, false} {
						resp, err := (OrganizationActionSecret{}).Create(ctx, infer.CreateRequest[OrganizationActionSecretArgs]{Inputs: args, DryRun: dryRun})
						if err != nil {
							t.Fatal(err)
						}
						assertEqual(t, resp.ID, "platform/DEPLOY_TOKEN")
						assertEqual(t, resp.Output.Name, "DEPLOY_TOKEN")
						assertEqual(t, resp.Output.Value, secret.Value)
						if !dryRun {
							assertEqual(t, resp.Output.Created, created)
						}
					}
					state := OrganizationActionSecretState{OrganizationActionSecretArgs: args, Created: created}
					args.Name, args.Value = "deploy_token", "new-value"
					for _, dryRun := range []bool{true, false} {
						resp, err := (OrganizationActionSecret{}).Update(ctx, infer.UpdateRequest[OrganizationActionSecretArgs, OrganizationActionSecretState]{Inputs: args, State: state, DryRun: dryRun})
						if err != nil {
							t.Fatal(err)
						}
						assertEqual(t, resp.Output.Name, "DEPLOY_TOKEN")
						assertEqual(t, resp.Output.Value, args.Value)
						assertEqual(t, resp.Output.Created, created)
					}
					read, err := (OrganizationActionSecret{}).Read(ctx, infer.ReadRequest[OrganizationActionSecretArgs, OrganizationActionSecretState]{State: state})
					if err != nil {
						t.Fatal(err)
					}
					assertEqual(t, read.ID, "platform/DEPLOY_TOKEN")
					assertEqual(t, read.State.Name, "DEPLOY_TOKEN")
					assertEqual(t, read.Inputs.Value, secret.Value)
					assertEqual(t, read.State.Value, secret.Value)
				}
			})
		})
	}
}

func TestAncillarySecretCaseDiff(t *testing.T) {
	t.Parallel()
	old := ActionSecretArgs{Name: "DEPLOY_TOKEN", Value: "unchanged"}
	inputs := ActionSecretArgs{Name: "Deploy_Token", Value: old.Value}
	repo, err := (RepositoryActionSecret{}).Diff(context.Background(), infer.DiffRequest[RepositoryActionSecretArgs, RepositoryActionSecretState]{
		State:  RepositoryActionSecretState{RepositoryActionSecretArgs: RepositoryActionSecretArgs{ActionSecretArgs: old}},
		Inputs: RepositoryActionSecretArgs{ActionSecretArgs: inputs},
	})
	if err != nil || repo.HasChanges {
		t.Errorf("repository casing diff: %#v, %v", repo.DetailedDiff, err)
	}
	org, err := (OrganizationActionSecret{}).Diff(context.Background(), infer.DiffRequest[OrganizationActionSecretArgs, OrganizationActionSecretState]{
		State:  OrganizationActionSecretState{OrganizationActionSecretArgs: OrganizationActionSecretArgs{ActionSecretArgs: old}},
		Inputs: OrganizationActionSecretArgs{ActionSecretArgs: inputs},
	})
	if err != nil || org.HasChanges {
		t.Errorf("organization casing diff: %#v, %v", org.DetailedDiff, err)
	}
}

func TestAncillaryTeamOwnerAlias(t *testing.T) {
	t.Parallel()
	args := OrganizationTeamArgs{Organization: "platform", Name: "maintainers", Permission: "owner"}
	state := OrganizationTeamState{OrganizationTeamArgs: args}
	state.Permission = "admin"
	diff, err := (OrganizationTeam{}).Diff(context.Background(), infer.DiffRequest[OrganizationTeamArgs, OrganizationTeamState]{State: state, Inputs: args})
	if err != nil || diff.HasChanges {
		t.Fatalf("owner/admin alias diff: %#v, %v", diff.DetailedDiff, err)
	}
	preview, err := (OrganizationTeam{}).Create(context.Background(), infer.CreateRequest[OrganizationTeamArgs]{Inputs: args, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, preview.Output.Permission, "admin")
}

func TestAncillaryTeamUpdateLongName(t *testing.T) {
	t.Parallel()
	name := strings.Repeat("t", 255)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/teams/42" {
			t.Errorf("unexpected team path: %s", r.URL.Path)
		}
		if r.Method == http.MethodPatch {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			want := map[string]any{"name": name, "description": "", "permission": "admin", "can_create_org_repo": false, "includes_all_repositories": false, "units_map": nil, "units": nil}
			if !reflect.DeepEqual(body, want) {
				t.Errorf("team PATCH payload = %#v, want %#v", body, want)
			}
		} else if r.Method != http.MethodGet {
			t.Errorf("unexpected team method: %s", r.Method)
		}
		ancillaryJSON(t, w, map[string]any{"id": 42, "name": name, "permission": "admin"})
	}))
	defer server.Close()
	withAncillaryConfig(t, server.URL, func(ctx context.Context) {
		updated, err := (OrganizationTeam{}).Update(ctx, infer.UpdateRequest[OrganizationTeamArgs, OrganizationTeamState]{
			ID: "42", Inputs: OrganizationTeamArgs{Organization: "platform", Name: name, Permission: "owner"},
		})
		if err != nil {
			t.Fatal(err)
		}
		assertEqual(t, updated.Output.Name, name)
		assertEqual(t, updated.Output.Permission, "admin")
	})
}

func TestAncillarySDKValidationPreserved(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("invalid input reached API: %s %s", r.Method, r.URL)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()
	withAncillaryConfig(t, server.URL, func(ctx context.Context) {
		for _, args := range []OrganizationArgs{{}, {Name: "platform", Visibility: "invalid"}} {
			if _, err := (Organization{}).Create(ctx, infer.CreateRequest[OrganizationArgs]{Inputs: args}); err == nil {
				t.Errorf("accepted invalid organization: %#v", args)
			}
		}
		if _, err := (Organization{}).Update(ctx, infer.UpdateRequest[OrganizationArgs, OrganizationState]{ID: "platform", Inputs: OrganizationArgs{Visibility: "invalid"}}); err == nil {
			t.Error("accepted invalid organization visibility on update")
		}
		for _, args := range []OrganizationTeamArgs{
			{Permission: "read"},
			{Name: strings.Repeat("t", 256), Permission: "read"},
			{Name: "maintainers", Permission: "read", Description: strings.Repeat("d", 256)},
			{Name: "maintainers", Permission: "invalid"},
		} {
			if _, err := (OrganizationTeam{}).Update(ctx, infer.UpdateRequest[OrganizationTeamArgs, OrganizationTeamState]{ID: "42", Inputs: args}); err == nil {
				t.Errorf("accepted invalid team: %#v", args)
			}
		}
	})
}

func TestAncillaryForgejoURL(t *testing.T) {
	t.Parallel()
	for raw, want := range map[string]string{
		" https://forgejo.test/ ":               "https://forgejo.test",
		"http://localhost:3000/forgejo/api/v1/": "http://localhost:3000/forgejo",
		"https://forgejo.test/subpath/":         "https://forgejo.test/subpath",
	} {
		got, err := normalizeForgejoURL(raw)
		if err != nil || got != want {
			t.Errorf("normalize %q: got %q, %v; want %q", raw, got, err, want)
		}
	}
	for _, raw := range []string{
		"", "forgejo.test", "//forgejo.test", "https://:3000", "https://forgejo.test:invalid", "ftp://forgejo.test", "file:///forgejo",
		"https://forgejo.test?token=credential", "https://forgejo.test?", "https://forgejo.test/#credential", "https://forgejo.test/#",
		"https://user:credential@forgejo.test", "https://user:credential%zz@forgejo.test", "https://credential bad.test",
	} {
		_, err := normalizeForgejoURL(raw)
		if err == nil {
			t.Errorf("accepted invalid base URL %q", raw)
		} else if strings.Contains(err.Error(), "credential") {
			t.Errorf("URL validation exposed credentials: %v", err)
		}
		withAncillaryConfig(t, raw, func(ctx context.Context) {
			if _, err := clientFromConfig(ctx); err == nil || strings.Contains(err.Error(), "credential") {
				t.Errorf("SDK client URL validation: %v", err)
			}
			if _, err := forgejoAPIClientFromConfig(ctx); err == nil || strings.Contains(err.Error(), "credential") {
				t.Errorf("raw client URL validation: %v", err)
			}
		})
	}
}

func withAncillaryConfig(t *testing.T, baseURL string, run func(context.Context)) {
	t.Helper()
	provider := infer.Wrap(p.Provider{
		Invoke: func(ctx context.Context, _ p.InvokeRequest) (p.InvokeResponse, error) {
			run(ctx)
			return p.InvokeResponse{}, nil
		},
	}, infer.Options{Config: infer.Config(&Config{URL: baseURL, Token: "test-token"})})
	if _, err := provider.Invoke(context.Background(), p.InvokeRequest{}); err != nil {
		t.Fatal(err)
	}
}

func ancillaryJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Error(err)
	}
}
