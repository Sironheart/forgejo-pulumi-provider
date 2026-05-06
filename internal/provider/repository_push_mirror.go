package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v3"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

type RepositoryPushMirror struct{}

type RepositoryPushMirrorArgs struct {
	Owner          string `pulumi:"owner"`
	Repository     string `pulumi:"repository"`
	RemoteAddress  string `pulumi:"remoteAddress"`
	RemoteUsername string `pulumi:"remoteUsername,optional"`
	RemotePassword string `pulumi:"remotePassword,optional" provider:"secret"`
	Interval       string `pulumi:"interval,optional"`
	BranchFilter   string `pulumi:"branchFilter,optional"`
	SyncOnCommit   bool   `pulumi:"syncOnCommit,optional"`
	UseSSH         bool   `pulumi:"useSsh,optional"`
}

type RepositoryPushMirrorState struct {
	RepositoryPushMirrorArgs
	RemoteName string `pulumi:"remoteName"`
	PublicKey  string `pulumi:"publicKey"`
	Created    string `pulumi:"created"`
	LastUpdate string `pulumi:"lastUpdate"`
	LastError  string `pulumi:"lastError"`
}

type repositoryPushMirrorCreateOption struct {
	BranchFilter   string `json:"branch_filter,omitempty"`
	Interval       string `json:"interval,omitempty"`
	RemoteAddress  string `json:"remote_address"`
	RemotePassword string `json:"remote_password,omitempty"`
	RemoteUsername string `json:"remote_username,omitempty"`
	SyncOnCommit   bool   `json:"sync_on_commit,omitempty"`
	UseSSH         bool   `json:"use_ssh,omitempty"`
}

type repositoryPushMirrorAPIResponse struct {
	BranchFilter  string `json:"branch_filter"`
	Created       string `json:"created"`
	Interval      string `json:"interval"`
	LastError     string `json:"last_error"`
	LastUpdate    string `json:"last_update"`
	PublicKey     string `json:"public_key"`
	RemoteAddress string `json:"remote_address"`
	RemoteName    string `json:"remote_name"`
	RepoName      string `json:"repo_name"`
	SyncOnCommit  bool   `json:"sync_on_commit"`
}

type forgejoAPIClient struct {
	baseURL    string
	token      string
	ctx        context.Context
	httpClient *http.Client
}

func (m *RepositoryPushMirror) Annotate(a infer.Annotator) {
	a.Describe(m, "A Forgejo repository push mirror, optionally limited to matching branches.")
	a.SetToken("index", "RepositoryPushMirror")
}

func (a *RepositoryPushMirrorArgs) Annotate(ann infer.Annotator) {
	ann.Describe(&a.Owner, "Repository owner.")
	ann.Describe(&a.Repository, "Repository name.")
	ann.Describe(&a.RemoteAddress, "Target remote URL for the push mirror.")
	ann.Describe(&a.RemoteUsername, "Username for authenticating to the remote.")
	ann.Describe(&a.RemotePassword, "Password or token for authenticating to the remote.")
	ann.Describe(&a.Interval, "Mirror sync interval, for example 8h30m0s. Leave empty to use Forgejo's default.")
	ann.Describe(&a.BranchFilter, "Optional branch filter for the push mirror. Leave empty to mirror all branches.")
	ann.Describe(&a.SyncOnCommit, "Whether pushes to this repository trigger the mirror.")
	ann.Describe(&a.UseSSH, "Whether Forgejo should use an SSH key for the push mirror remote.")
}

func (RepositoryPushMirror) Create(ctx context.Context, req infer.CreateRequest[RepositoryPushMirrorArgs]) (infer.CreateResponse[RepositoryPushMirrorState], error) {
	if req.DryRun {
		return infer.CreateResponse[RepositoryPushMirrorState]{ID: repositoryPushMirrorPreviewID(req.Inputs), Output: repositoryPushMirrorStateFromArgs(req.Inputs, "", "", "", "", "")}, nil
	}

	client, err := forgejoAPIClientFromConfig(ctx)
	if err != nil {
		return infer.CreateResponse[RepositoryPushMirrorState]{}, err
	}

	mirror := new(repositoryPushMirrorAPIResponse)
	_, err = client.do(http.MethodPost, repositoryPushMirrorCollectionPath(req.Inputs.Owner, req.Inputs.Repository), repositoryPushMirrorCreateOption{
		BranchFilter:   req.Inputs.BranchFilter,
		Interval:       req.Inputs.Interval,
		RemoteAddress:  req.Inputs.RemoteAddress,
		RemotePassword: req.Inputs.RemotePassword,
		RemoteUsername: req.Inputs.RemoteUsername,
		SyncOnCommit:   req.Inputs.SyncOnCommit,
		UseSSH:         req.Inputs.UseSSH,
	}, mirror)
	if err != nil {
		return infer.CreateResponse[RepositoryPushMirrorState]{}, err
	}
	if mirror.RemoteName == "" {
		return infer.CreateResponse[RepositoryPushMirrorState]{}, fmt.Errorf("forgejo push mirror response did not include remote name")
	}

	state := repositoryPushMirrorStateFromAPI(req.Inputs, mirror)
	return infer.CreateResponse[RepositoryPushMirrorState]{ID: repositoryPushMirrorID(state.Owner, state.Repository, state.RemoteName), Output: state}, nil
}

