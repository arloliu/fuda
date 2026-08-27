# 500 - Development Workflow

## Before Commit
1. Run `make lint` — Fix all issues.
2. Run `make test` — All tests must pass.
3. Verify docs are updated if API changed.

## Git Conventions
- **Branches:** `feat/`, `fix/`, `docs/`, `chore/`, `test/`.
- **Commits:** Conventional format. Present tense. First line < 50 chars.
    - `feat: add dotenv support`
    - `fix: handle empty env values`

## Release
1. Add a `## [x.y.z] - YYYY-MM-DD` section to `CHANGELOG.md` (Keep a Changelog headings; add a compare link at the bottom).
2. Create an annotated tag `vx.y.z` whose body has "Highlights" and "Migration notes".
3. `gh release create vx.y.z` with a Markdown body derived from the changelog entry.

## Code Review Checklist
- [ ] Correctness
- [ ] Performance (no unnecessary allocs)
- [ ] Test coverage for new code
- [ ] Docs updated

## Environment
- **Dev:** `go run`, `go test`
- **Prod:** Logging, health checks, graceful shutdown.
