package provider

import (
	"context"
	"fmt"
	"strings"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v3"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

type RepositorySettings struct{}

type RepositorySettingsConfig struct {
	Issues                                          *bool   `pulumi:"issues,optional"`
	PullRequests                                    *bool   `pulumi:"pullRequests,optional"`
	Wiki                                            *bool   `pulumi:"wiki,optional"`
	Projects                                        *bool   `pulumi:"projects,optional"`
	Releases                                        *bool   `pulumi:"releases,optional"`
	Packages                                        *bool   `pulumi:"packages,optional"`
	Actions                                         *bool   `pulumi:"actions,optional"`
	ExternalWikiURL                                 *string `pulumi:"externalWikiUrl,optional"`
	GloballyEditableWiki                            *bool   `pulumi:"globallyEditableWiki,optional"`
	WikiBranch                                      *string `pulumi:"wikiBranch,optional"`
	InternalTrackerEnableTimeTracker                *bool   `pulumi:"internalTrackerEnableTimeTracker,optional"`
	InternalTrackerAllowOnlyContributorsToTrackTime *bool   `pulumi:"internalTrackerAllowOnlyContributorsToTrackTime,optional"`
	InternalTrackerEnableIssueDependencies          *bool   `pulumi:"internalTrackerEnableIssueDependencies,optional"`
	ExternalTrackerURL                              *string `pulumi:"externalTrackerUrl,optional"`
	ExternalTrackerFormat                           *string `pulumi:"externalTrackerFormat,optional"`
	ExternalTrackerStyle                            *string `pulumi:"externalTrackerStyle,optional"`
	ExternalTrackerRegexPattern                     *string `pulumi:"externalTrackerRegexPattern,optional"`
}

type RepositorySettingsArgs struct {
	RepositorySettingsConfig

	Owner      string `pulumi:"owner"`
	Repository string `pulumi:"repository"`
}

type RepositorySettingsState struct {
	RepositorySettingsArgs
}

func (s *RepositorySettings) Annotate(a infer.Annotator) {
	a.Describe(s, "Settings for enabled Forgejo repository units and their wiki or issue tracker configuration.")
	a.SetToken("index", "RepositorySettings")
}

func (a *RepositorySettingsArgs) Annotate(ann infer.Annotator) {
	ann.Describe(&a.Owner, "Repository owner.")
	ann.Describe(&a.Repository, "Repository name.")
	annotateRepositorySettingsConfig(&a.RepositorySettingsConfig, ann)
}

func (a *RepositorySettingsConfig) Annotate(ann infer.Annotator) {
	annotateRepositorySettingsConfig(a, ann)
}

func annotateRepositorySettingsConfig(a *RepositorySettingsConfig, ann infer.Annotator) {
	ann.Describe(&a.Issues, "Whether the issue tracker unit is enabled. Leave unset to avoid managing it.")
	ann.Describe(&a.PullRequests, "Whether the pull request unit is enabled. Leave unset to avoid managing it.")
	ann.Describe(&a.Wiki, "Whether the wiki unit is enabled. Leave unset to avoid managing it.")
	ann.Describe(&a.Projects, "Whether the projects unit is enabled. Leave unset to avoid managing it.")
	ann.Describe(&a.Releases, "Whether the releases unit is enabled. Leave unset to avoid managing it.")
	ann.Describe(&a.Packages, "Whether the packages unit is enabled. Leave unset to avoid managing it.")
	ann.Describe(&a.Actions, "Whether the actions unit is enabled. Leave unset to avoid managing it.")
	ann.Describe(&a.ExternalWikiURL, "External wiki URL. Setting this also enables the wiki unless wiki is explicitly set.")
	ann.Describe(&a.GloballyEditableWiki, "Whether the wiki is globally editable.")
	ann.Describe(&a.WikiBranch, "Branch used for the repository wiki.")
	ann.Describe(&a.InternalTrackerEnableTimeTracker, "Whether the internal issue tracker has time tracking enabled.")
	ann.Describe(&a.InternalTrackerAllowOnlyContributorsToTrackTime, "Whether only contributors may track time in the internal issue tracker.")
	ann.Describe(&a.InternalTrackerEnableIssueDependencies, "Whether issue dependencies are enabled in the internal issue tracker.")
	ann.Describe(&a.ExternalTrackerURL, "External issue tracker URL. Setting this also enables issues unless issues is explicitly set.")
	ann.Describe(&a.ExternalTrackerFormat, "External issue tracker URL format. Forgejo supports placeholders such as {user}, {repo}, and {index}.")
	ann.Describe(&a.ExternalTrackerStyle, "External issue tracker number style, for example numeric or alphanumeric.")
	ann.Describe(&a.ExternalTrackerRegexPattern, "External issue tracker regular expression pattern.")
}

func (RepositorySettings) Create(ctx context.Context, req infer.CreateRequest[RepositorySettingsArgs]) (infer.CreateResponse[RepositorySettingsState], error) {
	if req.DryRun {
		return infer.CreateResponse[RepositorySettingsState]{ID: repositorySettingsID(req.Inputs.Owner, req.Inputs.Repository), Output: repositorySettingsStateFromArgs(req.Inputs)}, nil
	}

	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.CreateResponse[RepositorySettingsState]{}, err
	}
	repo, _, err := client.GetRepo(req.Inputs.Owner, req.Inputs.Repository)
	if err != nil {
		return infer.CreateResponse[RepositorySettingsState]{}, err
	}
	updated, _, err := client.EditRepo(req.Inputs.Owner, req.Inputs.Repository, repositorySettingsEditOption(req.Inputs, repo))
	if err != nil {
		return infer.CreateResponse[RepositorySettingsState]{}, err
	}

	state := repositorySettingsStateFromAPI(req.Inputs, updated)
	return infer.CreateResponse[RepositorySettingsState]{ID: repositorySettingsID(state.Owner, state.Repository), Output: state}, nil
}