func (RepositoryPushMirror) Read(ctx context.Context, req infer.ReadRequest[RepositoryPushMirrorArgs, RepositoryPushMirrorState]) (infer.ReadResponse[RepositoryPushMirrorArgs, RepositoryPushMirrorState], error) {
	owner, repo, remoteName, err := repositoryPushMirrorParts(req.ID, req.State)
	if err != nil {
		return infer.ReadResponse[RepositoryPushMirrorArgs, RepositoryPushMirrorState]{}, err
	}

	client, err := forgejoAPIClientFromConfig(ctx)
	if err != nil {
		return infer.ReadResponse[RepositoryPushMirrorArgs, RepositoryPushMirrorState]{}, err
	}

	mirror := new(repositoryPushMirrorAPIResponse)
	resp, err := client.do(http.MethodGet, repositoryPushMirrorRemotePath(owner, repo, remoteName), nil, mirror)
	if isNotFound(resp) {
		return infer.ReadResponse[RepositoryPushMirrorArgs, RepositoryPushMirrorState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[RepositoryPushMirrorArgs, RepositoryPushMirrorState]{}, err
	}
	if mirror.RemoteName == "" {
		mirror.RemoteName = remoteName
	}

	state := repositoryPushMirrorStateFromAPI(repositoryPushMirrorReadArgs(owner, repo, req), mirror)
	return infer.ReadResponse[RepositoryPushMirrorArgs, RepositoryPushMirrorState]{ID: repositoryPushMirrorID(state.Owner, state.Repository, state.RemoteName), Inputs: state.RepositoryPushMirrorArgs, State: state}, nil
}

func (RepositoryPushMirror) Diff(_ context.Context, req infer.DiffRequest[RepositoryPushMirrorArgs, RepositoryPushMirrorState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	addReplaceDiff(diff, "owner", req.State.Owner != req.Inputs.Owner)
	addReplaceDiff(diff, "repository", req.State.Repository != req.Inputs.Repository)
	addReplaceDiff(diff, "remoteAddress", req.State.RemoteAddress != req.Inputs.RemoteAddress)
	addReplaceDiff(diff, "remoteUsername", req.State.RemoteUsername != req.Inputs.RemoteUsername)
	addReplaceDiff(diff, "remotePassword", req.State.RemotePassword != req.Inputs.RemotePassword)
	addReplaceDiff(diff, "interval", req.State.Interval != req.Inputs.Interval)
	addReplaceDiff(diff, "branchFilter", req.State.BranchFilter != req.Inputs.BranchFilter)
	addReplaceDiff(diff, "syncOnCommit", req.State.SyncOnCommit != req.Inputs.SyncOnCommit)
	addReplaceDiff(diff, "useSsh", req.State.UseSSH != req.Inputs.UseSSH)
	return p.DiffResponse{HasChanges: len(diff) > 0, DeleteBeforeReplace: true, DetailedDiff: diff}, nil
}

func (RepositoryPushMirror) Delete(ctx context.Context, req infer.DeleteRequest[RepositoryPushMirrorState]) (infer.DeleteResponse, error) {
	owner, repo, remoteName, err := repositoryPushMirrorParts(req.ID, req.State)
	if err != nil {
		return infer.DeleteResponse{}, err
	}

	client, err := forgejoAPIClientFromConfig(ctx)
	if err != nil {
		return infer.DeleteResponse{}, err
	}

	resp, err := client.do(http.MethodDelete, repositoryPushMirrorRemotePath(owner, repo, remoteName), nil, nil)
	if isNotFound(resp) {
		return infer.DeleteResponse{}, nil
	}
	return infer.DeleteResponse{}, err
}

func forgejoAPIClientFromConfig(ctx context.Context) (*forgejoAPIClient, error) {
	cfg := infer.GetConfig[Config](ctx)
	if strings.TrimSpace(cfg.Token) == "" {
		return nil, fmt.Errorf("forgejo token is required")
	}
	baseURL, err := normalizeForgejoURL(cfg.URL)
	if err != nil {
		return nil, err
	}
	return &forgejoAPIClient{baseURL: baseURL, token: cfg.Token, ctx: ctx, httpClient: &http.Client{Timeout: defaultTimeout}}, nil
}

func (c *forgejoAPIClient) do(method, path string, input, output any) (*forgejo.Response, error) {
	var body io.Reader
	if input != nil {
		data, err := json.Marshal(input)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(c.ctx, method, strings.TrimRight(c.baseURL, "/")+"/api/v1"+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "token "+c.token)
	if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	forgejoResp := &forgejo.Response{Response: resp}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return forgejoResp, err
	}
	if resp.StatusCode/100 != 2 {
		return forgejoResp, forgejoAPIError(resp, data)
	}
	if output != nil && len(data) > 0 {
		if err := json.Unmarshal(data, output); err != nil {
			return forgejoResp, err
		}
	}
	return forgejoResp, nil
}

func forgejoAPIError(resp *http.Response, data []byte) error {
	payload := map[string]any{}
	if err := json.Unmarshal(data, &payload); err == nil {
		if message, ok := payload["message"]; ok {
			return fmt.Errorf("%v", message)
		}
	}
	if len(data) == 0 {
		return fmt.Errorf("%s", resp.Status)
	}
	return fmt.Errorf("%s: %s", resp.Status, string(data))
}

func repositoryPushMirrorStateFromAPI(args RepositoryPushMirrorArgs, mirror *repositoryPushMirrorAPIResponse) RepositoryPushMirrorState {
	if mirror == nil {
		return repositoryPushMirrorStateFromArgs(args, "", "", "", "", "")
	}

	stateArgs := args
	stateArgs.RemoteAddress = mirror.RemoteAddress
	stateArgs.Interval = mirror.Interval
	stateArgs.BranchFilter = mirror.BranchFilter
	stateArgs.SyncOnCommit = mirror.SyncOnCommit
	stateArgs.UseSSH = args.UseSSH || mirror.PublicKey != ""

	return repositoryPushMirrorStateFromArgs(stateArgs, mirror.RemoteName, mirror.PublicKey, mirror.Created, mirror.LastUpdate, mirror.LastError)
}

func repositoryPushMirrorStateFromArgs(args RepositoryPushMirrorArgs, remoteName, publicKey, created, lastUpdate, lastError string) RepositoryPushMirrorState {
	return RepositoryPushMirrorState{RepositoryPushMirrorArgs: args, RemoteName: remoteName, PublicKey: publicKey, Created: created, LastUpdate: lastUpdate, LastError: lastError}
}

func repositoryPushMirrorReadArgs(owner, repo string, req infer.ReadRequest[RepositoryPushMirrorArgs, RepositoryPushMirrorState]) RepositoryPushMirrorArgs {
	args := req.Inputs
	args.Owner = owner
	args.Repository = repo
	if args.RemoteUsername == "" {
		args.RemoteUsername = req.State.RemoteUsername
	}
	if args.RemotePassword == "" {
		args.RemotePassword = req.State.RemotePassword
	}
	args.UseSSH = args.UseSSH || req.State.UseSSH
	return args
}

func repositoryPushMirrorCollectionPath(owner, repo string) string {
	return fmt.Sprintf("/repos/%s/%s/push_mirrors", url.PathEscape(owner), url.PathEscape(repo))
}

func repositoryPushMirrorRemotePath(owner, repo, remoteName string) string {
	return fmt.Sprintf("%s/%s", repositoryPushMirrorCollectionPath(owner, repo), url.PathEscape(remoteName))
}

func repositoryPushMirrorID(owner, repo, remoteName string) string {
	return fmt.Sprintf("%s/%s/%s", owner, repo, remoteName)
}

func repositoryPushMirrorPreviewID(args RepositoryPushMirrorArgs) string {
	return repositoryPushMirrorID(args.Owner, args.Repository, args.RemoteAddress)
}

func repositoryPushMirrorParts(id string, state RepositoryPushMirrorState) (string, string, string, error) {
	owner, repo, remoteName := state.Owner, state.Repository, state.RemoteName
	if owner == "" || repo == "" || remoteName == "" {
		parts := strings.SplitN(id, "/", 3)
		if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
			return "", "", "", fmt.Errorf("repository push mirror ID must have the form owner/repository/remoteName, got %q", id)
		}
		if owner == "" {
			owner = parts[0]
		}
		if repo == "" {
			repo = parts[1]
		}
		if remoteName == "" {
			remoteName = parts[2]
		}
	}
	return owner, repo, remoteName, nil
}
