# Agent Token Diet Implementation Plan

> **For agentic workers:** Execute inline in the current session. Steps use checkbox (- [ ]) syntax.

**Goal:** Cut per-session agent context to a small map + recipes, make verification risk-based, and automate the two mechanical steps agents forget.

**Architecture:** Tooling first (scaffold `sqlc.yaml` registration, `task sync`, `task check:quick`) so the rewritten docs can name real commands; then docs (verification + ADR-0013, recipes, AGENTS.md map); then one integrated gate.

**Tech Stack:** Go (scaffold), Node ESM + `node:test` (check script), Task, Bun test (Taskfile tests), Markdown.

**Spec:** `docs/superpowers/specs/2026-11-25-agent-token-diet-design.md`

## Global Constraints

- Work on `main`; one conventional commit per batch; no push, no merge.
- Size targets: `AGENTS.md` ≤ 4 KB, `.agent/verification.md` ≤ 3 KB, `.agent/recipes.md` ≤ 5 KB.
- Safety rules kept verbatim in meaning: no existing/user DB, no real `.env`/secrets, no operational/remote migrations, no deploy, no Compose stack, no real mail, no backend listeners against real config.
- Binding architecture rules kept in AGENTS.md: UI primitives via `@repo/ui/<name>`; `//vkit:handler`, one operation per file, no `routes.go`; never hand-edit generated files; queries via API-owned persistence, mutations via application usecases; DBML updated with every schema migration; config isolation / no feature-level env reads; ADR-0001 phasing.
- No new generators beyond `sqlc.yaml` registration; existing `gen`, `sync:web|bo|api` stay.
- Every file path named in a doc must exist.

## Review Focus

- `sqlc.yaml` already lists the domain (re-run after deleting `service.go`) → no duplicate entry. Test in Task 1.1.
- `sqlc.yaml` with no `queries:` list → clear error, nothing written to `service.go` either. Test in Task 1.1.
- `check:quick` with a deleted Go/TS file → owner still checked, no crash on missing path. Test in Task 1.3.
- `check:quick` with a renamed file (porcelain `-z` gives new path then old path) → maps the new path only. Test in Task 1.3.
- `check:quick` after a change in a shared package (`packages/ui`, `packages/db`) → consumers are checked too, not just the package. Test in Task 1.3.
- `check:quick` with files outside any owner (root `README.md`, `docs/**`) → skipped, `ok: nothing to check`. Test in Task 1.3.

---

## Batch 1 — Tooling

Goal: the three commands the docs will reference exist and are tested.
Dependencies: none. Acceptance gate: `cd packages/tools && go test ./internal/scaffold/ ./cmd/scaffold/ -count=1`, `node --test scripts/check-quick.test.mjs`, `bun test scripts/taskfile.test.ts` all pass; `task --list` shows `sync` and `check:quick`.

### Task 1.1: `create:usecase` registers new domains in `sqlc.yaml`

**Files:**

- Modify: `packages/tools/internal/scaffold/scaffold.go` (`Usecase`)
- Modify: `packages/tools/internal/scaffold/templates/service.go.tmpl`, `usecase.sql.tmpl` (drop "add to sqlc.yaml" instructions; keep "run `task gen:db` after writing queries")
- Modify: `packages/tools/cmd/scaffold/main.go` (usecase `next:` line → "write the SQL, run `task gen:db`"; print `registered <domain> in packages/application/sqlc.yaml` when it happened)
- Modify: golden files `packages__application__projects__{service.go,create.sql,archive.sql}.golden` via `-update`
- Test: `packages/tools/internal/scaffold/scaffold_test.go`

**Interfaces:**

- Produces: `func registerQueries(path, domain string) (bool, error)` — returns true when appended. `Usecase` calls it only when it creates `service.go`, **before** rendering any file, and includes the yaml path in `written` when true.

- [ ] Step 1: Tests. `fakeRoot` also writes `packages/application/sqlc.yaml` with `queries:\n      - staff\n      - customers\n    gen:`.
  - `TestRegisterQueriesAppends`: after `Usecase(root, {Domain:"projects", Name:"Create"})`, yaml contains `      - customers\n      - projects\n    gen:`.
  - `TestRegisterQueriesIdempotent`: yaml already listing `projects`, `registerQueries` returns `false` and file unchanged.
  - `TestRegisterQueriesMissingList`: yaml without `queries:` → `Usecase` errors containing `queries`, and `packages/application/projects/service.go` does not exist.
  - Existing usecase golden test still passes (second usecase in same domain does not touch yaml).
- [ ] Step 2: `go test ./internal/scaffold/ -run Register -count=1` → FAIL (undefined).
- [ ] Step 3: Implement: find the `queries:` line, then the contiguous `- name` lines under it; insert `<same indent>- <domain>` after the last; skip if any entry equals domain. Text edit, no YAML dependency.
- [ ] Step 4: Update templates/main.go, run `go test ./internal/scaffold/ -count=1 -update` then without `-update` → PASS; review golden diff is only the comment text.

### Task 1.2: `task sync`

