package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v3"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi-go-provider/middleware/schema"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
)

const defaultTimeout = 30 * time.Second

// Pulumi appends pulumi-resource-forgejo-v<version>-<os>-<arch>.tar.gz to this Forgejo release asset base.
const pluginDownloadURL = "https://forgejo.siron.casa/sironheart/forgejo-pulumi-provider/releases/download/v${VERSION}/"

type Config struct {
	URL   string `pulumi:"url"`
	Token string `pulumi:"token" provider:"secret"`
}

func Provider() p.Provider {
	return infer.Provider(infer.Options{
		Resources: []infer.InferredResource{
			infer.Resource(Repository{}),
			infer.Resource(Organization{}),
			infer.Resource(OrganizationTeam{}),
			infer.Resource(DeployKey{}),
			infer.Resource(PublicKey{}),
			infer.Resource(RepositoryActionVariable{}),
			infer.Resource(RepositoryTagProtection{}),
			infer.Resource(RepositoryPushMirror{}),
		},
		Functions: []infer.InferredFunction{
			infer.Function(GetCurrentUser{}),
		},
		Config: infer.Config(&Config{}),
		Metadata: schema.Metadata{
			DisplayName:       "Forgejo Pulumi Provider",
			Description:       "A Pulumi provider for managing Forgejo resources.",
			Keywords:          []string{"pulumi", "forgejo", "git", "gitea"},
			Homepage:          "https://forgejo.siron.casa/sironheart/forgejo-pulumi-provider",
			Repository:        "https://forgejo.siron.casa/sironheart/forgejo-pulumi-provider",
			Publisher:         "sironheart",
			License:           "Apache-2.0",
			PluginDownloadURL: pluginDownloadURL,
			LanguageMap: map[string]any{
				"nodejs": map[string]any{
					"packageName":          "@sironheart/pulumi-forgejo-provider",
					"packageDescription":   "A Pulumi provider for managing Forgejo resources.",
					"respectSchemaVersion": true,
				},
				"go": map[string]any{
					"importBasePath":  "forgejo.siron.casa/sironheart/forgejo-pulumi-provider/sdk/go",
					"modulePath":      "sdk/go",
					"rootPackageName": "forgejo-pulumi-provider",
				},
				"csharp": map[string]any{
					"rootNamespace": "Pulumi",
				},
				"java": map[string]any{
					"basePackage": "casa.siron.forgejo",
					"buildFiles":  "gradle",
				},
			},
		},
		ModuleMap: map[tokens.ModuleName]tokens.ModuleName{
			"provider": "index",
		},
	})
}

func (c *Config) Annotate(a infer.Annotator) {
	a.Describe(&c.URL, "Base URL of the Forgejo instance. Can also be set with FORGEJO_URL.")
	a.Describe(&c.Token, "Forgejo API token. Can also be set with FORGEJO_TOKEN.")
	a.SetDefault(&c.URL, nil, "FORGEJO_URL")
	a.SetDefault(&c.Token, nil, "FORGEJO_TOKEN")
}

func clientFromConfig(ctx context.Context) (*forgejo.Client, error) {
	cfg := infer.GetConfig[Config](ctx)
	if strings.TrimSpace(cfg.Token) == "" {
		return nil, errors.New("forgejo token is required")
	}
	baseURL, err := normalizeForgejoURL(cfg.URL)
	if err != nil {
		return nil, err
	}

	client, err := forgejo.NewClient(
		baseURL,
		forgejo.SetToken(cfg.Token),
		forgejo.SetContext(ctx),
		forgejo.SetForgejoVersion(""),
		forgejo.SetHTTPClient(&http.Client{Timeout: defaultTimeout}),
	)
	if err != nil {
		return nil, fmt.Errorf("configure Forgejo client: %w", err)
	}
	return client, nil
}

func normalizeForgejoURL(rawURL string) (string, error) {
	trimmed := strings.TrimRight(strings.TrimSpace(rawURL), "/")
	if trimmed == "" {
		return "", errors.New("forgejo url is required")
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("parse forgejo url: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("forgejo url must be absolute: %q", rawURL)
	}

	parsed.Path = strings.TrimRight(parsed.Path, "/")
	if strings.HasSuffix(parsed.Path, "/api/v1") {
		parsed.Path = strings.TrimRight(strings.TrimSuffix(parsed.Path, "/api/v1"), "/")
	}
	return parsed.String(), nil
}

func isNotFound(resp *forgejo.Response) bool {
	return resp != nil && resp.StatusCode == http.StatusNotFound
}