func (RepositorySettings) Read(ctx context.Context, req infer.ReadRequest[RepositorySettingsArgs, RepositorySettingsState]) (infer.ReadResponse[RepositorySettingsArgs, RepositorySettingsState], error) {
	owner, repoName, err := repositorySettingsParts(req.ID, req.State)
	if err != nil {
		return infer.ReadResponse[RepositorySettingsArgs, RepositorySettingsState]{}, err
	}
	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.ReadResponse[RepositorySettingsArgs, RepositorySettingsState]{}, err
	}

	repo, resp, err := client.GetRepo(owner, repoName)
	if isNotFound(resp) {
		return infer.ReadResponse[RepositorySettingsArgs, RepositorySettingsState]{}, nil
	}
	if err != nil {
		return infer.ReadResponse[RepositorySettingsArgs, RepositorySettingsState]{}, err
	}

	state := repositorySettingsStateFromAPI(repositorySettingsReadArgs(owner, repoName, req), repo)
	return infer.ReadResponse[RepositorySettingsArgs, RepositorySettingsState]{ID: repositorySettingsID(state.Owner, state.Repository), Inputs: state.RepositorySettingsArgs, State: state}, nil
}

func (RepositorySettings) Diff(_ context.Context, req infer.DiffRequest[RepositorySettingsArgs, RepositorySettingsState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	addReplaceDiff(diff, "owner", req.State.Owner != req.Inputs.Owner)
	addReplaceDiff(diff, "repository", req.State.Repository != req.Inputs.Repository)
	addUpdateDiff(diff, "issues", !equalBoolPtr(req.State.Issues, req.Inputs.Issues))
	addUpdateDiff(diff, "pullRequests", !equalBoolPtr(req.State.PullRequests, req.Inputs.PullRequests))
	addUpdateDiff(diff, "wiki", !equalBoolPtr(req.State.Wiki, req.Inputs.Wiki))
	addUpdateDiff(diff, "projects", !equalBoolPtr(req.State.Projects, req.Inputs.Projects))
	addUpdateDiff(diff, "releases", !equalBoolPtr(req.State.Releases, req.Inputs.Releases))
	addUpdateDiff(diff, "packages", !equalBoolPtr(req.State.Packages, req.Inputs.Packages))
	addUpdateDiff(diff, "actions", !equalBoolPtr(req.State.Actions, req.Inputs.Actions))
	addUpdateDiff(diff, "externalWikiUrl", !equalStringPtr(req.State.ExternalWikiURL, req.Inputs.ExternalWikiURL))
	addUpdateDiff(diff, "globallyEditableWiki", !equalBoolPtr(req.State.GloballyEditableWiki, req.Inputs.GloballyEditableWiki))
	addUpdateDiff(diff, "wikiBranch", !equalStringPtr(req.State.WikiBranch, req.Inputs.WikiBranch))
	addUpdateDiff(diff, "internalTrackerEnableTimeTracker", !equalBoolPtr(req.State.InternalTrackerEnableTimeTracker, req.Inputs.InternalTrackerEnableTimeTracker))
	addUpdateDiff(diff, "internalTrackerAllowOnlyContributorsToTrackTime", !equalBoolPtr(req.State.InternalTrackerAllowOnlyContributorsToTrackTime, req.Inputs.InternalTrackerAllowOnlyContributorsToTrackTime))
	addUpdateDiff(diff, "internalTrackerEnableIssueDependencies", !equalBoolPtr(req.State.InternalTrackerEnableIssueDependencies, req.Inputs.InternalTrackerEnableIssueDependencies))
	addUpdateDiff(diff, "externalTrackerUrl", !equalStringPtr(req.State.ExternalTrackerURL, req.Inputs.ExternalTrackerURL))
	addUpdateDiff(diff, "externalTrackerFormat", !equalStringPtr(req.State.ExternalTrackerFormat, req.Inputs.ExternalTrackerFormat))
	addUpdateDiff(diff, "externalTrackerStyle", !equalStringPtr(req.State.ExternalTrackerStyle, req.Inputs.ExternalTrackerStyle))
	addUpdateDiff(diff, "externalTrackerRegexPattern", !equalStringPtr(req.State.ExternalTrackerRegexPattern, req.Inputs.ExternalTrackerRegexPattern))
	return p.DiffResponse{HasChanges: len(diff) > 0, DeleteBeforeReplace: true, DetailedDiff: diff}, nil
}

func (RepositorySettings) Update(ctx context.Context, req infer.UpdateRequest[RepositorySettingsArgs, RepositorySettingsState]) (infer.UpdateResponse[RepositorySettingsState], error) {
	if req.DryRun {
		return infer.UpdateResponse[RepositorySettingsState]{Output: repositorySettingsStateFromArgs(req.Inputs)}, nil
	}

	client, err := clientFromConfig(ctx)
	if err != nil {
		return infer.UpdateResponse[RepositorySettingsState]{}, err
	}
	repo, _, err := client.GetRepo(req.State.Owner, req.State.Repository)
	if err != nil {
		return infer.UpdateResponse[RepositorySettingsState]{}, err
	}
	updated, _, err := client.EditRepo(req.State.Owner, req.State.Repository, repositorySettingsEditOption(req.Inputs, repo))
	if err != nil {
		return infer.UpdateResponse[RepositorySettingsState]{}, err
	}

	return infer.UpdateResponse[RepositorySettingsState]{Output: repositorySettingsStateFromAPI(req.Inputs, updated)}, nil
}

func (RepositorySettings) Delete(_ context.Context, _ infer.DeleteRequest[RepositorySettingsState]) (infer.DeleteResponse, error) {
	return infer.DeleteResponse{}, nil
}

func repositorySettingsEditOption(args RepositorySettingsArgs, repo *forgejo.Repository) forgejo.EditRepoOption {
	return repositorySettingsConfigEditOption(args.RepositorySettingsConfig, repo)
}

func repositorySettingsConfigEditOption(settings RepositorySettingsConfig, repo *forgejo.Repository) forgejo.EditRepoOption {
	opt := forgejo.EditRepoOption{
		HasIssues:            settings.Issues,
		HasPullRequests:      settings.PullRequests,
		HasWiki:              settings.Wiki,
		HasProjects:          settings.Projects,
		HasReleases:          settings.Releases,
		HasPackages:          settings.Packages,
		HasActions:           settings.Actions,
		GloballyEditableWiki: settings.GloballyEditableWiki,
		WikiBranch:           settings.WikiBranch,
	}
	if settings.ExternalWikiURL != nil {
		opt.ExternalWiki = &forgejo.ExternalWiki{ExternalWikiURL: *settings.ExternalWikiURL}
		if opt.HasWiki == nil {
			opt.HasWiki = boolPtr(true)
		}
	}
	if hasInternalTrackerSettings(settings) {
		tracker := forgejo.InternalTracker{}
		if repo != nil && repo.InternalTracker != nil {
			tracker = *repo.InternalTracker
		}
		if settings.InternalTrackerEnableTimeTracker != nil {
			tracker.EnableTimeTracker = *settings.InternalTrackerEnableTimeTracker
		}
		if settings.InternalTrackerAllowOnlyContributorsToTrackTime != nil {
			tracker.AllowOnlyContributorsToTrackTime = *settings.InternalTrackerAllowOnlyContributorsToTrackTime
		}
		if settings.InternalTrackerEnableIssueDependencies != nil {
			tracker.EnableIssueDependencies = *settings.InternalTrackerEnableIssueDependencies
		}
		opt.InternalTracker = &tracker
		if opt.HasIssues == nil {
			opt.HasIssues = boolPtr(true)
		}
	}
	if hasExternalTrackerSettings(settings) {
		tracker := forgejo.ExternalTracker{}
		if repo != nil && repo.ExternalTracker != nil {
			tracker = *repo.ExternalTracker
		}
		if settings.ExternalTrackerURL != nil {
			tracker.ExternalTrackerURL = *settings.ExternalTrackerURL
		}
		if settings.ExternalTrackerFormat != nil {
			tracker.ExternalTrackerFormat = *settings.ExternalTrackerFormat
		}
		if settings.ExternalTrackerStyle != nil {
			tracker.ExternalTrackerStyle = *settings.ExternalTrackerStyle
		}
		if settings.ExternalTrackerRegexPattern != nil {
			tracker.ExternalTrackerRegexPattern = *settings.ExternalTrackerRegexPattern
		}
		opt.ExternalTracker = &tracker
		if opt.HasIssues == nil {
			opt.HasIssues = boolPtr(true)
		}
	}
	return opt
}

func repositorySettingsStateFromAPI(template RepositorySettingsArgs, repo *forgejo.Repository) RepositorySettingsState {
	args := template
	if repo == nil {
		return repositorySettingsStateFromArgs(args)
	}
	repoState := repositoryStateFromAPI(repo)
	args.Owner = repoState.Owner
	args.Repository = repoState.Name
	args.RepositorySettingsConfig = repositorySettingsConfigFromAPI(template.RepositorySettingsConfig, repo)
	return repositorySettingsStateFromArgs(args)
}

func repositorySettingsConfigFromAPI(template RepositorySettingsConfig, repo *forgejo.Repository) RepositorySettingsConfig {
	settings := template
	if repo == nil {
		return settings
	}
	if template.Issues != nil {
		settings.Issues = boolPtr(repo.HasIssues)
	}
	if template.PullRequests != nil {
		settings.PullRequests = boolPtr(repo.HasPullRequests)
	}
	if template.Wiki != nil {
		settings.Wiki = boolPtr(repo.HasWiki)
	}
	if template.Projects != nil {
		settings.Projects = boolPtr(repo.HasProjects)
	}
	if template.Releases != nil {
		settings.Releases = boolPtr(repo.HasReleases)
	}
	if template.Packages != nil {
		settings.Packages = boolPtr(repo.HasPackages)
	}
	if template.Actions != nil {
		settings.Actions = boolPtr(repo.HasActions)
	}
	if template.ExternalWikiURL != nil {
		value := ""
		if repo.ExternalWiki != nil {
			value = repo.ExternalWiki.ExternalWikiURL
		}
		settings.ExternalWikiURL = &value
	}
	if template.InternalTrackerEnableTimeTracker != nil {
		settings.InternalTrackerEnableTimeTracker = boolPtr(repo.InternalTracker != nil && repo.InternalTracker.EnableTimeTracker)
	}
	if template.InternalTrackerAllowOnlyContributorsToTrackTime != nil {
		settings.InternalTrackerAllowOnlyContributorsToTrackTime = boolPtr(repo.InternalTracker != nil && repo.InternalTracker.AllowOnlyContributorsToTrackTime)
	}
	if template.InternalTrackerEnableIssueDependencies != nil {
		settings.InternalTrackerEnableIssueDependencies = boolPtr(repo.InternalTracker != nil && repo.InternalTracker.EnableIssueDependencies)
	}
	if template.ExternalTrackerURL != nil {
		settings.ExternalTrackerURL = stringPtr(repositoryExternalTrackerValue(repo, func(t *forgejo.ExternalTracker) string { return t.ExternalTrackerURL }))
	}
	if template.ExternalTrackerFormat != nil {
		settings.ExternalTrackerFormat = stringPtr(repositoryExternalTrackerValue(repo, func(t *forgejo.ExternalTracker) string { return t.ExternalTrackerFormat }))
	}
	if template.ExternalTrackerStyle != nil {
		settings.ExternalTrackerStyle = stringPtr(repositoryExternalTrackerValue(repo, func(t *forgejo.ExternalTracker) string { return t.ExternalTrackerStyle }))
	}
	if template.ExternalTrackerRegexPattern != nil {
		settings.ExternalTrackerRegexPattern = stringPtr(repositoryExternalTrackerValue(repo, func(t *forgejo.ExternalTracker) string { return t.ExternalTrackerRegexPattern }))
	}
	return settings
}

func repositorySettingsReadArgs(owner, repo string, req infer.ReadRequest[RepositorySettingsArgs, RepositorySettingsState]) RepositorySettingsArgs {
	args := req.Inputs
	if args.Owner == "" {
		args.Owner = owner
	}
	if args.Repository == "" {
		args.Repository = repo
	}
	if args.Issues == nil {
		args.Issues = req.State.Issues
	}
	if args.PullRequests == nil {
		args.PullRequests = req.State.PullRequests
	}
	if args.Wiki == nil {
		args.Wiki = req.State.Wiki
	}
	if args.Projects == nil {
		args.Projects = req.State.Projects
	}
	if args.Releases == nil {
		args.Releases = req.State.Releases
	}
	if args.Packages == nil {
		args.Packages = req.State.Packages
	}
	if args.Actions == nil {
		args.Actions = req.State.Actions
	}
	if args.ExternalWikiURL == nil {
		args.ExternalWikiURL = req.State.ExternalWikiURL
	}
	if args.GloballyEditableWiki == nil {
		args.GloballyEditableWiki = req.State.GloballyEditableWiki
	}
	if args.WikiBranch == nil {
		args.WikiBranch = req.State.WikiBranch
	}
	if args.InternalTrackerEnableTimeTracker == nil {
		args.InternalTrackerEnableTimeTracker = req.State.InternalTrackerEnableTimeTracker
	}
	if args.InternalTrackerAllowOnlyContributorsToTrackTime == nil {
		args.InternalTrackerAllowOnlyContributorsToTrackTime = req.State.InternalTrackerAllowOnlyContributorsToTrackTime
	}
	if args.InternalTrackerEnableIssueDependencies == nil {
		args.InternalTrackerEnableIssueDependencies = req.State.InternalTrackerEnableIssueDependencies
	}
	if args.ExternalTrackerURL == nil {
		args.ExternalTrackerURL = req.State.ExternalTrackerURL
	}
	if args.ExternalTrackerFormat == nil {
		args.ExternalTrackerFormat = req.State.ExternalTrackerFormat
	}
	if args.ExternalTrackerStyle == nil {
		args.ExternalTrackerStyle = req.State.ExternalTrackerStyle
	}
	if args.ExternalTrackerRegexPattern == nil {
		args.ExternalTrackerRegexPattern = req.State.ExternalTrackerRegexPattern
	}
	return args
}

func repositorySettingsStateFromArgs(args RepositorySettingsArgs) RepositorySettingsState {
	return RepositorySettingsState{RepositorySettingsArgs: args}
}

func repositorySettingsID(owner, repo string) string {
	return fmt.Sprintf("%s/%s", owner, repo)
}

func repositorySettingsParts(id string, state RepositorySettingsState) (string, string, error) {
	owner, repo := state.Owner, state.Repository
	if owner == "" || repo == "" {
		parts := strings.SplitN(id, "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return "", "", fmt.Errorf("repository settings ID must have the form owner/repository, got %q", id)
		}
		if owner == "" {
			owner = parts[0]
		}
		if repo == "" {
			repo = parts[1]
		}
	}
	return owner, repo, nil
}

func hasInternalTrackerSettings(args RepositorySettingsConfig) bool {
	return args.InternalTrackerEnableTimeTracker != nil || args.InternalTrackerAllowOnlyContributorsToTrackTime != nil || args.InternalTrackerEnableIssueDependencies != nil
}

func hasExternalTrackerSettings(args RepositorySettingsConfig) bool {
	return args.ExternalTrackerURL != nil || args.ExternalTrackerFormat != nil || args.ExternalTrackerStyle != nil || args.ExternalTrackerRegexPattern != nil
}

func repositoryExternalTrackerValue(repo *forgejo.Repository, value func(*forgejo.ExternalTracker) string) string {
	if repo == nil || repo.ExternalTracker == nil {
		return ""
	}
	return value(repo.ExternalTracker)
}

func boolPtr(value bool) *bool {
	return &value
}

func stringPtr(value string) *string {
	return &value
}

func equalBoolPtr(a, b *bool) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func equalStringPtr(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func repositorySettingsTemplate(inputs, state *RepositorySettingsConfig) *RepositorySettingsConfig {
	if inputs != nil {
		return inputs
	}
	return state
}

func equalRepositorySettingsConfigPtr(a, b *RepositorySettingsConfig) bool {
	if a == nil || b == nil {
		return a == b
	}
	return equalBoolPtr(a.Issues, b.Issues) &&
		equalBoolPtr(a.PullRequests, b.PullRequests) &&
		equalBoolPtr(a.Wiki, b.Wiki) &&
		equalBoolPtr(a.Projects, b.Projects) &&
		equalBoolPtr(a.Releases, b.Releases) &&
		equalBoolPtr(a.Packages, b.Packages) &&
		equalBoolPtr(a.Actions, b.Actions) &&
		equalStringPtr(a.ExternalWikiURL, b.ExternalWikiURL) &&
		equalBoolPtr(a.GloballyEditableWiki, b.GloballyEditableWiki) &&
		equalStringPtr(a.WikiBranch, b.WikiBranch) &&
		equalBoolPtr(a.InternalTrackerEnableTimeTracker, b.InternalTrackerEnableTimeTracker) &&
		equalBoolPtr(a.InternalTrackerAllowOnlyContributorsToTrackTime, b.InternalTrackerAllowOnlyContributorsToTrackTime) &&
		equalBoolPtr(a.InternalTrackerEnableIssueDependencies, b.InternalTrackerEnableIssueDependencies) &&
		equalStringPtr(a.ExternalTrackerURL, b.ExternalTrackerURL) &&
		equalStringPtr(a.ExternalTrackerFormat, b.ExternalTrackerFormat) &&
		equalStringPtr(a.ExternalTrackerStyle, b.ExternalTrackerStyle) &&
		equalStringPtr(a.ExternalTrackerRegexPattern, b.ExternalTrackerRegexPattern)
}

func mergeRepositoryEditOption(base *forgejo.EditRepoOption, settings forgejo.EditRepoOption) {
	if settings.HasIssues != nil {
		base.HasIssues = settings.HasIssues
	}
	if settings.HasPullRequests != nil {
		base.HasPullRequests = settings.HasPullRequests
	}
	if settings.HasWiki != nil {
		base.HasWiki = settings.HasWiki
	}
	if settings.HasProjects != nil {
		base.HasProjects = settings.HasProjects
	}
	if settings.HasReleases != nil {
		base.HasReleases = settings.HasReleases
	}
	if settings.HasPackages != nil {
		base.HasPackages = settings.HasPackages
	}
	if settings.HasActions != nil {
		base.HasActions = settings.HasActions
	}
	if settings.ExternalWiki != nil {
		base.ExternalWiki = settings.ExternalWiki
	}
	if settings.GloballyEditableWiki != nil {
		base.GloballyEditableWiki = settings.GloballyEditableWiki
	}
	if settings.WikiBranch != nil {
		base.WikiBranch = settings.WikiBranch
	}
	if settings.InternalTracker != nil {
		base.InternalTracker = settings.InternalTracker
	}
	if settings.ExternalTracker != nil {
		base.ExternalTracker = settings.ExternalTracker
	}
}
