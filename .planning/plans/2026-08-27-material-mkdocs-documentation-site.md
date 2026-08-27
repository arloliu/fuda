# Material for MkDocs Documentation Site Implementation Plan

> **For implementers:** Complete tasks in order and update the checkboxes as work finishes.

**Goal:** Publish a beginner-first Fuda documentation site with Material for MkDocs and GitHub Pages.

**Architecture:** `docs/` holds all public site content.
Material reads `mkdocs.yml` and creates the deployable site artifact.
A dedicated workflow validates pull requests and deploys only from `main`.
Executable programs in `examples/` remain the proof for tutorial code.

**Tech stack:** Material for MkDocs, pinned Python requirements, GitHub Pages Actions, Markdown, and Go tests.

**Spec:** `.planning/specs/2026-08-27-documentation-site-design.md`

## Global constraints

- Write all public pages in conversational English.
- Keep learning pages readable in five to ten minutes.
- Use semantic linefeeds in prose.
- Keep plans and specs outside `docs/`.
- Treat `examples/` as the source of truth for runnable programs.
- Verify behaviour claims against `internal/loader/engine.go` and the tag processors.
- Do not publish secrets, `.env` files, or generated output.
- Run `make lint` before every commit.

## File structure

| Path | Purpose |
| --- | --- |
| `mkdocs.yml` | Site metadata, navigation, Material features, and Markdown extensions. |
| `requirements-docs.txt` | Exact Python packages for local and CI documentation builds. |
| `.github/workflows/docs.yml` | Strict build validation and GitHub Pages deployment. |
| `docs/getting-started/` | Installation, first configuration, and value selection. |
| `docs/concepts/` | Configuration sources, validation, and data types. |
| `docs/guides/` | Focused task guides. |
| `docs/integrations/` | Watcher, Vault, and custom resolver guides. |
| `docs/tooling/` | `fuda-doc` guide. |
| `docs/reference/` | Tags, Builder options, loading behaviour, and errors. |
| `docs/assets/images/` | Public site images. |
| `docs/stylesheets/extra.css` | Small branding adjustments. |
| `tests/documentation_examples_test.go` | Regression proof for the first tutorial. |

### Task 1: Establish the site build and Pages deployment

**Files:**

- Create: `requirements-docs.txt`
- Create: `mkdocs.yml`
- Create: `docs/index.md`
- Create: `.github/workflows/docs.yml`
- Create: `docs/stylesheets/extra.css`
- Move: `docs/fuda-logo.png` to `docs/assets/images/fuda-logo.png`

**Consumes:** Existing workflow conventions in `.github/workflows/go.yml`.

**Produces:** A strict site build and a Pages artifact deployment pipeline.

- [ ] Pin one exact, compatible `mkdocs-material` version in `requirements-docs.txt`.
- [ ] Set site name `Fuda`, site URL `https://arloliu.github.io/fuda/`, and repository URL `https://github.com/arloliu/fuda` in `mkdocs.yml`.
- [ ] Enable search highlighting, search sharing, and code-copy buttons.
- [ ] Enable navigation footer links and light and dark color schemes.
- [ ] Register the moved logo and `stylesheets/extra.css`.
- [ ] Declare only the `Home: index.md` navigation entry in the first task.
- [ ] Enable `admonition`, `attr_list`, `md_in_html`, and `pymdownx.superfences`.
- [ ] Limit `extra.css` to Fuda colors, landing-page logo sizing, and code-block spacing.
- [ ] Create a short `docs/index.md` foundation page that identifies Fuda and links to the repository.
- [ ] Create a workflow that builds on matching pull requests and pushes to `main`.
- [ ] Give the build job `contents: read`, install the pinned requirements, and run `python -m mkdocs build --strict`.
- [ ] Guard the deploy job with `github.event_name == 'push' && github.ref == 'refs/heads/main'`.
- [ ] Give the deploy job `pages: write` and `id-token: write` permissions.
- [ ] Use `actions/configure-pages` and `actions/upload-pages-artifact`.
- [ ] Use `actions/deploy-pages` with the `github-pages` environment.
- [ ] Use a concurrency group that permits one deployment and does not cancel an active one.
- [ ] Run `python3 -m venv .venv-docs`, install requirements, and run `.venv-docs/bin/python -m mkdocs build --strict`.
- [ ] Expect a clean foundation build that creates `site/`.
- [ ] Run `make lint`, stage only task files, and commit `docs: add Material site build and Pages workflow`.

### Task 2: Publish the first-reader onboarding path

**Files:**

- Create: `docs/index.md`
- Create: `docs/getting-started/install.md`
- Create: `docs/getting-started/first-configuration.md`
- Create: `docs/getting-started/choosing-a-value.md`
- Modify: `mkdocs.yml`
- Modify: `docs/index.md`
- Modify: `examples/basic/main.go`
- Modify: `examples/basic/config.yaml`
- Modify: `examples/basic/README.md`
- Create: `tests/documentation_examples_test.go`

