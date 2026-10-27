# Forgejo Pulumi Provider

A native Pulumi provider for managing Forgejo resources through the Forgejo REST API.

Provider plugins are attached to this repository's Forgejo Releases and generated SDKs are published to the same instance's package registries. The GitHub repository also publishes SDKs to GitHub Packages and attaches the schema and SDK sources to GitHub Releases.

For GitHub installation instructions, see [Using GitHub releases](docs/github-releases.md).

## Supported Languages

The provider schema generates SDKs for the published Pulumi languages:

| Language | SDK path | Package or namespace |
| --- | --- | --- |
| TypeScript / JavaScript | `sdk/nodejs` | `@sironheart/pulumi-forgejo-provider` |
| Go | `sdk/go` | `forgejo.siron.casa/sironheart/forgejo-pulumi-provider/sdk/go` |
| .NET | `sdk/dotnet` | `Pulumi.Forgejo` |
| Java | `sdk/java` | `casa.siron.forgejo.forgejo` |
| YAML | no SDK required | uses the provider schema directly |

## Provider Resources

`forgejo:index:Repository` manages a Forgejo repository.

Inputs: `name`, `owner`, `description`, `private`, `defaultBranch`, `website`, `template`, `topics`, `settings`.

Creation-only inputs: `initializeWithReadme`, `gitignores`, `license`, `readme`, `issueLabels`, `trustModel`, `objectFormatName`. `readme` selects a template, not literal contents; generating README, gitignore, or license files requires `initializeWithReadme`. Changing these inputs on an existing repository has no effect.

Outputs: `fullName`, `htmlUrl`, `sshUrl`, `cloneUrl`.

`forgejo:index:Organization` manages a Forgejo organization.

Inputs: `name`, `fullName`, `description`, `email`, `website`, `location`, `visibility`, `repoAdminChangeTeamAccess`.

Outputs: `avatarUrl`.

`forgejo:index:OrganizationTeam` manages a Forgejo organization team.

Inputs: `organization`, `name`, `description`, `permission`, `canCreateOrgRepo`, `includesAllRepositories`, `unitsMap`.

Outputs: `teamId`.

`forgejo:index:OrganizationTeamMember` manages a Forgejo organization team membership.

Inputs: `organization`, `team`, `teamId`, `username`.

Outputs: `teamId`, `userId`.

`forgejo:index:DeployKey` manages a repository deploy key.

Inputs: `owner`, `repository`, `title`, `key`, `readOnly`.

Outputs: `keyId`, `url`, `fingerprint`.

`forgejo:index:PublicKey` manages an SSH public key for the authenticated user.

Inputs: `title`, `key`, `readOnly`.

Outputs: `keyId`, `url`, `fingerprint`, `keyType`, `owner`.

`forgejo:index:RepositoryActionVariable` manages a Forgejo Actions variable for a repository.

Inputs: `owner`, `repository`, `name`, `value`.

Outputs: `ownerId`, `repoId`.

`forgejo:index:RepositoryActionSecret` manages a Forgejo Actions secret for a repository.

Inputs: `owner`, `repository`, `name`, `value`.

Outputs: `created`.

`forgejo:index:OrganizationActionVariable` manages a Forgejo Actions variable for an organization.

Inputs: `organization`, `name`, `value`.

Outputs: `ownerId`.

`forgejo:index:OrganizationActionSecret` manages a Forgejo Actions secret for an organization.

Inputs: `organization`, `name`, `value`.

Outputs: `created`.

`forgejo:index:RepositoryTagProtection` manages a repository tag protection rule.

Inputs: `owner`, `repository`, `namePattern`, `whitelistUsernames`, `whitelistTeams`.

Outputs: `protectionId`.

`forgejo:index:RepositoryBranchProtection` manages a repository branch protection rule.

Inputs include `owner`, `repository`, `name`, push and merge whitelists, status-check settings, review approval settings, `applyToAdmins`, `ignoreStaleApprovals`, signed-commit requirements, and protected/unprotected file patterns.

Outputs: `created`, `updated`.

`forgejo:index:RepositorySettings` manages repository units, merge behavior, archiving, wiki and issue trackers, and pull-mirror settings. The same settings are available through `Repository.settings`; leave optional fields unset to avoid managing them.

Inputs include `owner`, `repository`, `issues`, `pullRequests`, `defaultDeleteBranchAfterMerge`, `wiki`, `projects`, `releases`, `packages`, `actions`, `externalWikiUrl`, internal tracker settings, and external tracker settings.

