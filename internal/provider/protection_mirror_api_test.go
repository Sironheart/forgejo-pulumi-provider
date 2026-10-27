package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
)

func TestRepositoryBranchProtectionAPILifecycle(t *testing.T) {
	t.Parallel()

	listFields := map[string]string{
		"pushWhitelistUsernames":      "push_whitelist_usernames",
		"pushWhitelistTeams":          "push_whitelist_teams",
		"mergeWhitelistUsernames":     "merge_whitelist_usernames",
		"mergeWhitelistTeams":         "merge_whitelist_teams",
		"statusCheckContexts":         "status_check_contexts",
		"approvalsWhitelistUsernames": "approvals_whitelist_username",
		"approvalsWhitelistTeams":     "approvals_whitelist_teams",
	}
	requests := make(chan map[string]json.RawMessage, 2)
	prov := protectionMirrorTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		path := "/api/v1/repos/platform/infra/branch_protections"
		if r.Method != http.MethodPost {
			path += "/release%2F%2A"
		}
		if r.URL.EscapedPath() != path {
			t.Errorf("path = %q, want %q", r.URL.EscapedPath(), path)
		}
		switch r.Method {
		case http.MethodPost, http.MethodPatch:
			body := protectionMirrorRequestJSON(t, r)
			requests <- body
			protectionMirrorWriteJSON(t, w, body)
		case http.MethodGet:
			protectionMirrorWriteJSON(t, w, map[string]any{
				"branch_name": "release/*", "apply_to_admins": true, "ignore_stale_approvals": true,
				"created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-10-27T00:00:00Z",
			})
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected method: %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	urn := protectionMirrorURN("RepositoryBranchProtection")
	inputs := property.NewMap(map[string]property.Value{
		"owner": property.New("platform"), "repository": property.New("infra"), "name": property.New("release/*"),
		"applyToAdmins": property.New(true), "ignoreStaleApprovals": property.New(true),
	})
	for field := range listFields {
		inputs = inputs.Set(field, property.New([]property.Value{property.New("allowed")}))
	}
	created, err := prov.Create(t.Context(), p.CreateRequest{Urn: urn, Properties: inputs})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	assertEqual(t, created.ID, "platform/infra/release/*")
	assertEqual(t, created.Properties.Get("applyToAdmins"), property.New(true))
	assertEqual(t, created.Properties.Get("ignoreStaleApprovals"), property.New(true))
	body := <-requests
	assertEqual(t, string(body["rule_name"]), `"release/*"`)
	assertEqual(t, string(body["branch_name"]), "")
	assertEqual(t, string(body["apply_to_admins"]), "true")
	assertEqual(t, string(body["ignore_stale_approvals"]), "true")
	assertEqual(t, string(body["approvals_whitelist_username"]), `["allowed"]`)
	assertEqual(t, string(body["approvals_whitelist_usernames"]), "")

	updatedInputs := inputs.Set("applyToAdmins", property.New(false)).Set("ignoreStaleApprovals", property.New(false))
	for field := range listFields {
		updatedInputs = updatedInputs.Delete(field)
	}
	updatedInputs = updatedInputs.Set("pushWhitelistTeams", property.New([]property.Value{}))
	diff, err := prov.Diff(t.Context(), p.DiffRequest{Urn: urn, ID: created.ID, State: created.Properties, Inputs: updatedInputs})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	for field := range listFields {
		assertDiffKind(t, diff, field, p.Update)
	}
	assertDiffKind(t, diff, "applyToAdmins", p.Update)
	assertDiffKind(t, diff, "ignoreStaleApprovals", p.Update)
	updated, err := prov.Update(t.Context(), p.UpdateRequest{Urn: urn, ID: created.ID, State: created.Properties, Inputs: updatedInputs})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	body = <-requests
	for field, wire := range listFields {
		assertEqual(t, string(body[wire]), "[]")
		value := updated.Properties.Get(field)
		if !value.IsNull() && (!value.IsArray() || value.AsArray().Len() != 0) {
			t.Errorf("%s was not cleared: %v", field, value)
		}
	}
	assertEqual(t, string(body["apply_to_admins"]), "false")
	assertEqual(t, string(body["ignore_stale_approvals"]), "false")
	assertEqual(t, updated.Properties.Get("applyToAdmins"), property.New(false))
	assertEqual(t, updated.Properties.Get("ignoreStaleApprovals"), property.New(false))
	diff, err = prov.Diff(t.Context(), p.DiffRequest{Urn: urn, ID: created.ID, State: updated.Properties, Inputs: updatedInputs})
	if err != nil || diff.HasChanges {
		t.Fatalf("diff after clearing lists: %#v, %v", diff, err)
	}
	read, err := prov.Read(t.Context(), p.ReadRequest{Urn: urn, ID: created.ID, Properties: updated.Properties, Inputs: updatedInputs})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	assertEqual(t, read.ID, created.ID)
	assertEqual(t, read.Properties.Get("applyToAdmins"), property.New(true))
	assertEqual(t, read.Inputs.Get("ignoreStaleApprovals"), property.New(true))
	assertEqual(t, read.Properties.Get("created"), property.New("2026-01-01T00:00:00Z"))
	assertEqual(t, read.Properties.Get("updated"), property.New("2026-10-27T00:00:00Z"))
	if err := prov.Delete(t.Context(), p.DeleteRequest{Urn: urn, ID: read.ID, Properties: read.Properties}); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestRepositoryBranchProtectionAPIUnmanagedFields(t *testing.T) {
	t.Parallel()

	prov := protectionMirrorTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPatch {
			body := protectionMirrorRequestJSON(t, r)
			for _, field := range []string{"apply_to_admins", "ignore_stale_approvals", "branch_name"} {
				if _, ok := body[field]; ok {
					t.Errorf("unmanaged/deprecated field %s sent: %s", field, body[field])
				}
			}
		}
		protectionMirrorWriteJSON(t, w, map[string]any{"rule_name": "main", "apply_to_admins": true, "ignore_stale_approvals": true})
	})
	urn := protectionMirrorURN("RepositoryBranchProtection")
	inputs := property.NewMap(map[string]property.Value{
		"owner": property.New("platform"), "repository": property.New("infra"), "name": property.New("main"),
	})
	created, err := prov.Create(t.Context(), p.CreateRequest{Urn: urn, Properties: inputs})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	read, err := prov.Read(t.Context(), p.ReadRequest{Urn: urn, ID: created.ID, Properties: created.Properties, Inputs: inputs})
	if err != nil {
		t.Fatalf("read old state: %v", err)
	}
	updated, err := prov.Update(t.Context(), p.UpdateRequest{Urn: urn, ID: created.ID, State: read.Properties, Inputs: inputs})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	for _, state := range []property.Map{created.Properties, read.Properties, read.Inputs, updated.Properties} {
		for _, field := range []string{"applyToAdmins", "ignoreStaleApprovals"} {
			if !state.Get(field).IsNull() {
				t.Errorf("unmanaged %s became managed: %v", field, state.Get(field))
			}
		}
	}
	for _, state := range []property.Map{created.Properties, created.Properties.Set("applyToAdmins", property.New(true)).Set("ignoreStaleApprovals", property.New(true))} {
		diff, err := prov.Diff(t.Context(), p.DiffRequest{Urn: urn, ID: created.ID, State: state, Inputs: inputs})
		if err != nil || diff.HasChanges {
			t.Fatalf("unmanaged fields diff: %#v, %v", diff, err)
		}
		diff, err = prov.Diff(t.Context(), p.DiffRequest{Urn: urn, ID: created.ID, State: state, Inputs: inputs.Set("applyToAdmins", property.New(false)).Set("ignoreStaleApprovals", property.New(false))})
		if err != nil {
			t.Fatalf("explicit false diff: %v", err)
		}
		assertDiffKind(t, diff, "applyToAdmins", p.Update)
		assertDiffKind(t, diff, "ignoreStaleApprovals", p.Update)
	}
}

