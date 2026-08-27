# Fuda Documentation Site Design

## Goal

Publish a Material for MkDocs documentation site on GitHub Pages.
The site must help a Go developer add Fuda to a project.
It must explain the first configuration load in five to ten minutes.

## Audience

The primary reader has a Go project and wants a dependable way to load application configuration.
They may know YAML and environment variables, but they do not know Fuda's struct tags or loading model.
The site introduces advanced capabilities after the reader understands the basic configuration path.

## User journey

The reader starts with installation, a runnable struct, a YAML file, and an environment override in one flow.
The next tutorial explains why each value won.
The reader then chooses a focused concept or how-to guide instead of scanning one long user guide.

## Information architecture

```text
Getting started
  Install Fuda
  Your first configuration
  How Fuda chooses a value

Core concepts
  Configuration files
  Defaults
  Environment variables
  Validation
  Data types

How-to guides
  Load dotenv files
  Load secrets with references
  Build a DSN
  Render a config template
  Override values in code
  Use Setter and Scanner

Integrations
  Watch configuration changes
  Use HashiCorp Vault
  Write a custom resolver
  Generate configuration documentation with fuda-doc

Reference
  Struct tags
  Builder options
  Loading behaviour and edge cases
  Errors and troubleshooting
```

The navigation uses task names when a new user needs a result.
Reference pages retain exact API names for readers who arrive through search.

## Page contract

Each learning page takes five to ten minutes to read.
Each page starts with a short statement of the problem it solves.
Each page has numbered steps and runnable or directly adaptable snippets.
It also explains the result and lists a few common mistakes.
Each page ends with one next-step link.

Pages use clear conversational English and address the reader as "you."
The site keeps the Fuda talisman motif on the landing page.
Interior pages use plain technical language and avoid decorative metaphors or emoji.

## First tutorial

`Your first configuration` must show `yaml`, `default`, and `env` tags in one `Config` struct.
It must include a `config.yaml` and an initial `go run` command.
It must also include an `APP_PORT=... go run` command and expected output for both commands.
It must teach the beginner model as `environment variable -> configuration file -> default`.

The tutorial uses the existing `examples/basic` program as the full runnable example.
The page may show reduced snippets, but each snippet must match the executable example's supported API.

## Loading behaviour

The documentation must separate the beginner model from the complete loading pipeline.
The complete pipeline prepares the source before it processes fields.
It loads dotenv files, renders a configured source template, merges programmatic overrides, and decodes the source.
For each field, Fuda applies the following steps in declaration order:

1. Check `env`.
2. Attempt `refFrom` and `ref` while the value is zero.
3. Apply `default` when the document did not supply a value and no environment variable or reference did.
4. Expand `dsn`.
Fuda runs `SetDefaults()` after the fields of a struct and validates after loading finishes.

The reference must explain these edge cases with separate examples:

- A YAML `0`, `false`, or empty string prevents a default from replacing that field.
- YAML `null` is absent for default handling.
- `WithOverrides` changes the source before decoding, so a matching `env` tag can still override it.
- `ref`, `refFrom`, and `dsn` use current field values and can depend on declaration order.
- `dsn` only writes a zero-valued string field.

The implementation must check each claim against source before publishing it.
Read `internal/loader/engine.go` and the tag processors.
The existing processing-order table is an input to the rewrite, not proof that every edge case is accurate.

## Content ownership and migration

The Material site owns long-form user documentation in `docs/`.
`README.md` remains a short repository entry point with installation, a compact first example, and links to the site.
`examples/` remains the canonical source for runnable programs.
Each website example links to its full program under `examples/`.

The existing `docs/user-guide.md` is split into the new learning pages.
`docs/tag-spec.md` becomes the tag reference.
The watcher, resolver, Setter and Scanner documents become focused integration or guide pages.
The root README becomes a concise repository entry point that links to the site.
`vault/README.md` and `cmd/fuda-doc/README.md` become concise package entry points that link to their site guides.

Delete a legacy long-form page after every unique topic has a destination page.
Update every incoming repository link before deletion.
Do not retain duplicate long-form explanations in both locations.

## Repository layout

```text
mkdocs.yml
requirements-docs.txt

docs/
  index.md
  getting-started/
  concepts/
  guides/
  integrations/
  tooling/
  reference/
  assets/images/
  stylesheets/

.github/workflows/docs.yml
examples/
.planning/
  specs/
  plans/
```

`docs/` is the MkDocs source directory and contains only public site content and assets.
`.planning/specs/` and `.planning/plans/` stay outside `docs/`.
The published site does not index internal planning records.

## Site behaviour

The site uses Material for MkDocs with search, code-copy buttons, and syntax highlighting.
It provides light and dark color schemes, repository links, and a small Fuda stylesheet.
The site uses the existing logo after moving it to `docs/assets/images/`.
The first release does not add blogs, analytics, comments, versioned documentation, or template overrides.

## Build and release

`requirements-docs.txt` pins the Python packages for a reproducible documentation build.
The documentation workflow builds the site in strict mode for pull requests and pushes to `main`.
Only a `main` push deploys the Pages artifact.
The workflow uses GitHub's Pages artifact deployment actions and the `github-pages` deployment environment.

An administrator must set the repository Pages source to GitHub Actions before the first deployment.
The deployment never publishes secret values, local `.env` files, generated output, or example secret files.

## Validation

The documentation build runs `python -m mkdocs build --strict` after installing the pinned requirements.
The build must pass in pull requests before the documentation workflow can deploy from `main`.
The implementation adds a focused Go regression test for the first tutorial's precedence example.
The test covers environment, YAML, and default values.
The implementation runs `make lint`, the focused test, relevant example programs, and `go test ./...`.
It also runs the strict MkDocs build before the final commit.
