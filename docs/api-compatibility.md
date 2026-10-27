# Forgejo API compatibility

Checked **2026-10-27** against Forgejo **17.0.1**, the latest stable release
published on 2026-10-26. This note records the implemented existing-resource updates.
[Official releases](https://forgejo.org/releases/)

## Baseline

- `https://forgejo.siron.casa/api/v1/version` returned
  `17.0.1+gitea-1.22.0`. [Public version endpoint](https://forgejo.siron.casa/api/v1/version)
- **14 definitions matched exactly** between the deployed schema and the tagged
  upstream schema: `Repository`, `CreateRepoOption`, `EditRepoOption`,
  `InternalTracker`, `ExternalTracker`, `ExternalWiki`, branch protection and its
  create/edit options, tag protection and its create/edit options, `PushMirror`,
  and `CreatePushMirrorOption`.
  [Deployed schema](https://forgejo.siron.casa/swagger.v1.json) · [Upstream v17.0.1 schema](https://codeberg.org/forgejo/forgejo/src/tag/v17.0.1/templates/swagger/v1_json.tmpl)
- The pinned SDK is `codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v3 v3.0.0`.
  Its handwritten types lack current request/response members, including
  creation-time `object_format_name`, the two branch settings below, and repository
  response fields. The provider extends the SDK wire types to cover these fields.
  [SDK repository types](https://codeberg.org/mvdkleijn/forgejo-sdk/src/commit/6502d5e6039c3716dd9ce6e703b9bf6efe870548/forgejo/repo.go) · [SDK branch types](https://codeberg.org/mvdkleijn/forgejo-sdk/src/commit/6502d5e6039c3716dd9ce6e703b9bf6efe870548/forgejo/repo_branch_protection.go)

## Added settings inventory

The eight editable repository settings belong to both `Repository.settings` and
`RepositorySettings`. All eight are writable with the pinned SDK.
[Current repository API structs](https://codeberg.org/forgejo/forgejo/src/tag/v17.0.1/modules/structs/repo.go)

| Pulumi property | API field | Repository API read-back |
| --- | --- | --- |
| `ignoreWhitespaceConflicts` | `ignore_whitespace_conflicts` | Available |
| `allowManualMerge` | `allow_manual_merge` | Unavailable |
| `autodetectManualMerge` | `autodetect_manual_merge` | Unavailable |
| `allowRebaseUpdate` | `allow_rebase_update` | Available |
| `defaultAllowMaintainerEdit` | `default_allow_maintainer_edit` | Available |
| `defaultUpdateStyle` | `default_update_style` | Available |
| `mirrorInterval` | `mirror_interval` | Available |
| `enablePrune` | `enable_prune` | Unavailable |

PR settings require an enabled pull-request unit. `mirrorInterval` and
`enablePrune` apply only to **pull mirrors**; Forgejo ignores them on ordinary
repositories. [Repository update handlers](https://codeberg.org/forgejo/forgejo/src/tag/v17.0.1/routers/api/v1/repo/repo.go)

Six creation-only repository options are added: `gitignores`, `issueLabels`
(`issue_labels`), `license`, `readme`, `trustModel` (`trust_model`), and
`objectFormatName` (`object_format_name`). `readme` names a template, not literal
content. README, license, and gitignore generation require `auto_init`; initial
issue labels do not. [Creation implementation](https://codeberg.org/forgejo/forgejo/src/tag/v17.0.1/services/repository/create.go)

The implementation preserves creation-only options in state through refresh,
without automatic destructive replacements. Changing these options does not
retroactively initialize or reconfigure an existing repository.

Two branch-protection settings are added: `applyToAdmins` (`apply_to_admins`) and
`ignoreStaleApprovals` (`ignore_stale_approvals`). The former restricts repository
administrator bypass, not instance-administrator bypass; the latter excludes stale
approvals without dismissing them and is redundant when stale approvals are dismissed.
[Branch API](https://codeberg.org/forgejo/forgejo/src/tag/v17.0.1/modules/structs/repo_branch.go) · [Administrator bypass](https://codeberg.org/forgejo/forgejo/src/tag/v17.0.1/services/pull/check.go) · [Approval counting](https://codeberg.org/forgejo/forgejo/src/tag/v17.0.1/models/issues/pull.go)

Organization additions are `email` and `repoAdminChangeTeamAccess`
(`repo_admin_change_team_access`), confirmed in deployed `CreateOrgOption`,
`EditOrgOption`, and `Organization`. [Deployed definitions](https://forgejo.siron.casa/swagger.v1.json)

Team settings, SSH keys, Actions variables, and Actions secrets already cover their
writable settings. Team updates now accept the API's 255-character name limit and
normalize the SDK's `owner` permission alias to `admin`. Secret names are matched
case-insensitively and stored in uppercase.
[Team structs](https://codeberg.org/forgejo/forgejo/src/tag/v17.0.1/modules/structs/org_team.go) · [Secret identity](https://codeberg.org/forgejo/forgejo/src/tag/v17.0.1/models/secret/secret.go)

## Wire names and supported values

Use **`POST /orgs/{org}/repos`** instead of the SDK's deprecated `/org/{org}/repos`.
Use branch-protection **`rule_name`** instead of deprecated `branch_name`.
Keep public API spellings such as `whitelist_*` and singular
`approvals_whitelist_username`. [Current API schema](https://forgejo.siron.casa/swagger.v1.json)

| Setting | Supported values |
| --- | --- |
| Default merge style | `merge`, `rebase`, `rebase-merge`, `squash`, `fast-forward-only`, `manually-merged`, `rebase-update-only` |
| Default update style | `merge`, `rebase` |
| Trust model | `default`, `collaborator`, `committer`, `collaboratorcommitter` |
| Object format | `sha1`, `sha256` |
| External tracker style | `numeric`, `alphanumeric`, `regexp` |

These are the API's accepted/documented values. `rebase-update-only` is explicitly
an update operation rather than a real merge style. Defaults do not enable the
corresponding operation; rebase updates require `allow_rebase_update`.
[API validation](https://codeberg.org/forgejo/forgejo/src/tag/v17.0.1/modules/structs/repo.go) · [Constants](https://codeberg.org/forgejo/forgejo/src/tag/v17.0.1/models/repo/git.go) · [Operation eligibility](https://codeberg.org/forgejo/forgejo/src/tag/v17.0.1/models/repo/repo_unit.go)

## Read-back and lifecycle limits

- Additional decoding covers **seven** repository response fields for drift detection:
  `default_delete_branch_after_merge`, `allow_fast_forward_only_merge`,
  `globally_editable_wiki`, `wiki_branch`, `allow_rebase_update`,
  `default_allow_maintainer_edit`, and `default_update_style`.
- `allowManualMerge`, `autodetectManualMerge`, `enablePrune`, and creation-only
  `trustModel` cannot be observed through the repository API. Preserve configured
  values; imports cannot discover them, and refresh cannot detect their drift.
  [Response struct](https://codeberg.org/forgejo/forgejo/src/tag/v17.0.1/modules/structs/repo.go) · [Response conversion](https://codeberg.org/forgejo/forgejo/src/tag/v17.0.1/services/convert/repository.go)
- Tag protection and push mirrors already expose all current writable fields.
  Clear protection allowlists/status contexts with `[]`; `null` preserves them.
  [Schema](https://forgejo.siron.casa/swagger.v1.json) · [Branch PATCH](https://codeberg.org/forgejo/forgejo/src/tag/v17.0.1/routers/api/v1/repo/branch.go) · [Tag PATCH](https://codeberg.org/forgejo/forgejo/src/tag/v17.0.1/routers/api/v1/repo/tag.go)
- Push-mirror `interval` is **required**: empty input is rejected by
  Forgejo, while explicit `0s` disables periodic synchronization. Nonzero values
  must meet the server-configured minimum. Compare normalized durations to avoid
  replacing a mirror merely because `1h` is returned as `1h0m0s`.
  [Interval validation](https://codeberg.org/forgejo/forgejo/src/tag/v17.0.1/routers/api/v1/repo/mirror.go) · [Scheduling](https://codeberg.org/forgejo/forgejo/src/tag/v17.0.1/models/repo/pushmirror.go) · [Normalization](https://codeberg.org/forgejo/forgejo/src/tag/v17.0.1/services/convert/mirror.go)
- Push mirrors have no update API, so changes require replacement. Replacing an
  SSH mirror rotates its generated key; update the destination deploy key.
  SSH authentication excludes username/password authentication and does not mirror
  LFS objects. [Mirror API](https://codeberg.org/forgejo/forgejo/src/tag/v17.0.1/routers/api/v1/repo/mirror.go) · [Official mirror documentation](https://forgejo.org/docs/v17.0/user/repo-mirror/)

## Verification

`mise run check` passed (formatting, module tidy, lint, tests, and build). Schema
export and SDK generation passed for Node.js, Go, .NET, and Java; inspection artifacts
are under `.tmp/`. Java generation used `PULUMI_IGNORE_AMBIENT_PLUGINS=true` to bypass
an outdated local shim. SDK packages were generated, not compiled or published.

Tests use local HTTP contract mocks for payloads, read-back, and lifecycle behavior.
No live authenticated CRUD was performed; instance-specific permissions, configured
mirror limits, and remote SSH access remain deployment-time checks.
