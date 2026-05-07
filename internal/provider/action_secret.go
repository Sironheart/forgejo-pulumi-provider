package provider

import (
	"context"
	"fmt"
	"time"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v3"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

type RepositoryActionSecret struct{}

type ActionSecretArgs struct {
	Name  string `pulumi:"name"`
	Value string `pulumi:"value" provider:"secret"`
}

type RepositoryActionSecretArgs struct {
	ActionSecretArgs

	Owner      string `pulumi:"owner"`
	Repository string `pulumi:"repository"`
}

type RepositoryActionSecretState struct {
	RepositoryActionSecretArgs
	Created string `pulumi:"created"`
}

type OrganizationActionSecret struct{}

type OrganizationActionSecretArgs struct {
	ActionSecretArgs

	Organization string `pulumi:"organization"`
}

type OrganizationActionSecretState struct {
	OrganizationActionSecretArgs
	Created string `pulumi:"created"`
}

func (s *RepositoryActionSecret) Annotate(a infer.Annotator) {
	a.Describe(s, "A Forgejo Actions secret for a repository.")
	a.SetToken("index", "RepositoryActionSecret")
}

func (a *RepositoryActionSecretArgs) Annotate(ann infer.Annotator) {
	annotateActionSecretArgs(&a.ActionSecretArgs, ann)
	ann.Describe(&a.Owner, "Repository owner.")
	ann.Describe(&a.Repository, "Repository name.")
}

func (s *OrganizationActionSecret) Annotate(a infer.Annotator) {
	a.Describe(s, "A Forgejo Actions secret for an organization.")
	a.SetToken("index", "OrganizationActionSecret")
}

func (a *OrganizationActionSecretArgs) Annotate(ann infer.Annotator) {
	annotateActionSecretArgs(&a.ActionSecretArgs, ann)
	ann.Describe(&a.Organization, "Organization name.")
}

func annotateActionSecretArgs(a *ActionSecretArgs, ann infer.Annotator) {
	ann.Describe(&a.Name, "Actions secret name.")
	ann.Describe(&a.Value, "Actions secret value.")
}

func (RepositoryActionSecret) Create(ctx context.Context, req infer.CreateRequest[RepositoryActionSecretArgs]) (infer.CreateResponse[RepositoryActionSecretState], error) {
	if req.DryRun {
		return infer.CreateResponse[RepositoryActionSecretState]{ID: repositoryActionSecretID(req.Inputs.Owner, req.Inputs.Repository, req.Inputs.Name), Output: repositoryActionSecretStateFromArgs(req.Inputs, "")}, nil
	}

	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.CreateResponse[RepositoryActionSecretState]{}, err
	}
	_, err = client.CreateRepoActionSecret(req.Inputs.Owner, req.Inputs.Repository, forgejo.CreateSecretOption{Name: req.Inputs.Name, Data: req.Inputs.Value})
	if err != nil {
		return infer.CreateResponse[RepositoryActionSecretState]{}, err
	}

	secret, _, err := findRepositoryActionSecret(client, req.Inputs.Owner, req.Inputs.Repository, req.Inputs.Name)
	if err != nil {
		return infer.CreateResponse[RepositoryActionSecretState]{}, err
	}
	state := repositoryActionSecretStateFromAPI(req.Inputs, secret)
	return infer.CreateResponse[RepositoryActionSecretState]{ID: repositoryActionSecretID(state.Owner, state.Repository, state.Name), Output: state}, nil
}

func (RepositoryActionSecret) Read(ctx context.Context, req infer.ReadRequest[RepositoryActionSecretArgs, RepositoryActionSecretState]) (infer.ReadResponse[RepositoryActionSecretArgs, RepositoryActionSecretState], error) {
	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.ReadResponse[RepositoryActionSecretArgs, RepositoryActionSecretState]{}, err
	}

	secret, resp, err := findRepositoryActionSecret(client, req.State.Owner, req.State.Repository, req.State.Name)
	if isNotFound(resp) {
		return infer.ReadResponse[RepositoryActionSecretArgs, RepositoryActionSecretState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[RepositoryActionSecretArgs, RepositoryActionSecretState]{}, err
	}
	if secret == nil {
		return infer.ReadResponse[RepositoryActionSecretArgs, RepositoryActionSecretState]{}, nil
	}

	state := repositoryActionSecretStateFromAPI(req.State.RepositoryActionSecretArgs, secret)
	return infer.ReadResponse[RepositoryActionSecretArgs, RepositoryActionSecretState]{ID: repositoryActionSecretID(state.Owner, state.Repository, state.Name), Inputs: state.RepositoryActionSecretArgs, State: state}, nil
}