**Files:**

- Modify: `Taskfile.yml` (after `sync:api`), `scripts/taskfile.test.ts` (add `"sync"` and `"check:quick"` to the required list)

- [ ] Step 1: Add `"sync"` to the list in `scripts/taskfile.test.ts`; `bun test scripts/taskfile.test.ts` → FAIL.
- [ ] Step 2: Add task `sync` — desc `Regenerate everything after API changes: handlers, schemas, Web/BO SDKs, route trees`; cmds `[task gen:bo-client, task gen:web-client, task gen:routes, task gen:bo-routes]` (clients depend on `gen:openapi` → `gen:handlers` through turbo, as `sync:web` already relies on).
- [ ] Step 3: `bun test scripts/taskfile.test.ts` → PASS.

### Task 1.3: `task check:quick`

**Files:**

- Create: `scripts/check-quick.mjs`, `scripts/check-quick.test.mjs`
- Modify: `Taskfile.yml`

**Interfaces:**

- Produces (exported from `scripts/check-quick.mjs`):
  - `changedPaths(porcelain: string): string[]` — parses NUL-separated `git status --porcelain=v1 -z -uall` output (entries `XY path`; for `R`/`C` the next NUL field is the old path and is skipped).
  - `plan(paths: string[], exists: (p: string) => boolean): {label: string, cwd: string, cmd: string[]}[]` — deduplicated, stable order.
  - CLI (`import.meta.main`-style guard via `process.argv[1]`): runs each step with `spawnSync`, captured output.

Owner rules for `plan`:

- `*.go`, `go.mod`, `*.sql` → nearest ancestor dir containing `go.mod` (via `exists`) → `go vet ./...` in that dir (vet type-checks incl. tests; no separate build). If that module is under `packages/`, also vet `apps/web-api`, `apps/bo-api`, `apps/api`, `apps/worker`.
- `*.ts|*.tsx|*.mts` under a workspace from the fixed list `apps/{web,bo,landing,storybook}`, `packages/{ui,ui-email,ui-utils,brand,config/js}` → `bun --no-env-file run check-types` in that dir. A `packages/*` workspace also adds `apps/web` and `apps/bo` (consumers). `apps/landing|storybook` add one root step `bun --no-env-file run check:ui-primitives` (web/bo `check-types` already run it).
- `scripts/*.ts` → `bun --no-env-file x tsc --noEmit -p scripts/tsconfig.json` at root.
- anything else → ignored.

Output: success → `ok: <n> checks (<labels joined ", ">)`; nothing mapped → `ok: nothing to check`; first failure → print `FAIL <label>: <cmd>` + its stdout/stderr, exit 1 (stop at first failure).

- [ ] Step 1: Tests in `scripts/check-quick.test.mjs` (`node:test`):
  - `changedPaths(" M apps/bo-api/main.go\0R  packages/db/b.go\0a.go\0?? scripts/x.ts\0")` → `["apps/bo-api/main.go","packages/db/b.go","scripts/x.ts"]`.
  - `plan` with go.mod at `apps/bo-api` only: two files under `apps/bo-api/internal/...` → exactly one vet step (cwd `apps/bo-api`).
  - `plan` with a file in `packages/db` (go.mod there) → vet steps for `packages/db` + the four apps, no duplicates when an app file is also changed.
  - deleted file path (mapped by directory, file absent) → step produced.
  - `apps/bo/src/x.tsx` alone → one `check-types` step (cwd `apps/bo`), no root ui-primitives step.
  - `packages/ui/src/y.tsx` → `check-types` for `packages/ui`, `apps/web`, `apps/bo`.
  - `apps/landing/src/z.tsx` → `check-types` for `apps/landing` + root `check:ui-primitives`.
  - `["README.md","docs/a.md"]` → `[]`.
