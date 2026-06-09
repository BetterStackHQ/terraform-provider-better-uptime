# CLAUDE.md

Guidance for Claude Code (claude.ai/code) when working in this repository (the Better Uptime Terraform provider).

## Exercise every new feature in an example

When you add a resource, data source, attribute, or any new provider capability, use it in at least one config under `examples/` — new provider features generally go into `examples/advanced/`. The E2E matrix applies, re-plans (expecting no diff), and destroys every example config against the live API, so a feature that appears in no example is never covered end-to-end.

## Versioning: git tags ship releases; keep the `Makefile` in sync

Published provider versions on the Terraform registry are **solely tag-based**: pushing a `vX.Y.Z` git tag fires `.github/workflows/release.yml`, which builds and publishes. **That tag push is the only step that actually ships a release** — the `Makefile`'s `VERSION` does not determine what's published.

`VERSION` still drives local `make terraform` and E2E: E2E builds the provider at `VERSION`, then runs `terraform init` against each example's `versions.tf` constraint, so **a constraint ahead of `VERSION` fails `init`**. Keep `VERSION` current with the release tags rather than letting it fall behind (it had drifted to `0.20.19` while `v0.21.1` was already tagged).

So when an example starts using a brand-new capability, bump both in the same commit — anchored to the **latest git tag**, not the possibly-stale `Makefile` value:

1. **that example's `versions.tf`** (usually `examples/advanced/versions.tf`) — raise `version = ">= X.Y.Z"` to the release that introduces the feature (the next version above the latest tag), so registry users on an older provider get a clean "update your provider" error instead of a confusing "unsupported parameter" one.
2. **`Makefile`** — set `VERSION` to that same version, so local builds and E2E `init` stay green.
