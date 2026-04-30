#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(git rev-parse --show-toplevel)
cd "$ROOT_DIR"

VERSION=${FORGEJO_PROVIDER_VERSION:-${GITHUB_REF_NAME:-0.0.0-dev}}
VERSION=${VERSION#v}

rm -rf dist
mkdir -p bin dist

go build \
  -trimpath \
  -ldflags "-s -w -X main.version=$VERSION" \
  -o bin/pulumi-resource-forgejo \
  ./cmd/pulumi-resource-forgejo
pulumi package get-schema ./bin/pulumi-resource-forgejo > schema.json
FORGEJO_PROVIDER_VERSION=$VERSION scripts/gen-sdk.sh

targets=(
  "linux amd64"
  "linux arm64"
  "darwin amd64"
  "darwin arm64"
  "windows amd64"
)

for target in "${targets[@]}"; do
  read -r goos goarch <<<"$target"
  tmp=$(mktemp -d)
  binary="pulumi-resource-forgejo"
  if [[ "$goos" == "windows" ]]; then
    binary="pulumi-resource-forgejo.exe"
  fi

  GOOS=$goos GOARCH=$goarch CGO_ENABLED=0 go build \
    -trimpath \
    -ldflags "-s -w -X main.version=$VERSION" \
    -o "$tmp/$binary" \
    ./cmd/pulumi-resource-forgejo

  cp README.md "$tmp/README.md"
  cp LICENSE "$tmp/LICENSE"
  tar -czf "dist/pulumi-resource-forgejo-v${VERSION}-${goos}-${goarch}.tar.gz" -C "$tmp" .
  rm -rf "$tmp"
done

cp schema.json dist/schema.json
for sdk in go nodejs python dotnet java; do
  tar -czf "dist/pulumi-forgejo-sdk-${sdk}-v${VERSION}.tar.gz" -C "sdk/$sdk" .
done
if command -v sha256sum >/dev/null 2>&1; then
  sha256sum dist/*.tar.gz dist/schema.json > dist/checksums.txt
else
  shasum -a 256 dist/*.tar.gz dist/schema.json > dist/checksums.txt
fi