func TestRepositoryTagProtectionAPIClearsLists(t *testing.T) {
	t.Parallel()

	for _, explicitEmpty := range []bool{false, true} {
		t.Run(fmt.Sprintf("explicit-empty-%t", explicitEmpty), func(t *testing.T) {
			t.Parallel()
			requests := make(chan map[string]json.RawMessage, 2)
			prov := protectionMirrorTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodPost, http.MethodPatch:
					body := protectionMirrorRequestJSON(t, r)
					body["id"] = json.RawMessage("42")
					requests <- body
					protectionMirrorWriteJSON(t, w, body)
				case http.MethodGet:
					protectionMirrorWriteJSON(t, w, map[string]any{"id": 42, "name_pattern": "v*", "whitelist_usernames": []string{}, "whitelist_teams": []string{}})
				default:
					w.WriteHeader(http.StatusNoContent)
				}
			})
			urn := protectionMirrorURN("RepositoryTagProtection")
			inputs := property.NewMap(map[string]property.Value{
				"owner": property.New("platform"), "repository": property.New("infra"), "namePattern": property.New("v*"),
				"whitelistUsernames": property.New([]property.Value{property.New("alice")}),
				"whitelistTeams":     property.New([]property.Value{property.New("release")}),
			})
			created, err := prov.Create(t.Context(), p.CreateRequest{Urn: urn, Properties: inputs})
			if err != nil {
				t.Fatalf("create: %v", err)
			}
			assertEqual(t, created.ID, "42")
			body := <-requests
			assertEqual(t, string(body["whitelist_usernames"]), `["alice"]`)
			assertEqual(t, string(body["whitelist_teams"]), `["release"]`)
			for _, field := range []string{"whitelistUsernames", "whitelistTeams"} {
				if explicitEmpty {
					inputs = inputs.Set(field, property.New([]property.Value{}))
				} else {
					inputs = inputs.Delete(field)
				}
			}
			diff, err := prov.Diff(t.Context(), p.DiffRequest{Urn: urn, ID: created.ID, State: created.Properties, Inputs: inputs})
			if err != nil {
				t.Fatalf("diff: %v", err)
			}
			assertDiffKind(t, diff, "whitelistUsernames", p.Update)
			assertDiffKind(t, diff, "whitelistTeams", p.Update)
			updated, err := prov.Update(t.Context(), p.UpdateRequest{Urn: urn, ID: created.ID, State: created.Properties, Inputs: inputs})
			if err != nil {
				t.Fatalf("update: %v", err)
			}
			body = <-requests
			assertEqual(t, string(body["whitelist_usernames"]), "[]")
			assertEqual(t, string(body["whitelist_teams"]), "[]")
			read, err := prov.Read(t.Context(), p.ReadRequest{Urn: urn, ID: created.ID, Properties: updated.Properties, Inputs: inputs})
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			diff, err = prov.Diff(t.Context(), p.DiffRequest{Urn: urn, ID: read.ID, State: read.Properties, Inputs: inputs})
			if err != nil || diff.HasChanges {
				t.Fatalf("diff after clearing lists: %#v, %v", diff, err)
			}
			if err := prov.Delete(t.Context(), p.DeleteRequest{Urn: urn, ID: read.ID, Properties: read.Properties}); err != nil {
				t.Fatalf("delete: %v", err)
			}
		})
	}
}