- [ ] Step 2: `node --test scripts/check-quick.test.mjs` → FAIL.
- [ ] Step 3: Implement. CLI reads `git status --porcelain=v1 -z -uall` via `spawnSync`.
- [ ] Step 4: Taskfile: `check:quick` placed **after** `check:ui-primitives` (the existing test forbids `go run|test|vet` text between `check:go:` and `check:ui-primitives:`) — desc `Typecheck only what git status changed, plus its consumers; one line on success`; `dir: "{{.ROOT_DIR}}"`; cmds `[node scripts/check-quick.mjs]`. Add `"check:quick"` to `scripts/taskfile.test.ts`.
- [ ] Step 5: `node --test scripts/check-quick.test.mjs` and `bun test scripts/taskfile.test.ts` → PASS; `task check:quick` on the current dirty tree → prints `ok: …` or a real failure (investigate, don't mask).

- [ ] **Batch 1 gate + commit:** run the gate commands; `git add` the touched files; `git commit -m "feat(tools): register sqlc domains, add task sync and check:quick"`.

---

## Batch 2 — Docs

Goal: agents read a map, a risk table and recipes instead of ~100 KB.
Dependencies: Batch 1 command names (`task sync`, `task check:quick`, sqlc auto-registration). Invalidates: nothing in Batch 1. Acceptance gate: size targets met (`wc -c`), every backticked repo path in the three docs exists (Task 2.4 script), `bunx prettier --check AGENTS.md .agent docs/adr docs/superpowers`.

### Task 2.1: `.agent/verification.md` + ADR-0013

**Files:** Rewrite `.agent/verification.md`; create `docs/adr/0013-risk-based-verification.md` (Status: Accepted, 2026-11-25; context = six dated exceptions + "never run tests" cost tokens/permissions; decision = spec §2; consequences = exceptions removed, see git history); add row to `docs/adr/README.md`.

- [ ] Step 1: Write verification.md per spec §2: risk table (docs → prettier; Go handler/usecase → `task check:quick` + focused `go test` when logic/auth/money/concurrency; SQL/migration → `task check:db-schema`, `task gen:db`, `task check:db:integration`; TS logic → check:quick + focused `bun test`; UI → check:quick + Storybook 6006 browser review; generators → `task check:scaffold`; auth/security → existing `task check:auth:*`, `check:bo:security`, `check:web-api:security`), allowed-without-asking list, never-without-instruction list, Storybook single-instance paragraph, reporting rule, the resource rule (one Testcontainers suite at a time, affected suite only) and the anti-accretion rule (spec §2). Point counted-gate mechanics to `packages/tools/authtest/README.md`.
- [ ] Step 2: Write ADR-0013 and index row; `docs/adr/README.md` already fails prettier (pre-existing), so run `bunx prettier --write docs/adr/README.md` and keep that formatting in the commit. Grep for other docs referencing the removed exception names (`grep -rn "exception" AGENTS.md .agent apps/*/README.md packages/*/README.md`) and repoint them to `.agent/verification.md` with one line each.

### Task 2.2: `.agent/recipes.md`

**Files:** Create `.agent/recipes.md`.

- [ ] Step 1: Sections per spec §3 (BO query, BO mutation, Web API query, public API, new domain usecase, migration, BO screen). Each: golden files, steps (`task create:*` → edit → `task gen:db` if SQL → `task sync` → `task check:quick`), and one "watch out" line where real code differs from the scaffold (inline struct + `ItemPath`, `customerError`, customers domain groups usecases per resource file).
- [ ] Step 2: Add the steps valcentra history shows agents repeatedly reworked (hotspots: nav files touched 11×, dummy fixtures/types 6–7×):
  - Screen recipes end with "register the page in navigation": BO `apps/bo/src/app/_authenticated/-components/navigation-links.tsx`; Web `apps/web/src/app/w/$organizationSlug/-components/navigation.ts` (+ its `navigation.test.ts`).
  - Endpoint recipes have an ADR-0001 dummy-phase variant: golden `apps/web-api/internal/handlers/v1/api_keys_list.go` (data from `packages/fixtures`, list helpers in `dummy_page.go`); real persistence only in the real-functionality phase.
- [ ] Step 3: Read each golden file briefly to confirm it matches the claimed shape; swap for a better one if not.

### Task 2.3: `AGENTS.md` map

**Files:** Rewrite `AGENTS.md`; modify `.agent/backend/README.md` (point to recipes; drop text duplicated from verification).

- [ ] Step 1: Write per spec §1 (start-here, foundation docs, ADR-0001, repo-shape table, binding rules from Global Constraints, routing table, commit rule). Move any rule that exists only in AGENTS.md and isn't binding-for-all into the matching `.agent/*` file rather than dropping it — diff old vs new and list each removed paragraph with where it now lives.
- [ ] Step 2: `wc -c AGENTS.md .agent/verification.md .agent/recipes.md` within targets.

### Task 2.4: Path check

- [ ] Step 1: Run `grep -ohE '\`[a-zA-Z0-9_./$-]+/[a-zA-Z0-9_./$-]+\`' AGENTS.md .agent/*.md | tr -d '\`' | sort -u | while read p; do [ -e "$p" ] || echo MISSING $p; done` (ignore glob/placeholder entries, judged by eye) → no real missing paths.
- [ ] **Batch 2 gate + commit:** gate commands; include `docs/superpowers/` (spec + plan) in this commit; `git commit -m "docs(agents): map-style AGENTS.md, risk-based verification (ADR-0013), recipes"`.

---

## Batch 3 — Integrated gate

Goal: prove the whole change together. Dependencies: Batches 1–2. Invalidates: Batch 1 evidence if any tool file changed in Batch 2.

- [ ] `task check:scaffold` (golden + real compile of scaffold output in a temp copy) → PASS.
- [ ] `task check:go` → PASS.
- [ ] `task check-types` → PASS (pre-existing failures, if any, reported with output, not masked).
- [ ] `node --test scripts/check-quick.test.mjs && bun test scripts/taskfile.test.ts` → PASS.
- [ ] `task check:quick` → `ok: …`.
- [ ] Sizes and path check from Batch 2 re-run → within targets, no missing.
- [ ] Commit only if anything changed: `git commit -m "chore: fix integrated gate findings"`.
