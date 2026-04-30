# Forgejo Pulumi Provider

A native Pulumi provider for managing Forgejo resources through the Forgejo REST API.

This provider is released only through this repository's Forgejo instance. Provider plugins are attached to Forgejo Releases and generated SDKs are published to the same instance's package registries.

## Supported Languages

The provider schema generates SDKs for all standard Pulumi languages:

| Language | SDK path | Package or namespace |
| --- | --- | --- |
| TypeScript / JavaScript | `sdk/nodejs` | `@sironheart/pulumi-forgejo` |
| Python | `sdk/python` | `pulumi_forgejo` |
| Go | `sdk/go` | `forgejo.siron.casa/sironheart/forgejo-pulumi-provider/sdk/go` |
| .NET | `sdk/dotnet` | `Pulumi.Forgejo` |
| Java | `sdk/java` | `com.sironheart.pulumi.forgejo` |
| YAML | no SDK required | uses the provider schema directly |

## Provider Resources

`forgejo:index:Repository` manages a Forgejo repository.

Inputs: `name`, `owner`, `description`, `private`, `defaultBranch`, `website`, `issues`, `wiki`, `projects`, `template`.

Outputs: `fullName`, `htmlUrl`, `sshUrl`, `cloneUrl`.

`forgejo:index:Organization` manages a Forgejo organization.

Inputs: `name`, `fullName`, `description`, `website`, `location`, `visibility`.

Outputs: `avatarUrl`.

`forgejo:index:DeployKey` manages a repository deploy key.

Inputs: `owner`, `repository`, `title`, `key`, `readOnly`.

Outputs: `keyId`, `url`, `fingerprint`.

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
pulumi plugin install resource forgejo 0.1.0 --server https://forgejo.siron.casa/sironheart/forgejo-pulumi-provider/releases/download/v0.1.0
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

Python:

```sh
python -m pip install --index-url https://forgejo.siron.casa/api/packages/sironheart/pypi/simple pulumi_forgejo
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
    implementation "com.sironheart.pulumi:forgejo:0.1.0"
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

## Python Example

```python
import pulumi
import pulumi_forgejo as forgejo

repo = forgejo.Repository(
    "example",
    name="example",
    description="Managed by Pulumi",
    private=True,
)

pulumi.export("cloneUrl", repo.clone_url)
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
import com.sironheart.pulumi.forgejo.Repository;
import com.sironheart.pulumi.forgejo.RepositoryArgs;

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
| `mise run version` | Calculate the next semantic version with `git-cliff`. |
| `mise run build` | Build `bin/pulumi-resource-forgejo`. |
| `mise run schema` | Export `schema.json`. |
| `mise run sdk` | Generate all standard Pulumi SDKs. |
| `mise run release-check` | Validate the GoReleaser configuration. |
| `mise run check` | Run formatting, linting, tests, schema export, and SDK generation. |

Build release artifacts locally:

```sh
mise exec -- goreleaser release --snapshot --clean --skip=publish
```

Forgejo Actions run validation on pushes and pull requests. The CI workflow installs Go, Pulumi, `golangci-lint`, and `git-cliff` directly from explicit versions or actions in `.forgejo/workflows/ci.yml`; `mise` is only used for local tasks and release automation. CI calculates the provider version with `git-cliff --bumped-version` before building the provider and SDKs.

## Release Automation

Pushing a version tag such as `v0.1.0` triggers `.forgejo/workflows/release.yml`. The workflow can also be run manually for an existing tag by passing the version with or without the leading `v`.

The release workflow:

- Checks out the requested release tag.
- Generates the schema and SDKs for that version.
- Uses GoReleaser to build provider plugin archives, generate the changelog, and create or replace the Forgejo Release.
- Publishes SDK packages to the local Forgejo registries for npm, PyPI, NuGet, Maven, and Go.

The workflow expects `FORGEJO_TOKEN` to be configured as a repository secret. The token must be allowed to create releases/tags and publish packages for the `sironheart` owner.