func TestRepositoryPushMirrorAPILifecycle(t *testing.T) {
	t.Parallel()

	requests := make(chan map[string]json.RawMessage, 1)
	prov := protectionMirrorTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method == http.MethodPost {
			requests <- protectionMirrorRequestJSON(t, r)
		}
		protectionMirrorWriteJSON(t, w, map[string]any{
			"remote_name": "remote_mirror_one", "remote_address": "git@example.test:platform/infra.git",
			"interval": "1h0m0s", "branch_filter": "main,release/*", "public_key": "ssh-ed25519 test-key",
			"sync_on_commit": true, "created": "2026-01-01T00:00:00Z", "last_update": nil, "last_error": "",
		})
	})
	urn := protectionMirrorURN("RepositoryPushMirror")
	inputs := property.NewMap(map[string]property.Value{
		"owner": property.New("platform"), "repository": property.New("infra"),
		"remoteAddress": property.New("git@example.test:platform/infra.git"), "interval": property.New("1h"),
		"branchFilter": property.New("main,release/*"), "useSsh": property.New(true), "syncOnCommit": property.New(true),
	})
	created, err := prov.Create(t.Context(), p.CreateRequest{Urn: urn, Properties: inputs})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	body := <-requests
	assertEqual(t, string(body["interval"]), `"1h"`)
	assertEqual(t, string(body["branch_filter"]), `"main,release/*"`)
	assertEqual(t, string(body["use_ssh"]), "true")
	assertEqual(t, string(body["remote_password"]), "")
	assertEqual(t, created.ID, "platform/infra/remote_mirror_one")
	assertEqual(t, created.Properties.Get("publicKey"), property.New("ssh-ed25519 test-key"))
	read, err := prov.Read(t.Context(), p.ReadRequest{Urn: urn, ID: created.ID, Properties: created.Properties, Inputs: inputs})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	assertEqual(t, read.Properties.Get("interval"), property.New("1h0m0s"))
	assertEqual(t, read.Properties.Get("useSsh"), property.New(true))
	diff, err := prov.Diff(t.Context(), p.DiffRequest{Urn: urn, ID: read.ID, State: read.Properties, Inputs: inputs})
	if err != nil || diff.HasChanges {
		t.Fatalf("normalized interval would replace SSH mirror: %#v, %v", diff, err)
	}
	diff, err = prov.Diff(t.Context(), p.DiffRequest{Urn: urn, ID: read.ID, State: read.Properties, Inputs: inputs.Set("interval", property.New("2h"))})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	assertDiffKind(t, diff, "interval", p.UpdateReplace)
	assertEqual(t, diff.DeleteBeforeReplace, true)
	if err := prov.Delete(t.Context(), p.DeleteRequest{Urn: urn, ID: read.ID, Properties: read.Properties}); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestRepositoryPushMirrorIntervalCheck(t *testing.T) {
	t.Parallel()
	prov := protectionMirrorTestProvider(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Error("invalid input made an HTTP request")
		w.WriteHeader(http.StatusInternalServerError)
	})
	inputs := property.NewMap(map[string]property.Value{
		"owner": property.New("platform"), "repository": property.New("infra"), "remoteAddress": property.New("https://example.test/infra.git"),
	})
	for _, tc := range []struct {
		name     string
		value    property.Value
		valid    bool
		property string
	}{
		{"missing", property.Value{}, false, "interval"},
		{"empty", property.New(""), false, "interval"},
		{"malformed", property.New("daily"), false, "interval"},
		{"negative", property.New("-1h"), false, "interval"},
		{"overflow", property.New("999999999999999999999h"), false, "interval"},
		{"wrong-type", property.New(true), false, "interval"},
		{"zero", property.New("0s"), true, ""},
		{"bare-zero", property.New("0"), true, ""},
		{"duration", property.New("8h30m0s"), true, ""},
		{"unknown", property.New(property.Computed), true, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			news := inputs
			if !tc.value.IsNull() {
				news = news.Set("interval", tc.value)
			}
			checked, err := prov.Check(t.Context(), p.CheckRequest{Urn: protectionMirrorURN("RepositoryPushMirror"), Inputs: news})
			if err != nil {
				t.Fatalf("check: %v", err)
			}
			if tc.valid {
				assertEqual(t, len(checked.Failures), 0)
				assertEqual(t, checked.Inputs.Get("interval"), tc.value)
			} else {
				if len(checked.Failures) == 0 {
					t.Fatal("expected a check failure")
				}
				assertEqual(t, checked.Failures[0].Property, tc.property)
				if tc.value.IsString() {
					_, err := prov.Create(t.Context(), p.CreateRequest{Urn: protectionMirrorURN("RepositoryPushMirror"), Properties: news})
					if err == nil || !strings.Contains(err.Error(), "interval") {
						t.Fatalf("create error = %v, want interval validation", err)
					}
				}
			}
		})
	}
}

