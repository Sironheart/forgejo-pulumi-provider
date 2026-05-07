# Forgejo Pulumi Provider

A native Pulumi provider for managing Forgejo resources through the Forgejo REST API.

This provider is released only through this repository's Forgejo instance. Provider plugins are attached to Forgejo Releases and generated SDKs are published to the same instance's package registries.

## Supported Languages

The provider schema generates SDKs for the published Pulumi languages:

| Language | SDK path | Package or namespace |
| --- | --- | --- |
| TypeScript / JavaScript | `sdk/nodejs` | `@sironheart/pulumi-forgejo` |
| Go | `sdk/go` | `forgejo.siron.casa/sironheart/forgejo-pulumi-provider/sdk/go` |
| .NET | `sdk/dotnet` | `Pulumi.Forgejo` |
| Java | `sdk/java` | `casa.siron.pulumi.forgejo` |
| YAML | no SDK required | uses the provider schema directly |

## Provider Resources

`forgejo:index:Repository` manages a Forgejo repository.

Inputs: `name`, `owner`, `description`, `private`, `defaultBranch`, `website`, `issues`, `wiki`, `projects`, `template`.

Outputs: `fullName`, `htmlUrl`, `sshUrl`, `cloneUrl`.

`forgejo:index:Organization` manages a Forgejo organization.

Inputs: `name`, `fullName`, `description`, `website`, `location`, `visibility`.

Outputs: `avatarUrl`.

`forgejo:index:OrganizationTeam` manages a Forgejo organization team.

Inputs: `organization`, `name`, `description`, `permission`, `canCreateOrgRepo`, `includesAllRepositories`, `unitsMap`.

Outputs: `teamId`.

`forgejo:index:DeployKey` manages a repository deploy key.

Inputs: `owner`, `repository`, `title`, `key`, `readOnly`.

Outputs: `keyId`, `url`, `fingerprint`.

`forgejo:index:PublicKey` manages an SSH public key for the authenticated user.

Inputs: `title`, `key`, `readOnly`.

Outputs: `keyId`, `url`, `fingerprint`, `keyType`, `owner`.

`forgejo:index:RepositoryActionVariable` manages a Forgejo Actions variable for a repository.

Inputs: `owner`, `repository`, `name`, `value`.

Outputs: `ownerId`, `repoId`.

`forgejo:index:RepositoryTagProtection` manages a repository tag protection rule.

Inputs: `owner`, `repository`, `namePattern`, `whitelistUsernames`, `whitelistTeams`.

Outputs: `protectionId`.

`forgejo:index:RepositoryPushMirror` manages a repository push mirror and can limit it to matching branches.

Inputs: `owner`, `repository`, `remoteAddress`, `remoteUsername`, `remotePassword`, `interval`, `branchFilter`, `syncOnCommit`, `useSsh`.

Outputs: `remoteName`, `publicKey`, `created`, `lastUpdate`, `lastError`.

`forgejo:index:getCurrentUser` returns information about the authenticated Forgejo user.

Outputs: `userId`, `login`, `fullName`, `email`, `isAdmin`.

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
npm install @sironheart/pulumi-forgejo
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
    implementation "casa.siron.pulumi:forgejo:0.1.0"
}
```

## TypeScript Example

```ts
import * as pulumi from "@pulumi/pulumi";
import * as forgejo from "@sironheart/pulumi-forgejo";

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
import casa.siron.pulumi.forgejo.Repository;
import casa.siron.pulumi.forgejo.RepositoryArgs;

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
| `mise run check` | Run formatting, linting, tests, schema export, and SDK generation. |

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
