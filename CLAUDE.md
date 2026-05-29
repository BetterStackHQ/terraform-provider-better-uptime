# CLAUDE.md

Guidance for Claude Code (claude.ai/code) when working in this repository (the Better Uptime Terraform provider).

## Bump the version on every functional change

When you add or change a resource, data source, or any provider behavior, bump the version **in the same commit** — two files, kept in sync:

1. **`Makefile`** — bump the patch in `VERSION := X.Y.Z` (e.g. `0.20.17` → `0.20.18`).
2. **`examples/advanced/versions.tf`** — set the required-provider `version = ">= X.Y.Z"` to the same value.

Use the next patch above the current value on the default branch (it sits one ahead of the latest released git tag). Skipping the bump means the published provider — and the example that pins it — won't pick up your change.

The bump rides in the feature commit (PRs are squash-merged). The release itself — the git tag plus a separate "Bump version to vX" commit — happens independently; you only bump these two files.