func TestRepositoryPushMirrorIntervalDiff(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		old, next string
		changed   bool
	}{
		{"1h0m0s", "60m", false}, {"0s", "0", false}, {"1h0m0s", "2h", true},
		{"1h0m0s", "", true}, {"broken", "also-broken", true}, {"-1h", "-60m", true},
	} {
		diff, err := (RepositoryPushMirror{}).Diff(context.Background(), infer.DiffRequest[RepositoryPushMirrorArgs, RepositoryPushMirrorState]{
			State:  RepositoryPushMirrorState{RepositoryPushMirrorArgs: RepositoryPushMirrorArgs{Interval: tc.old}},
			Inputs: RepositoryPushMirrorArgs{Interval: tc.next},
		})
		if err != nil {
			t.Fatalf("diff: %v", err)
		}
		if diff.HasChanges != tc.changed {
			t.Errorf("interval %q -> %q: HasChanges = %t, want %t", tc.old, tc.next, diff.HasChanges, tc.changed)
		}
	}
}

func TestProtectionMirrorAPIErrors(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"RepositoryBranchProtection", "RepositoryTagProtection", "RepositoryPushMirror"} {
		for _, method := range []string{http.MethodPost, http.MethodGet, http.MethodPatch, http.MethodDelete} {
			if kind == "RepositoryPushMirror" && method == http.MethodPatch {
				continue
			}
			for _, status := range []int{http.StatusForbidden, http.StatusNotFound, http.StatusInternalServerError} {
				t.Run(fmt.Sprintf("%s/%s/%d", kind, method, status), func(t *testing.T) {
					t.Parallel()
					prov := protectionMirrorTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
						if r.Method != method {
							t.Errorf("method = %s, want %s", r.Method, method)
						}
						w.WriteHeader(status)
						protectionMirrorWriteJSON(t, w, map[string]string{"message": "upstream failure"})
					})
					state := property.NewMap(map[string]property.Value{
						"owner": property.New("platform"), "repository": property.New("infra"),
					})
					id := "platform/infra/main"
					switch kind {
					case "RepositoryBranchProtection":
						state = state.Set("name", property.New("main")).Set("created", property.New("")).Set("updated", property.New(""))
					case "RepositoryTagProtection":
						id = "42"
						state = state.Set("namePattern", property.New("v*")).Set("protectionId", property.New(42.0))
					case "RepositoryPushMirror":
						id = "platform/infra/remote_mirror_one"
						state = state.Set("interval", property.New("1h")).Set("remoteAddress", property.New("https://example.test/infra.git")).
							Set("remoteName", property.New("remote_mirror_one")).Set("publicKey", property.New("")).
							Set("lastUpdate", property.New("")).Set("lastError", property.New("")).Set("created", property.New(""))
					}
					inputs := state.Delete("created").Delete("updated").Delete("protectionId").Delete("remoteName").Delete("publicKey").Delete("lastUpdate").Delete("lastError")
					urn := protectionMirrorURN(kind)
					var err error
					switch method {
					case http.MethodPost:
						_, err = prov.Create(t.Context(), p.CreateRequest{Urn: urn, Properties: inputs})
					case http.MethodGet:
						var read p.ReadResponse
						read, err = prov.Read(t.Context(), p.ReadRequest{Urn: urn, ID: id, Properties: state, Inputs: inputs})
						if status == http.StatusNotFound {
							assertEqual(t, read.ID, "")
						}
					case http.MethodPatch:
						_, err = prov.Update(t.Context(), p.UpdateRequest{Urn: urn, ID: id, State: state, Inputs: inputs})
					case http.MethodDelete:
						err = prov.Delete(t.Context(), p.DeleteRequest{Urn: urn, ID: id, Properties: state})
					}
					if status == http.StatusNotFound && (method == http.MethodGet || method == http.MethodDelete) {
						if err != nil {
							t.Fatalf("not-found should be ignored: %v", err)
						}
					} else if err == nil || !strings.Contains(err.Error(), "upstream failure") {
						t.Fatalf("error = %v, want upstream failure", err)
					}
				})
			}
		}
	}
}

