#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(git rev-parse --show-toplevel)
cd "$ROOT_DIR"

VERSION=${FORGEJO_PROVIDER_VERSION:-${GORELEASER_CURRENT_TAG:-}}
if [ -z "$VERSION" ]; then
  printf 'FORGEJO_PROVIDER_VERSION or GORELEASER_CURRENT_TAG must be set\n'
  exit 1
fi
VERSION=${VERSION#v}
TAG="v$VERSION"

: "${PACKAGE_USERNAME:?PACKAGE_USERNAME must be set}"
: "${PACKAGE_TOKEN:?PACKAGE_TOKEN must be set}"
: "${REGISTRY_URL:?REGISTRY_URL must be set}"

printf 'Publishing SDKs for %s\n' "$TAG"

module=$(go list -m)
go_archive="pulumi-forgejo-go-${TAG}.zip"
git archive --format=zip --prefix="${module}@${TAG}/" HEAD --output="$go_archive"
curl --fail --silent --show-error \
  --user "${PACKAGE_USERNAME}:${PACKAGE_TOKEN}" \
  --upload-file "$go_archive" \
  "${REGISTRY_URL}/go/upload"

(
  cd sdk/nodejs
  auth_path=${REGISTRY_URL#https:}
  auth_path=${auth_path#http:}
  npm config set @sironheart:registry "${REGISTRY_URL}/npm/"
  npm config set -- "${auth_path}/npm/:_authToken" "$PACKAGE_TOKEN"
  npm install --ignore-scripts
  npm run build
  npm publish --access public --registry "${REGISTRY_URL}/npm/"
)

rm -rf sdk/dotnet/nupkg
dotnet pack sdk/dotnet/Pulumi.Forgejo.csproj --configuration Release --output sdk/dotnet/nupkg -p:Version="$VERSION"
dotnet nuget push sdk/dotnet/nupkg/*.nupkg --source "${REGISTRY_URL}/nuget/index.json" --api-key "$PACKAGE_TOKEN" --skip-duplicate

PACKAGE_VERSION="$VERSION" \
  PUBLISH_REPO_URL="${REGISTRY_URL}/maven" \
  PUBLISH_REPO_USERNAME="$PACKAGE_USERNAME" \
  PUBLISH_REPO_PASSWORD="$PACKAGE_TOKEN" \
  gradle --no-daemon -p sdk/java publish
