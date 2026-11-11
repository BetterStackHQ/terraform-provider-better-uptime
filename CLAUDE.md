# CLAUDE.md

Guidance for Claude Code (claude.ai/code) when working in this repository (the Better Uptime Terraform provider).

## Frozen: no PRs to this repo

This provider is frozen (Simon, 2026-10-10): don't open PRs or push branches here, even for a finished change. The next provider is one unified `betterstack` provider ([SH-640](https://linear.app/betterstack/issue/SH-640)), and new Terraform behaviour goes on top of it, not into this repo. File a gap as a Linear issue in Triage (team U): link the API PR, say it is to be built on the unified provider, and @-mention Pawel. Example: [U-9478](https://linear.app/betterstack/issue/U-9478), whose closed PR [#248](https://github.com/BetterStackHQ/terraform-provider-better-uptime/pull/248) is kept only as a reference for the shape.

## Exercise every new feature in an example

When you add a resource, data source, attribute, or any new provider capability, use it in at least one config under `examples/` — new provider features go into the resource's docs-integrated example `examples/resources/<type>/resource.tf` (data sources under `examples/data-sources/`), exercised by the combined E2E job. The E2E matrix applies, re-plans (expecting no diff), and destroys every example config against the live API, so a feature that appears in no example is never covered end-to-end.

## Versioning: bump `VERSION` to the intended release version

**How releases work:** publishing to the Terraform registry is **solely tag-based** — pushing a `vX.Y.Z` git tag fires `.github/workflows/release.yml`, which builds and publishes that version. In practice the tag is pushed onto the squash-merged PR commit on master, so the PR itself is what gets released. The `Makefile`'s `VERSION` plays no part in publishing, but it drives local `make terraform` and E2E: the provider is built at `VERSION`, then `terraform init` runs against each example's `versions.tf` constraint, so **a constraint ahead of `VERSION` fails `init`**.

**The rule:** in every PR, set `VERSION` in the `Makefile` to the **intended release version** — the version of the **next git tag** this PR should be released in, so always **higher than the latest git tag** (`git describe --tags --abbrev=0`; usually its next patch, a minor bump for bigger changes). Never derive it from the current `VERSION` value, which may be stale — it had drifted to `0.20.19` while `v0.21.1` was already tagged. The only exception is a change unrelated to a release, such as a CI, instructions, or tests-only update — those don't bump anything.

**When an example starts using a brand-new capability**, also raise `version = ">= X.Y.Z"` in `examples/versions.tf` to the same intended release version, in the same commit as the `Makefile` bump. Registry users on an older provider then get a clean "update your provider" error instead of a confusing "unsupported parameter" one — and E2E `init` stays green because the constraint never gets ahead of `VERSION`.
