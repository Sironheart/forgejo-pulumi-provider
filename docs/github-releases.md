# Using GitHub releases

Install the SDK from GitHub Packages, or download its source from
[GitHub Releases](https://github.com/sironheart/forgejo-pulumi-provider/releases).
The [GitHub release workflow](../.github/workflows/release.yml) publishes these packages and assets.
The provider plugin is downloaded from [Forgejo Releases](https://forgejo.siron.casa/sironheart/forgejo-pulumi-provider/releases).

## Choose a version

Examples use release `v2.2.0`. For another release, replace `2.2.0` in the shell
and Gradle examples with your selected version, including any prerelease suffix.
Release tags use `vX.Y.Z`; package versions use `X.Y.Z` without the leading `v`.
Shell examples use Bash from an existing Pulumi project with its language toolchain installed.

```bash
export VERSION=2.2.0
```

| GitHub release asset | Contents |
| --- | --- |
| `schema.json` | Pulumi provider schema |
| `forgejo-sdk-nodejs-v<VERSION>.tar.gz` | SDK sources under `nodejs/` |
| `forgejo-sdk-go-v<VERSION>.tar.gz` | SDK sources under `go/` |
| `forgejo-sdk-dotnet-v<VERSION>.tar.gz` | SDK sources under `dotnet/` |
| `forgejo-sdk-java-v<VERSION>.tar.gz` | SDK sources under `java/` |

These archives contain generated source, not compiled provider plugins or
installable npm tarballs. Use the package registries below for npm, NuGet and Maven.

## Install the matching provider plugin

Install the same version before using an SDK:

```bash
pulumi plugin install resource forgejo "$VERSION" --exact \
  --server "https://forgejo.siron.casa/sironheart/forgejo-pulumi-provider/releases/download/v${VERSION}/"
```

The SDKs also point to this Forgejo download location. GitHub releases do not
contain plugin binaries.

## TypeScript / JavaScript (npm)

Add this registry mapping to your project's `.npmrc`:

```ini
@sironheart:registry=https://npm.pkg.github.com
```

Install the exact version:

```bash
npm install --save-exact "@sironheart/pulumi-forgejo-provider@${VERSION}"
```

Prereleases use the `next` dist-tag; an explicit version also selects them.

## .NET (NuGet)

Add this repository's feed to `packageSources` in your existing `nuget.config`:

```xml
<add key="forgejo-github" value="https://nuget.pkg.github.com/sironheart/index.json" />
```

Install the package:

```bash
dotnet add package Pulumi.Forgejo --version "$VERSION"
```

## Java (Gradle)

Merge these entries into your `build.gradle` (Groovy), replacing `2.2.0` if you
selected another release above:

```groovy
repositories {
    maven {
        url = uri("https://maven.pkg.github.com/sironheart/forgejo-pulumi-provider")
    }
}

dependencies {
    implementation "casa.siron.forgejo:forgejo:2.2.0"
}
```

## Go (release source archive)

The Go SDK is distributed on GitHub as a source archive. Release `v2.2.0` requires Go **1.27.3**
or newer; this example also uses `jq`. Download into a persistent project directory:

```bash
SDK_DIR=".pulumi-sdk/forgejo-v${VERSION}"
mkdir -p "$SDK_DIR"
curl -fL \
  "https://github.com/sironheart/forgejo-pulumi-provider/releases/download/v${VERSION}/forgejo-sdk-go-v${VERSION}.tar.gz" \
  -o "$SDK_DIR/sdk.tar.gz"
tar -xzf "$SDK_DIR/sdk.tar.gz" -C "$SDK_DIR"
jq --arg version "$VERSION" '.version = $version' "$SDK_DIR/go/pulumi-plugin.json" \
  > "$SDK_DIR/go/pulumi-plugin.json.tmp" &&
  mv "$SDK_DIR/go/pulumi-plugin.json.tmp" "$SDK_DIR/go/pulumi-plugin.json"
go mod edit -replace="forgejo.siron.casa/sironheart/forgejo-pulumi-provider/sdk/go=./${SDK_DIR}/go"
go mod edit -require=forgejo.siron.casa/sironheart/forgejo-pulumi-provider/sdk/go@v0.0.0
```

The archive's top-level `go/` directory contains the module. Keep its original
module and import path, even though you downloaded it from GitHub:

```go
import forgejo "forgejo.siron.casa/sironheart/forgejo-pulumi-provider/sdk/go"
```

After using the SDK in your program, run `go mod tidy`. `v0.0.0` is only a local
module placeholder, preserving the module path without a `/v2` suffix.
Setting `pulumi-plugin.json.version` prevents discovery of nonexistent plugin `0.0.0`.

Keep the extracted SDK with your project, or restore the same directory in CI;
the `go.mod` replacement requires it when building.

## Configure Forgejo

Follow [the provider configuration instructions](../README.md#configuration) to
set the Forgejo URL and API token.
