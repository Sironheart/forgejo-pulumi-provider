#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(git rev-parse --show-toplevel)
cd "$ROOT_DIR"

VERSION=${FORGEJO_PROVIDER_VERSION:-$(svu next)}
VERSION=${VERSION#v}
printf 'Generating SDKs with version %s\n' "$VERSION"

rm -rf sdk/python
for language in nodejs go dotnet java; do
  pulumi package gen-sdk ./bin/pulumi-resource-forgejo --language "$language" --version "$VERSION"
done

# pulumi-java-gen writes the plugin URL into a Gradle GString; keep Pulumi's
# ${VERSION} placeholder literal instead of letting Gradle resolve it.
perl -0pi -e 's/(?<!\\)\$\{VERSION\}/\\\$\{VERSION\}/g' sdk/java/build.gradle

# The .NET SDK project embeds this file so the SDK can report the provider plugin version.
printf '%s\n' "$VERSION" > sdk/dotnet/version.txt