func (RepositoryActionSecret) Diff(_ context.Context, req infer.DiffRequest[RepositoryActionSecretArgs, RepositoryActionSecretState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	addReplaceDiff(diff, "owner", req.State.Owner != req.Inputs.Owner)
	addReplaceDiff(diff, "repository", req.State.Repository != req.Inputs.Repository)
	addReplaceDiff(diff, "name", req.State.Name != req.Inputs.Name)
	addUpdateDiff(diff, "value", req.State.Value != req.Inputs.Value)
	return p.DiffResponse{HasChanges: len(diff) > 0, DeleteBeforeReplace: true, DetailedDiff: diff}, nil
}

func (RepositoryActionSecret) Update(ctx context.Context, req infer.UpdateRequest[RepositoryActionSecretArgs, RepositoryActionSecretState]) (infer.UpdateResponse[RepositoryActionSecretState], error) {
	if req.DryRun {
		return infer.UpdateResponse[RepositoryActionSecretState]{Output: repositoryActionSecretStateFromArgs(req.Inputs, req.State.Created)}, nil
	}

	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.UpdateResponse[RepositoryActionSecretState]{}, err
	}
	_, err = client.CreateRepoActionSecret(req.Inputs.Owner, req.Inputs.Repository, forgejo.CreateSecretOption{Name: req.Inputs.Name, Data: req.Inputs.Value})
	if err != nil {
		return infer.UpdateResponse[RepositoryActionSecretState]{}, err
	}
	secret, _, err := findRepositoryActionSecret(client, req.Inputs.Owner, req.Inputs.Repository, req.Inputs.Name)
	if err != nil {
		return infer.UpdateResponse[RepositoryActionSecretState]{}, err
	}

	return infer.UpdateResponse[RepositoryActionSecretState]{Output: repositoryActionSecretStateFromAPI(req.Inputs, secret)}, nil
}

func (RepositoryActionSecret) Delete(ctx context.Context, req infer.DeleteRequest[RepositoryActionSecretState]) (infer.DeleteResponse, error) {
	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.DeleteResponse{}, err
	}

	resp, err := client.DeleteRepoActionSecret(req.State.Owner, req.State.Repository, req.State.Name)
	if isNotFound(resp) {
		return infer.DeleteResponse{}, nil
	}
	return infer.DeleteResponse{}, err
}

func (OrganizationActionSecret) Create(ctx context.Context, req infer.CreateRequest[OrganizationActionSecretArgs]) (infer.CreateResponse[OrganizationActionSecretState], error) {
	if req.DryRun {
		return infer.CreateResponse[OrganizationActionSecretState]{ID: organizationActionSecretID(req.Inputs.Organization, req.Inputs.Name), Output: organizationActionSecretStateFromArgs(req.Inputs, "")}, nil
	}

	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.CreateResponse[OrganizationActionSecretState]{}, err
	}
	_, err = client.CreateOrgActionSecret(req.Inputs.Organization, forgejo.CreateSecretOption{Name: req.Inputs.Name, Data: req.Inputs.Value})
	if err != nil {
		return infer.CreateResponse[OrganizationActionSecretState]{}, err
	}

	secret, _, err := findOrganizationActionSecret(client, req.Inputs.Organization, req.Inputs.Name)
	if err != nil {
		return infer.CreateResponse[OrganizationActionSecretState]{}, err
	}
	state := organizationActionSecretStateFromAPI(req.Inputs, secret)
	return infer.CreateResponse[OrganizationActionSecretState]{ID: organizationActionSecretID(state.Organization, state.Name), Output: state}, nil
}

func (OrganizationActionSecret) Read(ctx context.Context, req infer.ReadRequest[OrganizationActionSecretArgs, OrganizationActionSecretState]) (infer.ReadResponse[OrganizationActionSecretArgs, OrganizationActionSecretState], error) {
	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.ReadResponse[OrganizationActionSecretArgs, OrganizationActionSecretState]{}, err
	}

	secret, resp, err := findOrganizationActionSecret(client, req.State.Organization, req.State.Name)
	if isNotFound(resp) {
		return infer.ReadResponse[OrganizationActionSecretArgs, OrganizationActionSecretState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[OrganizationActionSecretArgs, OrganizationActionSecretState]{}, err
	}
	if secret == nil {
		return infer.ReadResponse[OrganizationActionSecretArgs, OrganizationActionSecretState]{}, nil
	}

	state := organizationActionSecretStateFromAPI(req.State.OrganizationActionSecretArgs, secret)
	return infer.ReadResponse[OrganizationActionSecretArgs, OrganizationActionSecretState]{ID: organizationActionSecretID(state.Organization, state.Name), Inputs: state.OrganizationActionSecretArgs, State: state}, nil
}