func protectionMirrorTestProvider(t *testing.T, handler http.HandlerFunc) p.Provider {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "token test-token" {
			t.Errorf("unexpected authorization header: %q", r.Header.Get("Authorization"))
		}
		handler(w, r)
	}))
	t.Cleanup(server.Close)
	prov := infer.Provider(infer.Options{
		Resources: []infer.InferredResource{infer.Resource(RepositoryBranchProtection{}), infer.Resource(RepositoryTagProtection{}), infer.Resource(RepositoryPushMirror{})},
		Config:    infer.Config(&Config{}),
	})
	if err := prov.Configure(t.Context(), p.ConfigureRequest{Args: property.NewMap(map[string]property.Value{
		"url": property.New(server.URL), "token": property.New("test-token"),
	})}); err != nil {
		t.Fatalf("configure: %v", err)
	}
	return prov
}

func protectionMirrorURN(kind string) resource.URN {
	return resource.URN("urn:pulumi:stack::project::forgejo:index:" + kind + "::test")
}

func protectionMirrorRequestJSON(t *testing.T, r *http.Request) map[string]json.RawMessage {
	t.Helper()
	if r.Header.Get("Content-Type") != "application/json" {
		t.Errorf("unexpected content type: %q", r.Header.Get("Content-Type"))
	}
	body := map[string]json.RawMessage{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		t.Errorf("request JSON: %v", err)
	}
	return body
}

func protectionMirrorWriteJSON(t *testing.T, w http.ResponseWriter, body any) {
	t.Helper()
	if err := json.NewEncoder(w).Encode(body); err != nil {
		t.Errorf("response JSON: %v", err)
	}
}
