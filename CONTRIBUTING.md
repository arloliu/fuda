# Contributing to Fuda

Thanks for taking the time to contribute.

## Before you start

For anything beyond a small fix, open an issue first to discuss the change.
This avoids wasted work on a pull request that doesn't fit the project's direction.

## Development setup

```bash
git clone https://github.com/arloliu/fuda.git
cd fuda
go build ./...
```

Requires Go 1.25 or later.

## Making a change

1. Create a branch named `feat/`, `fix/`, `docs/`, `chore/`, or `test/` followed by a short description,
   e.g. `fix/handle-empty-env-values`.
2. Write tests for new behavior.
   Bug fixes should include a regression test.
3. Run `make lint` and fix all issues.
4. Run `make test` and make sure everything passes.
5. Update documentation under `docs/` if the change affects the public API.

## Commit messages

Use the [Conventional Commits](https://www.conventionalcommits.org/) format, present tense,
with a first line under 50 characters:

```
fix: handle empty env values
feat: add dotenv support
```

## Pull requests

- Keep pull requests focused on a single change.
- Fill in the pull request template.
- CI (lint, vet, test) must pass before a maintainer reviews the change.

## Code review

Reviewers check for:

- Correctness
- Performance (no unnecessary allocations)
- Test coverage for new code
- Updated documentation

## Reporting bugs

Open an issue using the bug report template.
Include a minimal reproduction — a config snippet and struct definition are usually enough.

## License

By contributing, you agree that your contributions will be licensed under the
[Apache License 2.0](LICENSE).
