package provider

import (
	"context"
	"fmt"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v3"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

type RepositoryActionVariable struct{}

type RepositoryActionVariableArgs struct {
	Owner      string `pulumi:"owner"`
	Repository string `pulumi:"repository"`
	Name       string `pulumi:"name"`
	Value      string `pulumi:"value"`
}

type RepositoryActionVariableState struct {
	RepositoryActionVariableArgs
	OwnerID int64 `pulumi:"ownerId"`
	RepoID  int64 `pulumi:"repoId"`
}

func (v *RepositoryActionVariable) Annotate(a infer.Annotator) {
	a.Describe(v, "A Forgejo Actions variable for a repository.")
	a.SetToken("index", "RepositoryActionVariable")
}

func (a *RepositoryActionVariableArgs) Annotate(ann infer.Annotator) {
	ann.Describe(&a.Owner, "Repository owner.")
	ann.Describe(&a.Repository, "Repository name.")
	ann.Describe(&a.Name, "Actions variable name.")
	ann.Describe(&a.Value, "Actions variable value.")
}

func (RepositoryActionVariable) Create(ctx context.Context, req infer.CreateRequest[RepositoryActionVariableArgs]) (infer.CreateResponse[RepositoryActionVariableState], error) {
	if req.DryRun {
		return infer.CreateResponse[RepositoryActionVariableState]{ID: repositoryActionVariableID(req.Inputs.Owner, req.Inputs.Repository, req.Inputs.Name), Output: repositoryActionVariableStateFromArgs(req.Inputs, 0, 0)}, nil
	}

	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.CreateResponse[RepositoryActionVariableState]{}, err
	}

	_, err = client.CreateRepoActionVariable(req.Inputs.Owner, req.Inputs.Repository, forgejo.CreateVariableOption{Name: req.Inputs.Name, Data: req.Inputs.Value})
	if err != nil {
		return infer.CreateResponse[RepositoryActionVariableState]{}, err
	}
	variable, _, err := client.GetRepoActionVariable(req.Inputs.Owner, req.Inputs.Repository, req.Inputs.Name)
	if err != nil {
		return infer.CreateResponse[RepositoryActionVariableState]{}, err
	}

	state := repositoryActionVariableStateFromAPI(req.Inputs.Owner, req.Inputs.Repository, variable)
	return infer.CreateResponse[RepositoryActionVariableState]{ID: repositoryActionVariableID(state.Owner, state.Repository, state.Name), Output: state}, nil
}

func (RepositoryActionVariable) Read(ctx context.Context, req infer.ReadRequest[RepositoryActionVariableArgs, RepositoryActionVariableState]) (infer.ReadResponse[RepositoryActionVariableArgs, RepositoryActionVariableState], error) {
	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.ReadResponse[RepositoryActionVariableArgs, RepositoryActionVariableState]{}, err
	}

	variable, resp, err := client.GetRepoActionVariable(req.State.Owner, req.State.Repository, req.State.Name)
	if isNotFound(resp) {
		return infer.ReadResponse[RepositoryActionVariableArgs, RepositoryActionVariableState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[RepositoryActionVariableArgs, RepositoryActionVariableState]{}, err
	}

	state := repositoryActionVariableStateFromAPI(req.State.Owner, req.State.Repository, variable)
	return infer.ReadResponse[RepositoryActionVariableArgs, RepositoryActionVariableState]{ID: repositoryActionVariableID(state.Owner, state.Repository, state.Name), Inputs: state.RepositoryActionVariableArgs, State: state}, nil
}

func (RepositoryActionVariable) Diff(_ context.Context, req infer.DiffRequest[RepositoryActionVariableArgs, RepositoryActionVariableState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	addReplaceDiff(diff, "owner", req.State.Owner != req.Inputs.Owner)
	addReplaceDiff(diff, "repository", req.State.Repository != req.Inputs.Repository)
	addReplaceDiff(diff, "name", req.State.Name != req.Inputs.Name)
	addUpdateDiff(diff, "value", req.State.Value != req.Inputs.Value)
	return p.DiffResponse{HasChanges: len(diff) > 0, DeleteBeforeReplace: true, DetailedDiff: diff}, nil
}

func (RepositoryActionVariable) Update(ctx context.Context, req infer.UpdateRequest[RepositoryActionVariableArgs, RepositoryActionVariableState]) (infer.UpdateResponse[RepositoryActionVariableState], error) {
	if req.DryRun {
		return infer.UpdateResponse[RepositoryActionVariableState]{Output: repositoryActionVariableStateFromArgs(req.Inputs, req.State.OwnerID, req.State.RepoID)}, nil
	}

	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.UpdateResponse[RepositoryActionVariableState]{}, err
	}
	_, err = client.UpdateRepoActionVariable(req.State.Owner, req.State.Repository, req.State.Name, forgejo.CreateVariableOption{Name: req.Inputs.Name, Data: req.Inputs.Value})
	if err != nil {
		return infer.UpdateResponse[RepositoryActionVariableState]{}, err
	}
	variable, _, err := client.GetRepoActionVariable(req.Inputs.Owner, req.Inputs.Repository, req.Inputs.Name)
	if err != nil {
		return infer.UpdateResponse[RepositoryActionVariableState]{}, err
	}

	return infer.UpdateResponse[RepositoryActionVariableState]{Output: repositoryActionVariableStateFromAPI(req.Inputs.Owner, req.Inputs.Repository, variable)}, nil
}

func (RepositoryActionVariable) Delete(ctx context.Context, req infer.DeleteRequest[RepositoryActionVariableState]) (infer.DeleteResponse, error) {
	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.DeleteResponse{}, err
	}

	resp, err := client.DeleteRepoActionVariable(req.State.Owner, req.State.Repository, req.State.Name)
	if isNotFound(resp) {
		return infer.DeleteResponse{}, nil
	}
	return infer.DeleteResponse{}, err
}

func repositoryActionVariableStateFromAPI(owner, repo string, variable *forgejo.ActionVariable) RepositoryActionVariableState {
	return repositoryActionVariableStateFromArgs(RepositoryActionVariableArgs{Owner: owner, Repository: repo, Name: variable.Name, Value: variable.Data}, variable.OwnerID, variable.RepoID)
}

func repositoryActionVariableStateFromArgs(args RepositoryActionVariableArgs, ownerID, repoID int64) RepositoryActionVariableState {
	return RepositoryActionVariableState{RepositoryActionVariableArgs: args, OwnerID: ownerID, RepoID: repoID}
}

func repositoryActionVariableID(owner, repo, name string) string {
	return fmt.Sprintf("%s/%s/%s", owner, repo, name)
}
