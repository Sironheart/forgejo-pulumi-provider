package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v3"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

type Repository struct{}

type RepositoryArgs struct {
	Name                 string                    `pulumi:"name"`
	Owner                string                    `pulumi:"owner,optional"`
	Description          string                    `pulumi:"description,optional"`
	Private              bool                      `pulumi:"private,optional"`
	DefaultBranch        string                    `pulumi:"defaultBranch,optional"`
	Website              string                    `pulumi:"website,optional"`
	Template             bool                      `pulumi:"template,optional"`
	InitializeWithReadme bool                      `pulumi:"initializeWithReadme,optional"`
	Gitignores           string                    `pulumi:"gitignores,optional"`
	License              string                    `pulumi:"license,optional"`
	Readme               string                    `pulumi:"readme,optional"`
	IssueLabels          string                    `pulumi:"issueLabels,optional"`
	TrustModel           string                    `pulumi:"trustModel,optional"`
	ObjectFormatName     string                    `pulumi:"objectFormatName,optional"`
	Topics               []string                  `pulumi:"topics,optional"`
	Settings             *RepositorySettingsConfig `pulumi:"settings,optional"`
}

type RepositoryState struct {
	RepositoryArgs
	FullName string `pulumi:"fullName"`
	HTMLURL  string `pulumi:"htmlUrl"`
	SSHURL   string `pulumi:"sshUrl"`
	CloneURL string `pulumi:"cloneUrl"`
}

type legacyRepositoryState struct {
	RepositoryArgs
	Issues   *bool  `pulumi:"issues,optional"`
	Wiki     *bool  `pulumi:"wiki,optional"`
	Projects *bool  `pulumi:"projects,optional"`
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
	ann.Describe(&a.Template, "Whether the repository can be used as a template.")
	ann.Describe(&a.InitializeWithReadme, "Whether Forgejo should initialize the repository with a README on creation. This property is only used when the repository is first created; changing it later has no effect.")
	ann.Describe(&a.Gitignores, "Comma-separated gitignore template names. Requires initializeWithReadme. Only used on creation; changing it later has no effect.")
	ann.Describe(&a.License, "License template name. Requires initializeWithReadme. Only used on creation; changing it later has no effect.")
	ann.Describe(&a.Readme, "README template name, not literal contents. Requires initializeWithReadme; Forgejo defaults to Default. Only used on creation; changing it later has no effect.")
	ann.Describe(&a.IssueLabels, "Initial issue-label template set. Only used on creation; changing it later has no effect.")
	ann.Describe(&a.TrustModel, "Signature trust model: default, collaborator, committer, or collaboratorcommitter. Only used on creation; changing it later has no effect.")
	ann.Describe(&a.ObjectFormatName, "Git object format: sha1 (Forgejo's default) or sha256. Only used on creation; changing it later has no effect.")
	ann.Describe(&a.Topics, "Repository topics.")
	ann.Describe(&a.Settings, "Optional repository unit, merge, wiki, issue tracker, and pull-mirror settings to manage with this repository.")
}

func (Repository) StateMigrations(context.Context) []infer.StateMigrationFunc[RepositoryState] {
	return []infer.StateMigrationFunc[RepositoryState]{
		infer.StateMigration(migrateLegacyRepositoryState),
	}
}

