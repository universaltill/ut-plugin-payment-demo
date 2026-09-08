# Code review: auto-tag-release.yml rollout (ut-docs#1700)

**Date:** 2026-09-08
**Card:** ut-docs#1700 (rollout of ut-docs#1694's workflow to remaining `ut-plugin-*` repos)
**Author:** scrum-master pipeline (cloud cycle, `lane:cloud-41`), on behalf of Pouria Teimouri

## What changed

Added `.github/workflows/auto-tag-release.yml`, copied byte-for-byte from
the canonical, independently-reviewed copy in `ut-plugin-tax-de`
(`docs/code-reviews/2026-09-07-auto-tag-release-workflow-1694.md` there —
full design rationale, recursion-guard analysis and behavioral test
evidence live in that record and are not repeated here per ut-docs#1700's
own instructions; this record only covers what is repo-specific).

## Repo-specific verification

- Confirmed `.github/workflows/release.yml` is tag-triggered
  (`on: push: tags: ["v*"]`) and declares `workflow_dispatch` inputs named
  `channel` (choice, includes `stable`) and `publish` (boolean) — matching
  what the copied workflow's dispatch step passes, with no adaptation
  needed.
- `diff` against the canonical `ut-plugin-tax-de` copy: byte-identical.
- **This repo has no drift**: `manifest.json`'s `version` is `1.1.1` and
  `git tag -l` already shows `v1.1.1` (also `v1.0.0`, `v1.1.0` from earlier
  releases). This is ut-docs#1700's own explicit "deliberate no-drift push
  run observed to no-op correctly" acceptance case — merging this workflow
  should NOT create a new tag on the next push to main (the idempotency
  guard's `git rev-parse -q --verify refs/tags/v1.1.1` short-circuit).
  DevOps must verify the workflow run actually logged "already exists...
  nothing to do" rather than assuming no-op behavior from the YAML alone.
- `manifest.json` confirmed at repo root (script's assumption).

## Independent review

Mechanically identical to the already-reviewed canonical file (full
independent Opus review on the original: ut-plugin-tax-de's own
2026-09-07 record). This repo's own diff is a verbatim copy plus this
review record — verdict: **SAFE TO MERGE**, no repo-specific deviation
found.