Merge settings include `allowMergeCommits`, `allowRebase`, `allowRebaseMerge`, `allowSquashMerge`, `allowFastForwardOnlyMerge`, `defaultMergeStyle`, `ignoreWhitespaceConflicts`, `allowManualMerge`, `autodetectManualMerge`, `allowRebaseUpdate`, `defaultUpdateStyle`, and `defaultAllowMaintainerEdit`. `mirrorInterval` and `enablePrune` apply to existing **pull mirrors**.

`forgejo:index:RepositoryPushMirror` manages a repository push mirror and can limit it to matching branches.

Inputs: `owner`, `repository`, `remoteAddress`, `remoteUsername`, `remotePassword`, `interval`, `branchFilter`, `syncOnCommit`, `useSsh`.

Outputs: `remoteName`, `publicKey`, `created`, `lastUpdate`, `lastError`.

`interval` is required; use `0s` to disable periodic synchronization. Equivalent durations such as `1h` and `1h0m0s` do not trigger replacement. Other mirror changes require replacement; SSH mirror replacement generates a new key.

`forgejo:index:getCurrentUser` returns information about the authenticated Forgejo user.

Outputs: `userId`, `login`, `fullName`, `email`, `isAdmin`.

### API Compatibility

Validated against Forgejo **17.0.1**. See [API compatibility findings](docs/api-compatibility.md) for upstream sources, supported values, and version limitations. Forgejo does not return `allowManualMerge`, `autodetectManualMerge`, or `enablePrune`, so refresh preserves their configured values rather than detecting external changes.

## Configuration

The provider requires a Forgejo base URL and an API token:

```sh
pulumi config set forgejo:url https://forgejo.example
pulumi config set --secret forgejo:token <token>
```

The same values can be supplied through environment variables:

```sh
export FORGEJO_URL=https://forgejo.example
export FORGEJO_TOKEN=<token>
```

`forgejo:token` is marked as a Pulumi secret.

## Installing The Plugin

Install a released provider plugin from this Forgejo repository:

```sh
pulumi plugin install resource forgejo 0.1.0 --server https://forgejo.siron.casa/sironheart/forgejo-pulumi-provider/releases/download/v0.1.0/
```

Release archives follow Pulumi's plugin naming convention:

```text
pulumi-resource-forgejo-v<version>-<os>-<arch>.tar.gz
```

## Installing SDK Packages

SDK packages are published to this Forgejo instance's local package registries when a release is created.

TypeScript / JavaScript:

```sh
npm config set @sironheart:registry https://forgejo.siron.casa/api/packages/sironheart/npm/
npm install @sironheart/pulumi-forgejo-provider
```

Go:

```sh
GOPROXY=https://forgejo.siron.casa/api/packages/sironheart/go go get forgejo.siron.casa/sironheart/forgejo-pulumi-provider/sdk/go@v0.1.0
```

.NET:

```sh
dotnet nuget add source --name forgejo https://forgejo.siron.casa/api/packages/sironheart/nuget/index.json
dotnet add package Pulumi.Forgejo --source forgejo --version 0.1.0
```

Java:

```groovy
repositories {
    maven { url "https://forgejo.siron.casa/api/packages/sironheart/maven" }
}

dependencies {
    implementation "casa.siron.forgejo:forgejo:0.1.0"
}
```

## TypeScript Example

```ts
import * as pulumi from "@pulumi/pulumi";
import * as forgejo from "@sironheart/pulumi-forgejo-provider";

const repo = new forgejo.Repository("example", {
    name: "example",
    description: "Managed by Pulumi",
    private: true,
});

export const cloneUrl = repo.cloneUrl;
```

## Go Example

```go
package main

import (
    forgejo "forgejo.siron.casa/sironheart/forgejo-pulumi-provider/sdk/go"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        repo, err := forgejo.NewRepository(ctx, "example", &forgejo.RepositoryArgs{
            Name:        pulumi.String("example"),
            Description: pulumi.StringPtr("Managed by Pulumi"),
            Private:     pulumi.BoolPtr(true),
        })
        if err != nil {
            return err
        }

        ctx.Export("cloneUrl", repo.CloneUrl)
        return nil
    })
}
```

## .NET Example

```csharp
using System.Collections.Generic;
using Pulumi;
using Pulumi.Forgejo;

return await Deployment.RunAsync(() =>
{
    var repo = new Repository("example", new RepositoryArgs
    {
        Name = "example",
        Description = "Managed by Pulumi",
        Private = true,
    });

    return new Dictionary<string, object?>
    {
        ["cloneUrl"] = repo.CloneUrl,
    };
});
```

## Java Example

```java
package myproject;

import com.pulumi.Pulumi;
import casa.siron.forgejo.forgejo.Repository;
import casa.siron.forgejo.forgejo.RepositoryArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var repo = new Repository("example", RepositoryArgs.builder()
                .name("example")
                .description("Managed by Pulumi")
                .private_(true)
                .build());

            ctx.export("cloneUrl", repo.cloneUrl());
        });
    }
}
```

