package provider

import (
	"context"
	"fmt"
	"strings"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v3"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

type Repository struct{}

type RepositoryArgs struct {
	Name          string `pulumi:"name"`
	Owner         string `pulumi:"owner,optional"`
	Description   string `pulumi:"description,optional"`
	Private       bool   `pulumi:"private,optional"`
	DefaultBranch string `pulumi:"defaultBranch,optional"`
	Website       string `pulumi:"website,optional"`
	Issues        bool   `pulumi:"issues,optional"`
	Wiki          bool   `pulumi:"wiki,optional"`
	Projects      bool   `pulumi:"projects,optional"`
	Template      bool   `pulumi:"template,optional"`
}

type RepositoryState struct {
	RepositoryArgs
	FullName string `pulumi:"fullName"`
	HTMLURL  string `pulumi:"htmlUrl"`
	SSHURL   string `pulumi:"sshUrl"`
	CloneURL string `pulumi:"cloneUrl"`
}

func (r *Repository) Annotate(a infer.Annotator) {
	a.Describe(r, "A Forgejo repository.")
	a.SetToken("index", "Repository")
}

func (a *RepositoryArgs) Annotate(ann infer.Annotator) {
	ann.Describe(&a.Name, "Repository name.")
	ann.Describe(&a.Owner, "Repository owner. Leave empty to create a repository for the authenticated user; set an organization name to create an organization repository.")
	ann.Describe(&a.Description, "Repository description.")
	ann.Describe(&a.Private, "Whether the repository is private.")
	ann.Describe(&a.DefaultBranch, "Default branch name. Leave empty to use Forgejo's default.")
	ann.Describe(&a.Website, "Repository website URL.")
	ann.Describe(&a.Issues, "Whether the repository issue tracker is enabled.")
	ann.Describe(&a.Wiki, "Whether the repository wiki is enabled.")
	ann.Describe(&a.Projects, "Whether repository projects are enabled.")
	ann.Describe(&a.Template, "Whether the repository can be used as a template.")
	ann.SetDefault(&a.Issues, true)
	ann.SetDefault(&a.Wiki, true)
	ann.SetDefault(&a.Projects, true)
}

func (Repository) Create(ctx context.Context, req infer.CreateRequest[RepositoryArgs]) (infer.CreateResponse[RepositoryState], error) {
	if req.DryRun {
		return infer.CreateResponse[RepositoryState]{
			ID:     repositoryID(req.Inputs.Owner, req.Inputs.Name),
			Output: repositoryStateFromArgs(req.Inputs, nil),
		}, nil
	}

	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.CreateResponse[RepositoryState]{}, err
	}

	create := forgejo.CreateRepoOption{
		Name:          req.Inputs.Name,
		Description:   req.Inputs.Description,
		Private:       req.Inputs.Private,
		DefaultBranch: req.Inputs.DefaultBranch,
		Template:      req.Inputs.Template,
	}
	repo, err := createRepository(client, req.Inputs.Owner, create)
	if err != nil {
		return infer.CreateResponse[RepositoryState]{}, err
	}
	repo, err = editRepository(client, repo, repositoryEditOption(req.Inputs, false))
	if err != nil {
		return infer.CreateResponse[RepositoryState]{}, err
	}

	state := repositoryStateFromAPI(repo)
	return infer.CreateResponse[RepositoryState]{ID: repositoryID(state.Owner, state.Name), Output: state}, nil
}

func (Repository) Read(ctx context.Context, req infer.ReadRequest[RepositoryArgs, RepositoryState]) (infer.ReadResponse[RepositoryArgs, RepositoryState], error) {
	owner, name, err := repositoryOwnerName(req.ID, req.State)
	if err != nil {
		return infer.ReadResponse[RepositoryArgs, RepositoryState]{}, err
	}

	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.ReadResponse[RepositoryArgs, RepositoryState]{}, err
	}

	repo, resp, err := client.GetRepo(owner, name)
	if isNotFound(resp) {
		return infer.ReadResponse[RepositoryArgs, RepositoryState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[RepositoryArgs, RepositoryState]{}, err
	}

	state := repositoryStateFromAPI(repo)
	return infer.ReadResponse[RepositoryArgs, RepositoryState]{
		ID:     repositoryID(state.Owner, state.Name),
		Inputs: state.RepositoryArgs,
		State:  state,
	}, nil
}

