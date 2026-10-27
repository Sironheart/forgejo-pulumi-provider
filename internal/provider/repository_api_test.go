package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
)

func TestRepositorySettingsAPIUpdateAndRefresh(t *testing.T) {
	t.Parallel()

	for _, resourceName := range []string{"Repository", "RepositorySettings"} {
		t.Run(resourceName, func(t *testing.T) {
			t.Parallel()
			settings := map[string]property.Value{
				"ignoreWhitespaceConflicts":     property.New(false),
				"allowManualMerge":              property.New(false),
				"autodetectManualMerge":         property.New(true),
				"allowRebaseUpdate":             property.New(true),
				"defaultUpdateStyle":            property.New("rebase"),
				"defaultAllowMaintainerEdit":    property.New(false),
				"mirrorInterval":                property.New("1h"),
				"enablePrune":                   property.New(false),
				"defaultDeleteBranchAfterMerge": property.New(false),
				"allowFastForwardOnlyMerge":     property.New(false),
				"globallyEditableWiki":          property.New(false),
				"wikiBranch":                    property.New("docs"),
			}
			wire := map[string]any{
				"ignore_whitespace_conflicts":       false,
				"allow_manual_merge":                false,
				"autodetect_manual_merge":           true,
				"allow_rebase_update":               true,
				"default_update_style":              "rebase",
				"default_allow_maintainer_edit":     false,
				"mirror_interval":                   "1h",
				"enable_prune":                      false,
				"default_delete_branch_after_merge": false,
				"allow_fast_forward_only_merge":     false,
				"globally_editable_wiki":            false,
				"wiki_branch":                       "docs",
			}
			var drift atomic.Bool
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/repos/alice/infra" {
					t.Errorf("unexpected path: %s", r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
					return
				}
				if r.Header.Get("Authorization") != "token test-token" {
					t.Error("missing API authentication")
				}
				if r.Method == http.MethodPatch {
					var body map[string]any
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
					}
					for key, want := range wire {
						if body[key] != want {
							t.Errorf("%s: got %v, want %v", key, body[key], want)
						}
					}
					if body["has_pull_requests"] != true {
						t.Error("merge settings must enable pull requests")
					}
					if _, ok := body["has_actions"]; ok {
						t.Error("unmanaged actions setting was sent")
					}
				} else if r.Method != http.MethodGet {
					t.Errorf("unexpected method: %s", r.Method)
				}
				response := map[string]any{
					"name": "infra", "owner": map[string]any{"login": "alice"}, "full_name": "alice/infra", "mirror": true,
				}
				for key, value := range wire {
					switch key {
					case "allow_manual_merge", "autodetect_manual_merge", "enable_prune":
						continue // These fields are genuinely absent from Forgejo responses.
					case "mirror_interval":
						value = "1h0m0s"
					case "wiki_branch":
						if drift.Load() {
							value = "other"
						}
					case "default_update_style":
						if drift.Load() {
							value = "merge"
						}
					default:
						if drift.Load() {
							value = !value.(bool)
						}
					}
					response[key] = value
				}
				if err := json.NewEncoder(w).Encode(response); err != nil {
					t.Error(err)
				}
			}))
			defer server.Close()
			prov := repositoryAPITestProvider(t, server.URL)
			urn := resource.URN("urn:pulumi:test::test::forgejo:index:" + resourceName + "::infra")
			props := map[string]property.Value{"owner": property.New("alice")}
			if resourceName == "Repository" {
				props["name"] = property.New("infra")
				props["settings"] = property.New(settings)
			} else {
				props["repository"] = property.New("infra")
				for key, value := range settings {
					props[key] = value
				}
			}
			inputs := property.NewMap(props)
			state := inputs
			if resourceName == "Repository" {
				for _, key := range []string{"fullName", "htmlUrl", "sshUrl", "cloneUrl"} {
					state = state.Set(key, property.New(""))
				}
			}
			updated, err := prov.Update(context.Background(), p.UpdateRequest{Urn: urn, ID: "alice/infra", State: state, Inputs: inputs})
			if err != nil {
				t.Fatalf("update: %v", err)
			}
			diff, err := prov.Diff(context.Background(), p.DiffRequest{Urn: urn, ID: "alice/infra", State: updated.Properties, Inputs: inputs})
			if err != nil || diff.HasChanges {
				t.Fatalf("unexpected diff after update: %#v, %v", diff, err)
			}
			drift.Store(true)
			refreshed, err := prov.Read(context.Background(), p.ReadRequest{Urn: urn, ID: "alice/infra", Properties: updated.Properties, Inputs: inputs})
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			observed := refreshed.Properties
			if resourceName == "Repository" {
				observed = observed.Get("settings").AsMap()
			}
			for _, key := range []string{"ignoreWhitespaceConflicts", "defaultDeleteBranchAfterMerge", "allowFastForwardOnlyMerge", "globallyEditableWiki", "defaultAllowMaintainerEdit"} {
				assertEqual(t, observed.Get(key).AsBool(), true)
			}
			assertEqual(t, observed.Get("allowRebaseUpdate").AsBool(), false)
			assertEqual(t, observed.Get("defaultUpdateStyle").AsString(), "merge")
			assertEqual(t, observed.Get("wikiBranch").AsString(), "other")
			assertEqual(t, observed.Get("allowManualMerge").AsBool(), false)
			assertEqual(t, observed.Get("autodetectManualMerge").AsBool(), true)
			assertEqual(t, observed.Get("enablePrune").AsBool(), false)
			diff, err = prov.Diff(context.Background(), p.DiffRequest{Urn: urn, ID: "alice/infra", State: refreshed.Properties, Inputs: inputs})
			if err != nil || !diff.HasChanges {
				t.Fatalf("missing drift: %#v, %v", diff, err)
			}
			for key, change := range diff.DetailedDiff {
				if change.Kind != p.Update {
					t.Errorf("%s should update, got %v", key, change.Kind)
				}
			}
		})
	}
}

