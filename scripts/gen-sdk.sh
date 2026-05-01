#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(git rev-parse --show-toplevel)
cd "$ROOT_DIR"

VERSION=${FORGEJO_PROVIDER_VERSION:-$(svu next)}
VERSION=${VERSION#v}
printf 'Generating SDKs with version %s\n' "$VERSION"

rm -rf sdk/python
languages=${FORGEJO_SDK_LANGUAGES:-"nodejs go dotnet java"}
for language in $languages; do
  pulumi package gen-sdk ./bin/pulumi-resource-forgejo --language "$language" --version "$VERSION"
done

case " $languages " in
  *" nodejs "*) node <<'EOF'
const fs = require("fs")
const packagePath = "sdk/nodejs/package.json";
const pkg = JSON.parse(fs.readFileSync(packagePath, "utf8"));
pkg.repository = {
  type: "git",
  url: "git+https://forgejo.siron.casa/sironheart/forgejo-pulumi-provider.git",
};
pkg.bugs = {
  url: "https://forgejo.siron.casa/sironheart/forgejo-pulumi-provider/issues",
};
pkg.publishConfig = {
  "@sironheart:registry": "https://forgejo.siron.casa/api/packages/sironheart/npm/",
  access: "public",
};
pkg.main = "bin/index.js";
pkg.types = "bin/index.d.ts";
pkg.files = [
  "bin",
  "package.json",
  "README.md",
];
fs.writeFileSync(packagePath, `${JSON.stringify(pkg, null, 4)}\n`);
EOF
  ;;
esac

case " $languages " in
  *" go "*)
    if [ ! -f sdk/go/go.mod ] || ! grep -qx 'module forgejo.siron.casa/sironheart/forgejo-pulumi-provider/sdk/go' sdk/go/go.mod; then
      rm -f sdk/go/go.mod sdk/go/go.sum
      go -C sdk/go mod init forgejo.siron.casa/sironheart/forgejo-pulumi-provider/sdk/go
    fi
    go -C sdk/go get github.com/pulumi/pulumi/sdk/v3@v3.232.0
    go -C sdk/go mod tidy
  ;;
esac

# pulumi-java-gen writes the plugin URL into a Gradle GString; keep Pulumi's
# ${VERSION} placeholder literal instead of letting Gradle resolve it.
case " $languages " in
  *" java "*) perl -0pi -e 's/(?<!\\)\$\{VERSION\}/\\\$\{VERSION\}/g' sdk/java/build.gradle ;;
esac

# The .NET SDK project embeds this file so the SDK can report the provider plugin version.
case " $languages " in
  *" dotnet "*) printf '%s\n' "$VERSION" > sdk/dotnet/version.txt ;;
esac