## YAML Example

```yaml
name: forgejo-example
runtime: yaml

resources:
  repo:
    type: forgejo:index:Repository
    properties:
      name: example
      description: Managed by Pulumi
      private: true

outputs:
  cloneUrl: ${repo.cloneUrl}
```

## Development

Local development tools are managed with `mise-en-place`:

```sh
mise install
mise run check
```

Useful tasks:

| Task | Purpose |
| --- | --- |
| `mise run fmt` | Format Go code with `golangci-lint fmt` using `.golangci.yml`. |
| `mise run lint` | Run `golangci-lint`. |
| `mise run test` | Run Go tests. |
| `mise run version` | Calculate the next semantic version with `svu`. |
| `mise run build` | Build `bin/pulumi-resource-forgejo`. |
| `mise run schema` | Export `schema.json`. |
| `mise run sdk` | Generate all standard Pulumi SDKs. |
| `mise run release-check` | Validate the GoReleaser configuration. |
| `mise run check` | Run formatting, linting, tests, and the provider build. |

Build release artifacts locally:

```sh
mise exec -- goreleaser release --snapshot --clean --skip=publish
```

Forgejo Actions run validation on pushes and pull requests. Workflows install Go, Pulumi, GoReleaser, and `svu` directly from explicit versions or actions, and run `golangci-lint` through `https://github.com/golangci/golangci-lint-action`. `mise` is only used for local tasks. CI is split into version/comment, Go validation, and SDK validation jobs with Go, Pulumi, and linter caches where applicable.

## Release Automation

Merging into `main` triggers `.forgejo/workflows/release.yml`. The workflow calculates the next semantic version with `svu next`, skips the release when the calculated version already exists, creates the corresponding `vX.Y.Z` Git tag, and publishes the release. The workflow can also be run manually with an optional version override, with or without the leading `v`.

The release workflow:

- Checks out the merge commit on `main`.
- Calculates the next semantic version with `svu`.
- Runs Go linting through `https://github.com/golangci/golangci-lint-action`.
- Generates the schema and SDKs for that version.
- Creates and pushes the release Git tag after validation and SDK generation succeed.
- Uses the GoReleaser CLI to build provider plugin archives, generate the changelog, and create the Forgejo Release.
- Publishes the SDK packages for Go, npm, NuGet, and Maven after the Forgejo Release is created.

The workflow expects `FORGEJO_TOKEN` to be configured as a repository secret. The token must be allowed to create releases/tags and publish packages for the `sironheart` owner.

### GitHub Releases

To install a released version, follow [Using GitHub releases](docs/github-releases.md). The following section describes the publishing workflow.

`.github/workflows/release.yml` runs only on github.com when a `vX.Y.Z` tag (optionally with a prerelease suffix) is pushed by the Forgejo push mirror. It does not run for branch pushes or on Forgejo, and never creates commits or changes mirrored tags.

The workflow builds the provider at the tag, generates the schema and the existing four SDKs using `scripts/gen-sdk.sh`, then publishes:

| Destination | Artifacts |
| --- | --- |
| GitHub Release | `schema.json` and `forgejo-sdk-<language>-<tag>.tar.gz` source archives for Node.js, Go, .NET and Java |
| GitHub Packages (npm) | `@sironheart/pulumi-forgejo-provider` |
| GitHub Packages (NuGet) | `Pulumi.Forgejo` |
| GitHub Packages (Maven) | `casa.siron.forgejo:forgejo` |

Go is distributed as a source archive only; GitHub Packages has no Go module registry. Plugin binaries and their SDK download URLs remain hosted on Forgejo. Attaching the schema does not by itself register the provider in the Pulumi Registry.

Setup:

- Enable Actions in the GitHub repository. Publishing uses the automatic `GITHUB_TOKEN` with `contents: write` for releases and `packages: write` for packages; no additional publishing secrets are needed.
- Configure the Forgejo push mirror to include tags and use credentials allowed to push GitHub workflow files. The npm scope in `internal/provider/provider.go` must match the GitHub repository owner (currently `sironheart`).
- Ensure this workflow is included in the tagged commit. Existing tags are not rebuilt automatically, and [GitHub does not emit tag push events when more than three tags are pushed together](https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows#push).
- Check package visibility after first publication. GitHub npm and NuGet packages default to private, and [installation requires authentication even for public packages](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-npm-registry).

Prereleases become GitHub prereleases and use npm's `next` dist-tag instead of `latest`. Package publishing runs independently per language, so use **Re-run failed jobs** after a partial failure; published npm/Maven versions cannot be overwritten. Release assets can be uploaded again.