func (OrganizationActionSecret) Diff(_ context.Context, req infer.DiffRequest[OrganizationActionSecretArgs, OrganizationActionSecretState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	addReplaceDiff(diff, "organization", req.State.Organization != req.Inputs.Organization)
	addReplaceDiff(diff, "name", req.State.Name != req.Inputs.Name)
	addUpdateDiff(diff, "value", req.State.Value != req.Inputs.Value)
	return p.DiffResponse{HasChanges: len(diff) > 0, DeleteBeforeReplace: true, DetailedDiff: diff}, nil
}

func (OrganizationActionSecret) Update(ctx context.Context, req infer.UpdateRequest[OrganizationActionSecretArgs, OrganizationActionSecretState]) (infer.UpdateResponse[OrganizationActionSecretState], error) {
	if req.DryRun {
		return infer.UpdateResponse[OrganizationActionSecretState]{Output: organizationActionSecretStateFromArgs(req.Inputs, req.State.Created)}, nil
	}

	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.UpdateResponse[OrganizationActionSecretState]{}, err
	}
	_, err = client.CreateOrgActionSecret(req.Inputs.Organization, forgejo.CreateSecretOption{Name: req.Inputs.Name, Data: req.Inputs.Value})
	if err != nil {
		return infer.UpdateResponse[OrganizationActionSecretState]{}, err
	}
	secret, _, err := findOrganizationActionSecret(client, req.Inputs.Organization, req.Inputs.Name)
	if err != nil {
		return infer.UpdateResponse[OrganizationActionSecretState]{}, err
	}

	return infer.UpdateResponse[OrganizationActionSecretState]{Output: organizationActionSecretStateFromAPI(req.Inputs, secret)}, nil
}

func (OrganizationActionSecret) Delete(ctx context.Context, req infer.DeleteRequest[OrganizationActionSecretState]) (infer.DeleteResponse, error) {
	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.DeleteResponse{}, err
	}

	resp, err := client.DeleteOrgActionSecret(req.State.Organization, req.State.Name)
	if isNotFound(resp) {
		return infer.DeleteResponse{}, nil
	}
	return infer.DeleteResponse{}, err
}

func findRepositoryActionSecret(client *forgejo.Client, owner, repo, name string) (*forgejo.Secret, *forgejo.Response, error) {
	secrets, resp, err := client.ListRepoActionSecret(owner, repo, forgejo.ListRepoActionSecretOption{ListOptions: forgejo.ListOptions{Page: -1}})
	if err != nil {
		return nil, resp, err
	}
	for _, secret := range secrets {
		if secret.Name == name {
			return secret, resp, nil
		}
	}
	return nil, resp, nil
}

func findOrganizationActionSecret(client *forgejo.Client, org, name string) (*forgejo.Secret, *forgejo.Response, error) {
	secrets, resp, err := client.ListOrgActionSecret(org, forgejo.ListOrgActionSecretOption{ListOptions: forgejo.ListOptions{Page: -1}})
	if err != nil {
		return nil, resp, err
	}
	for _, secret := range secrets {
		if secret.Name == name {
			return secret, resp, nil
		}
	}
	return nil, resp, nil
}

func repositoryActionSecretStateFromAPI(args RepositoryActionSecretArgs, secret *forgejo.Secret) RepositoryActionSecretState {
	if secret == nil {
		return repositoryActionSecretStateFromArgs(args, "")
	}
	args.Name = secret.Name
	return repositoryActionSecretStateFromArgs(args, formatForgejoTime(secret.Created))
}

func repositoryActionSecretStateFromArgs(args RepositoryActionSecretArgs, created string) RepositoryActionSecretState {
	return RepositoryActionSecretState{RepositoryActionSecretArgs: args, Created: created}
}

func repositoryActionSecretID(owner, repo, name string) string {
	return fmt.Sprintf("%s/%s/%s", owner, repo, name)
}

func organizationActionSecretStateFromAPI(args OrganizationActionSecretArgs, secret *forgejo.Secret) OrganizationActionSecretState {
	if secret == nil {
		return organizationActionSecretStateFromArgs(args, "")
	}
	args.Name = secret.Name
	return organizationActionSecretStateFromArgs(args, formatForgejoTime(secret.Created))
}

func organizationActionSecretStateFromArgs(args OrganizationActionSecretArgs, created string) OrganizationActionSecretState {
	return OrganizationActionSecretState{OrganizationActionSecretArgs: args, Created: created}
}

func organizationActionSecretID(org, name string) string {
	return fmt.Sprintf("%s/%s", org, name)
}

func formatForgejoTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(time.RFC3339)
}
