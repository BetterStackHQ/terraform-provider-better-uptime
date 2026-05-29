# CLAUDE.md

Guidance for Claude Code (claude.ai/code) when working in this repository (the Better Uptime Terraform provider).

## Exercise every new feature in an example

When you add a resource, data source, attribute, or any new provider capability, use it in at least one config under `examples/` — new provider features generally go into `examples/advanced/`. The E2E matrix applies, re-plans (expecting no diff), and destroys every example config against the live API, so a feature that appears in no example is never covered end-to-end.

## Bump the version when an example requires it

Using a brand-new capability in an example means that example now needs a newer provider than it pins. When that happens — and only then, not for every functional change — bump the version **in the same commit**, two files kept in sync:

1. **`Makefile`** — bump the patch in `VERSION := X.Y.Z` (e.g. `0.20.17` → `0.20.18`).
2. **that example's `versions.tf`** (usually `examples/advanced/versions.tf`) — set the required-provider `version = ">= X.Y.Z"` to the same value.

Use the next patch above the current value on the default branch (it sits one ahead of the latest released git tag). Keep the two in sync: E2E builds the provider at `VERSION` and runs `terraform init` against the example's constraint, so a constraint ahead of `VERSION` fails init. Skipping the bump when it is needed means the published provider — and the example that pins it — won't pick up your change.

The bump rides in the feature commit (PRs are squash-merged). The release itself — the git tag plus a separate "Bump version to vX" commit — happens independently; you only bump these two files.