**Consumes:** The Material configuration from Task 1.

**Produces:** A complete beginning path and a test for its stated precedence rule.

- [ ] Add `TestDocumentationFirstConfigurationPrecedence` to `tests/documentation_examples_test.go`.
- [ ] Define local `Host string` and `Port int` fields with `yaml`, `default`, and `env` tags.
- [ ] Use `t.Setenv` and `fuda.New().FromBytes(...).Build()` for isolated source and environment inputs.
- [ ] Test file plus default values with `host: 0.0.0.0` and expected port `8080`.
- [ ] Test an `APP_PORT=9090` override against YAML port `3000`.
- [ ] Test empty source with expected defaults `localhost` and `8080`.
- [ ] Run `go test ./tests -run '^TestDocumentationFirstConfigurationPrecedence$' -count=1` before adding the test.
- [ ] Expect no matching test before implementation.
- [ ] Align `examples/basic` so it demonstrates file values, defaults, and an environment override before Builder usage.
- [ ] Write `index.md` with Fuda's purpose, audience, logo, and links to installation and first configuration.
- [ ] Write `install.md` with the dependency command and a link to the first tutorial.
- [ ] Write `first-configuration.md` with one struct, one YAML file, two run commands, and expected output.
- [ ] Write `choosing-a-value.md` with `environment variable -> configuration file -> default`.
- [ ] Expand `mkdocs.yml` with the Getting started navigation section.
- [ ] Use these sections on every onboarding page.
- [ ] Include What you will learn, Before you start, steps, What happened, Common mistakes, and Next step.
- [ ] Run the focused test, `go run .` in `examples/basic`, the environment override run, and the strict MkDocs build.
- [ ] Expect all four checks to pass.
- [ ] Run `make lint`, stage only task files, and commit `docs: add beginner onboarding guides`.

### Task 3: Split the core concepts into short pages

**Files:**

- Create: `docs/concepts/configuration-files.md`
- Create: `docs/concepts/defaults.md`
- Create: `docs/concepts/environment-variables.md`
- Create: `docs/concepts/validation.md`
- Create: `docs/concepts/data-types.md`
- Modify: `docs/getting-started/choosing-a-value.md`
- Modify: `mkdocs.yml`

**Consumes:** The first-reader path from Task 2.

**Produces:** One focused page for each foundational concept.

- [ ] Write configuration-file, default, and environment pages from the actual tag processors.
- [ ] Include a missing field, ordinary file value, and an `APP_`-prefixed lookup.
- [ ] Include an explicit zero value and YAML `null` behaviour.
- [ ] Write validation with one successful load and one failure.
- [ ] Write data types with `time.Duration`, `fuda.Duration`, numeric byte sizes, and `fuda.ByteSize`.
- [ ] Link readers to the tag reference rather than repeating exhaustive tables.
- [ ] Update `choosing-a-value.md` to link to the three source pages.
- [ ] Give every new page one natural next-step link.
- [ ] Expand `mkdocs.yml` with the Core concepts navigation section.
- [ ] Run `.venv-docs/bin/python -m mkdocs build --strict`.
- [ ] Run `go test ./... -run 'Test.*(Default|Env|Validation|Type)' -count=1`.
- [ ] Expect the strict build and narrowed tests to pass.
- [ ] Run `make lint`, stage only task files, and commit `docs: add core configuration concepts`.

### Task 4: Add task-oriented configuration guides

**Files:**

- Create: `docs/guides/dotenv.md`
- Create: `docs/guides/references.md`
- Create: `docs/guides/dsn.md`
- Create: `docs/guides/templates.md`
- Create: `docs/guides/overrides.md`
- Create: `docs/guides/setter-and-scanner.md`
- Modify: `examples/dotenv/README.md`
- Modify: `examples/refs/README.md`
- Modify: `examples/dsn/README.md`
- Modify: `examples/template/README.md`
- Modify: `examples/setter/README.md`
- Modify: `examples/scanner/README.md`
- Modify: `mkdocs.yml`

**Consumes:** The concepts from Task 3.

**Produces:** Searchable how-to pages with executable example links.

- [ ] Document `WithDotEnv`, `WithDotEnvFiles`, `WithDotEnvSearch`, and `DotEnvOverride` in the dotenv guide.
- [ ] State that dotenv loading precedes environment-tag resolution.
- [ ] Explain that `WithOverrides` changes source after templating and before decoding.
- [ ] State that a matching environment tag can still override that value.
- [ ] Document default-resolver `file://`, `http://`, `https://`, and `env://` sources.
- [ ] Explain `refFrom` before its fallback `ref` pattern.
- [ ] Explain declaration order before URI templates and DSN field references.
- [ ] Include a network timeout and a `dsnStrict:"true"` example.
- [ ] Show one `SetDefaults()` example and one `Scan(any) error` example.
- [ ] State the Setter and Scanner order relative to tags and validation.
- [ ] Reduce each touched example README to its exact command, brief description, and site link.
- [ ] Expand `mkdocs.yml` with the How-to guides navigation section.
- [ ] Run every touched example with `go run .` and run the strict MkDocs build.
- [ ] Expect all examples and the build to pass.
- [ ] Run `make lint`, stage only task files, and commit `docs: add configuration how-to guides`.

