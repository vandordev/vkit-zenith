# Contributing

Read `AGENTS.md`, relevant `.agent` guidance and approved spec/plan before editing.
Inspect branch and `git status --short`; preserve unrelated work and real secrets.
Keep changes within the owning runtime/package. No automatic push, merge or branch deletion.

This foundation has web plus independent public/web Go APIs, not embedded APIs,
auth, databases or job infrastructure. Business-layer design is deferred.
Generated files have explicit owners: regenerate, inspect and commit them; never
hand-edit FX composition, route trees or Hey API output.

Use `task install`, `task gen` and `task gen:routes` as needed. Verification follows
[the two tracks](.agent/verification.md): native TS typecheck/Go compilation for
system changes; affected typecheck plus Storybook inspection for UI. Do not run
system tests, Go vet, general build/quality gates, backend services, containers,
migrations or deployment checks without new explicit permission. Preserve retained
test fixtures without executing them. Report runtime limitations honestly.

Shared primitives belong in `packages/ui`, identity/tokens in
`packages/brand`. UI is developed as typed prop-driven presentation with colocated
stories before route/client integration. Never use real APIs in stories.
