# AGENTS.md

## Repo Shape

- This is a Go Pulumi native provider. The plugin entrypoint is `cmd/pulumi-resource-forgejo/main.go`; provider wiring, schema metadata, resources, functions, and env-backed config live in `internal/provider/provider.go`.
- Generated SDKs are committed under `sdk/nodejs`, `sdk/go`, `sdk/dotnet`, and `sdk/java`. Do not edit generated SDK files by hand for provider API changes; update `internal/provider`, regenerate `schema.json`, then regenerate SDKs.
- Resource and function tokens are set with `a.SetToken("index", ...)`; keep new resources/functions registered in `internal/provider/provider.go` so codegen emits them.

## Commands

- Use pinned local tools from `.mise.toml`: `mise install`.
- Full local validation: `mise run check`.
- Fast Go loop: `mise run fmt`, `mise run lint`, `mise run test`.
- Focused test: `go test ./internal/provider -run TestName`.
- Build plugin: `mise run build` creates `bin/pulumi-resource-forgejo`.
- Export schema: `mise run schema` writes `schema.json` from the built plugin.
- Regenerate SDKs: `mise run sdk`; limit languages with `FORGEJO_SDK_LANGUAGES="nodejs go dotnet java"` or a subset.
- Validate release config without publishing: `mise run release-check`; build local release artifacts with `mise exec -- goreleaser release --snapshot --clean --skip=publish`.

## Codegen And Versions

- `scripts/gen-sdk.sh` is the SDK source of truth: it runs `pulumi package gen-sdk`, removes `sdk/python`, fixes Node package metadata, maintains `sdk/go/go.mod`, escapes Java `${VERSION}`, and writes `sdk/dotnet/version.txt`.
- If a change affects provider types, config, descriptions, package metadata, or resources, expect changes in `schema.json` and `sdk/*`; CI fails if generation leaves the tree dirty.
- Build/codegen version defaults to `svu next`. For committed `schema.json` / `sdk/*` changes that CI will check, always regenerate with that default (or `FORGEJO_PROVIDER_VERSION` matching `svu next`, currently without the `v` prefix). Do **not** commit SDK version stamps as `0.0.0-dev` — CI regenerates with `svu next` and fails "Check generated files" on the mismatch.
- Use `FORGEJO_PROVIDER_VERSION=0.0.0-dev` only for throwaway local experiments that you will not commit.
- Current executable package metadata names the Node SDK `@sironheart/pulumi-forgejo-provider`; trust `internal/provider/provider.go`, `schema.json`, and generated package manifests over README examples if they disagree.

## CI Parity

- Forgejo Actions live in `.forgejo/workflows`, not `.github/workflows`.
- CI Go validation runs `go mod tidy`, `golangci-lint run ./cmd/... ./internal/...`, `golangci-lint fmt`, `go test ./...`, then checks `git status --short`.
- SDK validation builds the plugin with `FORGEJO_PROVIDER_VERSION` from `svu next`, exports `schema.json`, runs `scripts/gen-sdk.sh`, then checks generated files are clean.
- Tool versions are pinned in both `.mise.toml` and workflows; use Go `1.26.2`, Pulumi `3.232.0`, golangci-lint `v2.11.4`, svu `v3.4.0`, and GoReleaser `2.15.4` unless those files change.

## Testing Notes

- Existing tests are unit tests only; they exercise diffs and Pulumi dry-run preview behavior without a live Forgejo service.
- Provider config is `forgejo:url`/`forgejo:token`, with env fallbacks `FORGEJO_URL`/`FORGEJO_TOKEN`; `token` is secret and the URL must be absolute.
