package provider

import (
	"context"
	"fmt"
	"strconv"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v3"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

type PublicKey struct{}

type PublicKeyArgs struct {
	Title    string `pulumi:"title"`
	Key      string `pulumi:"key"`
	ReadOnly bool   `pulumi:"readOnly,optional"`
}

type PublicKeyState struct {
	PublicKeyArgs
	KeyID       int64  `pulumi:"keyId"`
	URL         string `pulumi:"url"`
	Fingerprint string `pulumi:"fingerprint"`
	KeyType     string `pulumi:"keyType"`
	Owner       string `pulumi:"owner"`
}

func (pkey *PublicKey) Annotate(a infer.Annotator) {
	a.Describe(pkey, "A Forgejo SSH public key for the authenticated user.")
	a.SetToken("index", "PublicKey")
}

func (a *PublicKeyArgs) Annotate(ann infer.Annotator) {
	ann.Describe(&a.Title, "Public key title.")
	ann.Describe(&a.Key, "OpenSSH public key material.")
	ann.Describe(&a.ReadOnly, "Whether the key has read-only access.")
}

func (PublicKey) Create(ctx context.Context, req infer.CreateRequest[PublicKeyArgs]) (infer.CreateResponse[PublicKeyState], error) {
	if req.DryRun {
		state := publicKeyStateFromArgs(req.Inputs, 0, "", "", "", "")
		return infer.CreateResponse[PublicKeyState]{ID: req.Inputs.Title, Output: state}, nil
	}

	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.CreateResponse[PublicKeyState]{}, err
	}

	key, _, err := client.CreatePublicKey(forgejo.CreateKeyOption{
		Title:    req.Inputs.Title,
		Key:      req.Inputs.Key,
		ReadOnly: req.Inputs.ReadOnly,
	})
	if err != nil {
		return infer.CreateResponse[PublicKeyState]{}, err
	}

	state := publicKeyStateFromAPI(key)
	return infer.CreateResponse[PublicKeyState]{ID: strconv.FormatInt(key.ID, 10), Output: state}, nil
}

func (PublicKey) Read(ctx context.Context, req infer.ReadRequest[PublicKeyArgs, PublicKeyState]) (infer.ReadResponse[PublicKeyArgs, PublicKeyState], error) {
	keyID, err := publicKeyID(req.ID, req.State)
	if err != nil {
		return infer.ReadResponse[PublicKeyArgs, PublicKeyState]{}, err
	}

	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.ReadResponse[PublicKeyArgs, PublicKeyState]{}, err
	}

	key, resp, err := client.GetPublicKey(keyID)
	if isNotFound(resp) {
		return infer.ReadResponse[PublicKeyArgs, PublicKeyState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[PublicKeyArgs, PublicKeyState]{}, err
	}

	state := publicKeyStateFromAPI(key)
	return infer.ReadResponse[PublicKeyArgs, PublicKeyState]{ID: strconv.FormatInt(key.ID, 10), Inputs: state.PublicKeyArgs, State: state}, nil
}

func (PublicKey) Diff(_ context.Context, req infer.DiffRequest[PublicKeyArgs, PublicKeyState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	addReplaceDiff(diff, "title", req.State.Title != req.Inputs.Title)
	addReplaceDiff(diff, "key", req.State.Key != req.Inputs.Key)
	addReplaceDiff(diff, "readOnly", req.State.ReadOnly != req.Inputs.ReadOnly)
	return p.DiffResponse{HasChanges: len(diff) > 0, DeleteBeforeReplace: true, DetailedDiff: diff}, nil
}

func (PublicKey) Delete(ctx context.Context, req infer.DeleteRequest[PublicKeyState]) (infer.DeleteResponse, error) {
	keyID, err := publicKeyID(req.ID, req.State)
	if err != nil {
		return infer.DeleteResponse{}, err
	}

	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.DeleteResponse{}, err
	}

	resp, err := client.DeletePublicKey(keyID)
	if isNotFound(resp) {
		return infer.DeleteResponse{}, nil
	}
	return infer.DeleteResponse{}, err
}

func publicKeyStateFromAPI(key *forgejo.PublicKey) PublicKeyState {
	owner := ""
	if key.Owner != nil {
		owner = key.Owner.UserName
	}
	return publicKeyStateFromArgs(PublicKeyArgs{Title: key.Title, Key: key.Key, ReadOnly: key.ReadOnly}, key.ID, key.URL, key.Fingerprint, key.KeyType, owner)
}

func publicKeyStateFromArgs(args PublicKeyArgs, keyID int64, url, fingerprint, keyType, owner string) PublicKeyState {
	return PublicKeyState{PublicKeyArgs: args, KeyID: keyID, URL: url, Fingerprint: fingerprint, KeyType: keyType, Owner: owner}
}

func publicKeyID(id string, state PublicKeyState) (int64, error) {
	if state.KeyID != 0 {
		return state.KeyID, nil
	}
	keyID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("public key ID must be numeric, got %q", id)
	}
	return keyID, nil
}