func migrateLegacyRepositoryState(_ context.Context, legacy legacyRepositoryState) (infer.MigrationResult[RepositoryState], error) {
	if legacy.Issues == nil && legacy.Wiki == nil && legacy.Projects == nil {
		return infer.MigrationResult[RepositoryState]{}, nil
	}

	state := RepositoryState{
		RepositoryArgs: legacy.RepositoryArgs,
		FullName:       legacy.FullName,
		HTMLURL:        legacy.HTMLURL,
		SSHURL:         legacy.SSHURL,
		CloneURL:       legacy.CloneURL,
	}
	if state.Settings == nil {
		state.Settings = &RepositorySettingsConfig{}
	}
	if legacy.Issues != nil && state.Settings.Issues == nil {
		state.Settings.Issues = legacy.Issues
	}
	if legacy.Wiki != nil && state.Settings.Wiki == nil {
		state.Settings.Wiki = legacy.Wiki
	}
	if legacy.Projects != nil && state.Settings.Projects == nil {
		state.Settings.Projects = legacy.Projects
	}

	return infer.MigrationResult[RepositoryState]{Result: &state}, nil
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

	api, err := forgejoAPIClientFromConfig(ctx)
	if err != nil {
		return infer.CreateResponse[RepositoryState]{}, err
	}
	create := repositoryCreateOption{
		CreateRepoOption: forgejo.CreateRepoOption{
			Name:          req.Inputs.Name,
			Description:   req.Inputs.Description,
			Private:       req.Inputs.Private,
			DefaultBranch: req.Inputs.DefaultBranch,
			Template:      req.Inputs.Template,
			AutoInit:      req.Inputs.InitializeWithReadme,
			Gitignores:    req.Inputs.Gitignores,
			License:       req.Inputs.License,
			Readme:        req.Inputs.Readme,
			IssueLabels:   req.Inputs.IssueLabels,
			TrustModel:    forgejo.TrustModel(req.Inputs.TrustModel),
		},
		ObjectFormatName: req.Inputs.ObjectFormatName,
	}
	repo, err := createRepository(client, api, req.Inputs.Owner, create)
	if err != nil {
		return infer.CreateResponse[RepositoryState]{}, err
	}
	repo, err = editRepository(api, repo, repositoryEditOption(req.Inputs, false, repo))
	if err != nil {
		return infer.CreateResponse[RepositoryState]{}, err
	}

	state := repositoryStateFromAPI(repo, req.Inputs)
	if req.Inputs.Topics != nil {
		_, err = client.SetRepoTopics(state.Owner, state.Name, req.Inputs.Topics)
		if err != nil {
			return infer.CreateResponse[RepositoryState]{}, err
		}
		state.Topics = req.Inputs.Topics
	}
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

	api, err := forgejoAPIClientFromConfig(ctx)
	if err != nil {
		return infer.ReadResponse[RepositoryArgs, RepositoryState]{}, err
	}
	repo, resp, err := api.getRepository(owner, name)
	if isNotFound(resp) {
		return infer.ReadResponse[RepositoryArgs, RepositoryState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[RepositoryArgs, RepositoryState]{}, err
	}

	template := req.State.RepositoryArgs
	template.Settings = repositorySettingsTemplate(req.Inputs.Settings, req.State.Settings)
	template.Topics = req.Inputs.Topics
	state := repositoryStateFromAPI(repo, template)
	if req.Inputs.Topics != nil {
		topics, _, err := client.ListRepoTopics(owner, name, forgejo.ListRepoTopicsOptions{})
		if err != nil {
			return infer.ReadResponse[RepositoryArgs, RepositoryState]{}, err
		}
		state.Topics = topics
	}
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
	addUpdateDiff(diff, "template", req.State.Template != req.Inputs.Template)
	addUpdateDiff(diff, "topics", !equalStringSlice(req.State.Topics, req.Inputs.Topics))
	addUpdateDiff(diff, "settings", !equalRepositorySettingsConfigPtr(req.State.Settings, req.Inputs.Settings))

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

	api, err := forgejoAPIClientFromConfig(ctx)
	if err != nil {
		return infer.UpdateResponse[RepositoryState]{}, err
	}
	var current *repositoryAPIResponse
	if req.Inputs.Settings != nil {
		current, _, err = api.getRepository(owner, name)
		if err != nil {
			return infer.UpdateResponse[RepositoryState]{}, err
		}
	}
	repo, _, err := api.editRepository(owner, name, repositoryEditOption(req.Inputs, true, current))
	if err != nil {
		return infer.UpdateResponse[RepositoryState]{}, err
	}

	if req.Inputs.Topics != nil {
		_, err = client.SetRepoTopics(owner, name, req.Inputs.Topics)
		if err != nil {
			return infer.UpdateResponse[RepositoryState]{}, err
		}
	}

	template := req.State.RepositoryArgs
	template.Settings = req.Inputs.Settings
	template.Topics = req.Inputs.Topics
	state := repositoryStateFromAPI(repo, template)
	return infer.UpdateResponse[RepositoryState]{Output: state}, nil
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

func repositoryStateFromAPI(repo *repositoryAPIResponse, template ...RepositoryArgs) RepositoryState {
	owner := ""
	if repo.Owner != nil {
		owner = repo.Owner.UserName
	}
	if repo.FullName != "" && owner == "" {
		owner, _, _ = strings.Cut(repo.FullName, "/")
	}

	var args RepositoryArgs
	if len(template) > 0 {
		args = template[0]
	}
	args.Name = repo.Name
	args.Owner = owner
	args.Description = repo.Description
	args.Private = repo.Private
	args.DefaultBranch = repo.DefaultBranch
	args.Website = repo.Website
	args.Template = repo.Template
	if repo.ObjectFormatName != "" {
		args.ObjectFormatName = repo.ObjectFormatName
	}
	if args.Settings != nil {
		settings := repositorySettingsConfigFromAPI(*args.Settings, repo)
		args.Settings = &settings
	}

	return RepositoryState{
		RepositoryArgs: args,
		FullName:       repo.FullName,
		HTMLURL:        repo.HTMLURL,
		SSHURL:         repo.SSHURL,
		CloneURL:       repo.CloneURL,
	}
}

func createRepository(client *forgejo.Client, api *forgejoAPIClient, owner string, opt repositoryCreateOption) (*repositoryAPIResponse, error) {
	if err := opt.Validate(client); err != nil {
		return nil, err
	}
	path := "/user/repos"
	if strings.TrimSpace(owner) != "" {
		user, _, err := client.GetMyUserInfo()
		if err != nil {
			return nil, err
		}
		if user == nil {
			return nil, fmt.Errorf("current Forgejo user response was empty")
		}
		if user.UserName != owner {
			path = "/orgs/" + url.PathEscape(owner) + "/repos"
		}
	}
	repo := new(repositoryAPIResponse)
	_, err := api.do(http.MethodPost, path, opt, repo)
	return repo, err
}

func editRepository(client *forgejoAPIClient, repo *repositoryAPIResponse, opt forgejo.EditRepoOption) (*repositoryAPIResponse, error) {
	if repo == nil || repo.Name == "" {
		return nil, fmt.Errorf("forgejo repository response was empty")
	}
	state := repositoryStateFromAPI(repo)
	updated, _, err := client.editRepository(state.Owner, state.Name, opt)
	return updated, err
}

func repositoryEditOption(args RepositoryArgs, includeDefaultBranch bool, repo *repositoryAPIResponse) forgejo.EditRepoOption {
	var opt forgejo.EditRepoOption
	if args.Settings != nil {
		opt = repositorySettingsConfigEditOption(*args.Settings, repo)
	}
	opt.Description = &args.Description
	opt.Private = &args.Private
	opt.Website = &args.Website
	opt.Template = &args.Template
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

func equalStringSlice(a, b []string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	a = slices.Clone(a)
	b = slices.Clone(b)
	slices.Sort(a)
	slices.Sort(b)
	return slices.Equal(a, b)
}