func (Repository) Diff(_ context.Context, req infer.DiffRequest[RepositoryArgs, RepositoryState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	if req.State.Name != req.Inputs.Name {
		diff["name"] = p.PropertyDiff{Kind: p.UpdateReplace, InputDiff: true}
	}
	if req.Inputs.Owner != "" && req.State.Owner != req.Inputs.Owner {
		diff["owner"] = p.PropertyDiff{Kind: p.UpdateReplace, InputDiff: true}
	}
	addUpdateDiff(diff, "description", req.State.Description != req.Inputs.Description)
	addUpdateDiff(diff, "private", req.State.Private != req.Inputs.Private)
	addUpdateDiff(diff, "defaultBranch", req.Inputs.DefaultBranch != "" && req.State.DefaultBranch != req.Inputs.DefaultBranch)
	addUpdateDiff(diff, "website", req.State.Website != req.Inputs.Website)
	addUpdateDiff(diff, "issues", req.State.Issues != req.Inputs.Issues)
	addUpdateDiff(diff, "wiki", req.State.Wiki != req.Inputs.Wiki)
	addUpdateDiff(diff, "projects", req.State.Projects != req.Inputs.Projects)
	addUpdateDiff(diff, "template", req.State.Template != req.Inputs.Template)

	return p.DiffResponse{HasChanges: len(diff) > 0, DeleteBeforeReplace: true, DetailedDiff: diff}, nil
}

func (Repository) Update(ctx context.Context, req infer.UpdateRequest[RepositoryArgs, RepositoryState]) (infer.UpdateResponse[RepositoryState], error) {
	if req.DryRun {
		return infer.UpdateResponse[RepositoryState]{Output: repositoryStateFromArgs(req.Inputs, &req.State)}, nil
	}

	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.UpdateResponse[RepositoryState]{}, err
	}
	owner, name, err := repositoryOwnerName(req.ID, req.State)
	if err != nil {
		return infer.UpdateResponse[RepositoryState]{}, err
	}

	repo, _, err := client.EditRepo(owner, name, repositoryEditOption(req.Inputs, true))
	if err != nil {
		return infer.UpdateResponse[RepositoryState]{}, err
	}

	return infer.UpdateResponse[RepositoryState]{Output: repositoryStateFromAPI(repo)}, nil
}

func (Repository) Delete(ctx context.Context, req infer.DeleteRequest[RepositoryState]) (infer.DeleteResponse, error) {
	owner, name, err := repositoryOwnerName(req.ID, req.State)
	if err != nil {
		return infer.DeleteResponse{}, err
	}

	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.DeleteResponse{}, err
	}

	resp, err := client.DeleteRepo(owner, name)
	if isNotFound(resp) {
		return infer.DeleteResponse{}, nil
	}
	return infer.DeleteResponse{}, err
}

func repositoryStateFromAPI(repo *forgejo.Repository) RepositoryState {
	owner := ""
	if repo.Owner != nil {
		owner = repo.Owner.UserName
	}
	if repo.FullName != "" && owner == "" {
		owner, _, _ = strings.Cut(repo.FullName, "/")
	}

	return RepositoryState{
		RepositoryArgs: RepositoryArgs{
			Name:          repo.Name,
			Owner:         owner,
			Description:   repo.Description,
			Private:       repo.Private,
			DefaultBranch: repo.DefaultBranch,
			Website:       repo.Website,
			Issues:        repo.HasIssues,
			Wiki:          repo.HasWiki,
			Projects:      repo.HasProjects,
			Template:      repo.Template,
		},
		FullName: repo.FullName,
		HTMLURL:  repo.HTMLURL,
		SSHURL:   repo.SSHURL,
		CloneURL: repo.CloneURL,
	}
}

func createRepository(client *forgejo.Client, owner string, opt forgejo.CreateRepoOption) (*forgejo.Repository, error) {
	if strings.TrimSpace(owner) == "" {
		repo, _, err := client.CreateRepo(opt)
		return repo, err
	}

	user, _, err := client.GetMyUserInfo()
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("current Forgejo user response was empty")
	}
	if user.UserName == owner {
		repo, _, err := client.CreateRepo(opt)
		return repo, err
	}
	repo, _, err := client.CreateOrgRepo(owner, opt)
	return repo, err
}

func editRepository(client *forgejo.Client, repo *forgejo.Repository, opt forgejo.EditRepoOption) (*forgejo.Repository, error) {
	if repo == nil {
		return nil, fmt.Errorf("Forgejo repository response was empty")
	}
	state := repositoryStateFromAPI(repo)
	updated, _, err := client.EditRepo(state.Owner, state.Name, opt)
	return updated, err
}

func repositoryEditOption(args RepositoryArgs, includeDefaultBranch bool) forgejo.EditRepoOption {
	opt := forgejo.EditRepoOption{
		Description: &args.Description,
		Private:     &args.Private,
		Website:     &args.Website,
		HasIssues:   &args.Issues,
		HasWiki:     &args.Wiki,
		HasProjects: &args.Projects,
		Template:    &args.Template,
	}
	if includeDefaultBranch && args.DefaultBranch != "" {
		opt.DefaultBranch = &args.DefaultBranch
	}
	return opt
}

func repositoryStateFromArgs(args RepositoryArgs, previous *RepositoryState) RepositoryState {
	state := RepositoryState{RepositoryArgs: args, FullName: repositoryID(args.Owner, args.Name)}
	if previous != nil {
		state.HTMLURL = previous.HTMLURL
		state.SSHURL = previous.SSHURL
		state.CloneURL = previous.CloneURL
	}
	return state
}

func repositoryID(owner, name string) string {
	if owner == "" {
		return name
	}
	return owner + "/" + name
}

func repositoryOwnerName(id string, state RepositoryState) (string, string, error) {
	owner, name := state.Owner, state.Name
	if owner == "" || name == "" {
		parts := strings.SplitN(id, "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return "", "", fmt.Errorf("repository ID must have the form owner/name, got %q", id)
		}
		owner, name = parts[0], parts[1]
	}
	return owner, name, nil
}

func addUpdateDiff(diff map[string]p.PropertyDiff, field string, changed bool) {
	if changed {
		diff[field] = p.PropertyDiff{Kind: p.Update, InputDiff: true}
	}
}