func TestRepositoryCreationAPIOptions(t *testing.T) {
	t.Parallel()
	for _, owner := range []string{"", "alice", "platform"} {
		t.Run("owner="+owner, func(t *testing.T) {
			t.Parallel()
			resolvedOwner := owner
			if resolvedOwner == "" {
				resolvedOwner = "alice"
			}
			createPath := "/api/v1/user/repos"
			if owner == "platform" {
				createPath = "/api/v1/orgs/platform/repos"
			}
			creation := map[string]string{
				"gitignores": "Go,Node", "license": "MIT", "readme": "Default", "issue_labels": "Default",
				"trust_model": "committer", "object_format_name": "sha256",
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet && r.URL.Path == "/api/v1/user" {
					_, _ = w.Write([]byte(`{"login":"alice"}`))
					return
				}
				var body map[string]any
				if r.Method != http.MethodGet {
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
					}
				}
				if r.Method == http.MethodPost {
					if r.URL.Path != createPath || body["auto_init"] != true {
						t.Errorf("unexpected creation request: %s, %#v", r.URL.Path, body)
					}
					for key, want := range creation {
						if body[key] != want {
							t.Errorf("%s: got %v, want %s", key, body[key], want)
						}
					}
				} else {
					if r.URL.Path != "/api/v1/repos/"+resolvedOwner+"/infra" {
						t.Errorf("unexpected repository path: %s", r.URL.Path)
					}
					for key := range creation {
						if _, ok := body[key]; ok {
							t.Errorf("creation-only field %s sent on update", key)
						}
					}
				}
				if err := json.NewEncoder(w).Encode(map[string]any{
					"name": "infra", "owner": map[string]string{"login": resolvedOwner}, "object_format_name": "sha256",
				}); err != nil {
					t.Error(err)
				}
			}))
			defer server.Close()
			prov := repositoryAPITestProvider(t, server.URL)
			urn := resource.URN("urn:pulumi:test::test::forgejo:index:Repository::infra")
			inputs := property.NewMap(map[string]property.Value{
				"name": property.New("infra"), "owner": property.New(owner), "initializeWithReadme": property.New(true),
				"gitignores": property.New("Go,Node"), "license": property.New("MIT"), "readme": property.New("Default"),
				"issueLabels": property.New("Default"), "trustModel": property.New("committer"), "objectFormatName": property.New("sha256"),
			})
			created, err := prov.Create(context.Background(), p.CreateRequest{Urn: urn, Properties: inputs})
			if err != nil {
				t.Fatalf("create: %v", err)
			}
			refreshed, err := prov.Read(context.Background(), p.ReadRequest{Urn: urn, ID: created.ID, Inputs: inputs, Properties: created.Properties})
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			updated, err := prov.Update(context.Background(), p.UpdateRequest{Urn: urn, ID: created.ID, State: refreshed.Properties, Inputs: inputs.Set("license", property.New("Apache-2.0"))})
			if err != nil {
				t.Fatalf("update: %v", err)
			}
			for _, key := range []string{"initializeWithReadme", "gitignores", "license", "readme", "issueLabels", "trustModel", "objectFormatName"} {
				if !updated.Properties.Get(key).Equals(inputs.Get(key)) {
					t.Errorf("creation-only input %s was lost or changed", key)
				}
			}
		})
	}
}

func TestRepositoryUnmanagedSettingsStayOmitted(t *testing.T) {
	t.Parallel()
	opt := repositorySettingsConfigEditOption(RepositorySettingsConfig{}, nil)
	data, err := json.Marshal(opt)
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, string(data), "{}")
	var repo repositoryAPIResponse
	if err := json.Unmarshal([]byte(`{"allow_rebase_update":true,"default_update_style":"rebase","globally_editable_wiki":true}`), &repo); err != nil {
		t.Fatal(err)
	}
	settings := repositorySettingsConfigFromAPI(RepositorySettingsConfig{}, &repo)
	assertEqual(t, settings, RepositorySettingsConfig{})
}

func TestRepositorySettingsReadFallsBackToManagedState(t *testing.T) {
	t.Parallel()
	settings := RepositorySettingsConfig{
		IgnoreWhitespaceConflicts: boolPtr(false), AllowManualMerge: boolPtr(false), AutodetectManualMerge: boolPtr(true),
		AllowRebaseUpdate: boolPtr(false), DefaultUpdateStyle: stringPtr("rebase"), DefaultAllowMaintainerEdit: boolPtr(false),
		MirrorInterval: stringPtr("1h"), EnablePrune: boolPtr(false),
	}
	args := repositorySettingsReadArgs("alice", "infra", infer.ReadRequest[RepositorySettingsArgs, RepositorySettingsState]{
		State: RepositorySettingsState{RepositorySettingsArgs: RepositorySettingsArgs{RepositorySettingsConfig: settings}},
	})
	assertEqual(t, args.RepositorySettingsConfig, settings)
}

func repositoryAPITestProvider(t *testing.T, url string) p.Provider {
	t.Helper()
	prov := Provider()
	if err := prov.Configure(context.Background(), p.ConfigureRequest{Args: property.NewMap(map[string]property.Value{
		"url": property.New(url), "token": property.New("test-token"),
	})}); err != nil {
		t.Fatal(err)
	}
	return prov
}