### Task 5: Add integrations, tooling, and the reference

**Files:**

- Create: `docs/integrations/watcher.md`
- Create: `docs/integrations/vault.md`
- Create: `docs/integrations/custom-resolvers.md`
- Create: `docs/tooling/fuda-doc.md`
- Create: `docs/reference/tags.md`
- Create: `docs/reference/builder-options.md`
- Create: `docs/reference/loading-behaviour.md`
- Create: `docs/reference/errors.md`
- Create: `docs/reference/troubleshooting.md`
- Modify: `vault/README.md`
- Modify: `cmd/fuda-doc/README.md`
- Modify: `mkdocs.yml`

**Consumes:** Core concepts and how-to guides from Tasks 3 and 4.

**Produces:** Advanced guides and an accurate API reference.

- [ ] Use `watcher.New().FromFile(...).Build()` and `Watch(&cfg)` in the watcher guide.
- [ ] Explain file events, polling, `Stop()`, channel closure, and safe application handoff.
- [ ] Use `vault.NewResolver`, `WithAddress`, and one authentication option in the Vault guide.
- [ ] Document `vault:///mount/path#field` without a real secret.
- [ ] Show `Resolve(context.Context, string) ([]byte, error)` and `WithRefResolver` in the resolver guide.
- [ ] Show `fuda-doc` installation and Markdown generation before TUI and utility modes.
- [ ] Derive every command-line flag from `cmd/fuda-doc/main.go`.
- [ ] Move full tag syntax into `reference/tags.md`.
- [ ] List public Builder methods from `fuda.go` in `reference/builder-options.md`.
- [ ] Derive `reference/loading-behaviour.md` from engine code, not the old processing-order table.
- [ ] Cover zero values, `null`, references, DSN, Setter, validation, and `WithOverrides`.
- [ ] Derive errors and troubleshooting from public error types and actual failure conditions.
- [ ] Reduce `vault/README.md` and `cmd/fuda-doc/README.md` to package entries with one example and a site link.
- [ ] Expand `mkdocs.yml` with Integrations, Tooling, and Reference navigation sections.
- [ ] Run `go test ./...` inside `vault` and `cmd/fuda-doc`, then run the strict MkDocs build.
- [ ] Expect both nested modules and the build to pass.
- [ ] Run `make lint`, stage only task files, and commit `docs: add integrations and reference guides`.

### Task 6: Retire legacy pages and validate the release

**Files:**

- Modify: `README.md`
- Modify or delete: `docs/user-guide.md`
- Modify or delete: `docs/tag-spec.md`
- Modify or delete: `docs/config-watcher.md`
- Modify or delete: `docs/custom-resolvers.md`
- Modify or delete: `docs/setter-scanner.md`
- Modify: `examples/README.md`
- Modify: `.gitignore`

**Consumes:** Every final documentation destination from Tasks 2 through 5.

**Produces:** One public documentation system without stale duplicates or broken links.

- [ ] Keep the root README's purpose, installation, compact example, feature summary, license, and site links.
- [ ] Remove duplicated Builder, template, dotenv, DSN, Setter, Scanner, and FAQ chapters from the README.
- [ ] Map every top-level section in the old long-form documents to a new destination before deletion.
- [ ] Replace an old page with a short pointer page when an inbound link cannot move in the same change.
- [ ] Repair links in the root README, examples index, nested READMEs, and site pages.
- [ ] Add `site/` and `.venv-docs/` to `.gitignore`.
- [ ] Keep `requirements-docs.txt`, `mkdocs.yml`, and all source pages tracked.
- [ ] Run `git diff --check`, `make lint`, root tests, nested module tests, all examples, and the strict build.
- [ ] Search for obsolete `docs/user-guide.md`, `docs/tag-spec.md`, and legacy guide links.
- [ ] Expect no whitespace errors, test failures, build warnings, or obsolete links.
- [ ] Ask a repository administrator to choose GitHub Actions under Settings, Pages, after the workflow reaches `main`.
- [ ] Stage only task files and commit `docs: consolidate documentation into the site`.

## Plan self-review

Tasks 2 through 5 cover the approved beginner-first information architecture.
Task 1 provides the Material build and Pages deployment path.
Task 6 removes duplicate content after replacement pages exist.
The executable examples and focused precedence test protect the first tutorial's core claim.
