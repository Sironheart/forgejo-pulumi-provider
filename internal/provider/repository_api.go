package provider

import (
	"net/http"
	"net/url"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v3"
)

// These response fields are exposed by Forgejo but missing from the SDK.
type repositoryAPIResponse struct {
	forgejo.Repository
	AllowFastForwardOnly          *bool   `json:"allow_fast_forward_only_merge"`
	AllowRebaseUpdate             *bool   `json:"allow_rebase_update"`
	DefaultDeleteBranchAfterMerge *bool   `json:"default_delete_branch_after_merge"`
	DefaultAllowMaintainerEdit    *bool   `json:"default_allow_maintainer_edit"`
	DefaultUpdateStyle            *string `json:"default_update_style"`
	GloballyEditableWiki          *bool   `json:"globally_editable_wiki"`
	WikiBranch                    *string `json:"wiki_branch"`
	ObjectFormatName              string  `json:"object_format_name"`
}

type repositoryCreateOption struct {
	forgejo.CreateRepoOption
	ObjectFormatName string `json:"object_format_name,omitempty"`
}

func repositoryAPIPath(owner, name string) string {
	return "/repos/" + url.PathEscape(owner) + "/" + url.PathEscape(name)
}

func (c *forgejoAPIClient) getRepository(owner, name string) (*repositoryAPIResponse, *forgejo.Response, error) {
	repo := new(repositoryAPIResponse)
	resp, err := c.do(http.MethodGet, repositoryAPIPath(owner, name), nil, repo)
	return repo, resp, err
}

func (c *forgejoAPIClient) editRepository(owner, name string, opt forgejo.EditRepoOption) (*repositoryAPIResponse, *forgejo.Response, error) {
	repo := new(repositoryAPIResponse)
	resp, err := c.do(http.MethodPatch, repositoryAPIPath(owner, name), opt, repo)
	return repo, resp, err
}
