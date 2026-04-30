package provider

import (
	"context"
	"fmt"
	"strconv"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v3"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

type DeployKey struct{}

type DeployKeyArgs struct {
	Owner      string `pulumi:"owner"`
	Repository string `pulumi:"repository"`
	Title      string `pulumi:"title"`
	Key        string `pulumi:"key"`
	ReadOnly   bool   `pulumi:"readOnly,optional"`
}

type DeployKeyState struct {
	DeployKeyArgs
	KeyID       int64  `pulumi:"keyId"`
	URL         string `pulumi:"url"`
	Fingerprint string `pulumi:"fingerprint"`
}

func (d *DeployKey) Annotate(a infer.Annotator) {
	a.Describe(d, "A Forgejo deploy key for a repository.")
	a.SetToken("index", "DeployKey")
}

func (a *DeployKeyArgs) Annotate(ann infer.Annotator) {
	ann.Describe(&a.Owner, "Repository owner.")
	ann.Describe(&a.Repository, "Repository name.")
	ann.Describe(&a.Title, "Deploy key title.")
	ann.Describe(&a.Key, "OpenSSH public key material.")
	ann.Describe(&a.ReadOnly, "Whether the deploy key is read-only.")
	ann.SetDefault(&a.ReadOnly, true)
}

func (DeployKey) Create(ctx context.Context, req infer.CreateRequest[DeployKeyArgs]) (infer.CreateResponse[DeployKeyState], error) {
	if req.DryRun {
		state := deployKeyStateFromArgs(req.Inputs, 0, "", "")
		return infer.CreateResponse[DeployKeyState]{ID: deployKeyPreviewID(req.Inputs), Output: state}, nil
	}

	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.CreateResponse[DeployKeyState]{}, err
	}

	key, _, err := client.CreateDeployKey(req.Inputs.Owner, req.Inputs.Repository, forgejo.CreateKeyOption{
		Title:    req.Inputs.Title,
		Key:      req.Inputs.Key,
		ReadOnly: req.Inputs.ReadOnly,
	})
	if err != nil {
		return infer.CreateResponse[DeployKeyState]{}, err
	}

	state := deployKeyStateFromAPI(req.Inputs.Owner, req.Inputs.Repository, key)
	return infer.CreateResponse[DeployKeyState]{ID: strconv.FormatInt(key.ID, 10), Output: state}, nil
}

func (DeployKey) Read(ctx context.Context, req infer.ReadRequest[DeployKeyArgs, DeployKeyState]) (infer.ReadResponse[DeployKeyArgs, DeployKeyState], error) {
	keyID, err := deployKeyID(req.ID, req.State)
	if err != nil {
		return infer.ReadResponse[DeployKeyArgs, DeployKeyState]{}, err
	}

	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.ReadResponse[DeployKeyArgs, DeployKeyState]{}, err
	}

	key, resp, err := client.GetDeployKey(req.State.Owner, req.State.Repository, keyID)
	if isNotFound(resp) {
		return infer.ReadResponse[DeployKeyArgs, DeployKeyState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[DeployKeyArgs, DeployKeyState]{}, err
	}

	state := deployKeyStateFromAPI(req.State.Owner, req.State.Repository, key)
	return infer.ReadResponse[DeployKeyArgs, DeployKeyState]{ID: strconv.FormatInt(key.ID, 10), Inputs: state.DeployKeyArgs, State: state}, nil
}

func (DeployKey) Diff(_ context.Context, req infer.DiffRequest[DeployKeyArgs, DeployKeyState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	addReplaceDiff(diff, "owner", req.State.Owner != req.Inputs.Owner)
	addReplaceDiff(diff, "repository", req.State.Repository != req.Inputs.Repository)
	addReplaceDiff(diff, "title", req.State.Title != req.Inputs.Title)
	addReplaceDiff(diff, "key", req.State.Key != req.Inputs.Key)
	addReplaceDiff(diff, "readOnly", req.State.ReadOnly != req.Inputs.ReadOnly)
	return p.DiffResponse{HasChanges: len(diff) > 0, DeleteBeforeReplace: true, DetailedDiff: diff}, nil
}

func (DeployKey) Delete(ctx context.Context, req infer.DeleteRequest[DeployKeyState]) (infer.DeleteResponse, error) {
	keyID, err := deployKeyID(req.ID, req.State)
	if err != nil {
		return infer.DeleteResponse{}, err
	}

	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.DeleteResponse{}, err
	}

	resp, err := client.DeleteDeployKey(req.State.Owner, req.State.Repository, keyID)
	if isNotFound(resp) {
		return infer.DeleteResponse{}, nil
	}
	return infer.DeleteResponse{}, err
}

func deployKeyStateFromAPI(owner, repo string, key *forgejo.DeployKey) DeployKeyState {
	return DeployKeyState{
		DeployKeyArgs: DeployKeyArgs{
			Owner:      owner,
			Repository: repo,
			Title:      key.Title,
			Key:        key.Key,
			ReadOnly:   key.ReadOnly,
		},
		KeyID:       key.ID,
		URL:         key.URL,
		Fingerprint: key.Fingerprint,
	}
}

func deployKeyStateFromArgs(args DeployKeyArgs, keyID int64, url, fingerprint string) DeployKeyState {
	return DeployKeyState{DeployKeyArgs: args, KeyID: keyID, URL: url, Fingerprint: fingerprint}
}

func deployKeyID(id string, state DeployKeyState) (int64, error) {
	if state.KeyID != 0 {
		return state.KeyID, nil
	}
	keyID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("deploy key ID must be numeric, got %q", id)
	}
	return keyID, nil
}

func deployKeyPreviewID(args DeployKeyArgs) string {
	return fmt.Sprintf("%s/%s/%s", args.Owner, args.Repository, args.Title)
}

func addReplaceDiff(diff map[string]p.PropertyDiff, field string, changed bool) {
	if changed {
		diff[field] = p.PropertyDiff{Kind: p.UpdateReplace, InputDiff: true}
	}
}
