# Changelog

All notable changes to this project are documented in this file.
Each version corresponds to an annotated git tag and a
[GitHub release](https://github.com/arloliu/fuda/releases).

## [1.7.1] - 2026-08-27

### Fixed

- `file://relative/path` and `file://./relative/path` now resolve relative to the working directory
  instead of dropping the first path component as a URI authority.
  Bare relative `ref:"..."` tags are normalized to `file://` URIs and are affected.
  ([#1](https://github.com/arloliu/fuda/pull/1), thanks @xdavidwu)
- `file://` prefix matching is case-insensitive.
- Percent-encoded relative paths are decoded.
- `file://host/path` authorities remain unsupported and are read as a relative path.

### Changed

- Documentation consolidated into a Material for MkDocs site.

### Migration notes

- A config that relied on `file://relative/path` reading `/path` should use `file:///path` instead.

## [1.7.0] - 2026-08-25

### Changed

- Validation errors render common go-playground tags as plain statements,
  e.g. `discovery.pprof.port: must be at most 65535`.
- A `default` tag no longer overwrites a value the config file explicitly set to zero (`0`, `false`, `""`);
  a default applies only when nothing supplied the field.
  Writing `null` resets a field to its default.

### Migration notes

- Tests or tooling matching validation error strings must be updated;
  use `errors.As` to retrieve `validator.ValidationErrors` for structured access instead of parsing messages.
- If a config relied on an explicit zero being replaced by the default,
  remove the key or set it to `null` to keep getting the default.

## [1.6.1] - 2026-02-23

### Changed

- Added architecture rule for agents (`.agent/rules/800-architecture.md`).

## [1.6.0] - 2026-02-01

### Changed

- **Breaking:** `SetDefaults` no longer performs validation by default.
- `SetDefaults`, `MustSetDefaults`, and `Validate` accept options.

### Added

- `WithValidation` and `WithValidator` options.

## [1.5.0] - 2026-01-20

### Added

- Documentation for the `fuda.Duration` type and its accessor methods.

## [1.4.1] - 2026-01-06

### Added

- Documentation and examples for `[]byte` support.

## [1.4.0] - 2025-12-29

### Added

- `WithOverrides(map[string]any)` builder method for programmatic config overrides.
  Supports dot notation for nested keys and works with an empty source.
  Priority: env > overrides > config file > ref/refFrom > default > dsn.

## [1.3.1] - 2025-12-28

### Added

- `env://` scheme support for the `ref` tag.

## [1.3.0] - 2025-12-28

### Added

- afero filesystem abstraction: `DefaultFs`, `SetDefaultFs`, `ResetDefaultFs`,
  and `WithFilesystem()` on `fuda.Builder` and `watcher.Builder`.

### Fixed

- Validator is propagated during watcher hot-reload.

## [1.2.2] - 2025-12-28

### Fixed

- Makefile coverage targets restored; `vet` target includes the `vault` module.

## [1.2.1] - 2025-12-28

### Added

- Duration parsing supports a `d` (day) suffix, including fractional, negative, and combined units (`1d12h`).

## [1.2.0] - 2025-12-28

### Changed

- Agent rules refactored into modular files under `.agent/rules/`.

## [1.1.0] - 2025-12-27

### Changed

- Updated logo.

## [1.0.0] - 2025-12-26

Initial release.

[1.7.1]: https://github.com/arloliu/fuda/compare/v1.7.0...v1.7.1
[1.7.0]: https://github.com/arloliu/fuda/compare/v1.6.1...v1.7.0
[1.6.1]: https://github.com/arloliu/fuda/compare/v1.6.0...v1.6.1
[1.6.0]: https://github.com/arloliu/fuda/compare/v1.5.0...v1.6.0
[1.5.0]: https://github.com/arloliu/fuda/compare/v1.4.1...v1.5.0
[1.4.1]: https://github.com/arloliu/fuda/compare/v1.4.0...v1.4.1
[1.4.0]: https://github.com/arloliu/fuda/compare/v1.3.1...v1.4.0
[1.3.1]: https://github.com/arloliu/fuda/compare/v1.3.0...v1.3.1
[1.3.0]: https://github.com/arloliu/fuda/compare/v1.2.2...v1.3.0
[1.2.2]: https://github.com/arloliu/fuda/compare/v1.2.1...v1.2.2
[1.2.1]: https://github.com/arloliu/fuda/compare/v1.2.0...v1.2.1
[1.2.0]: https://github.com/arloliu/fuda/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/arloliu/fuda/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/arloliu/fuda/releases/tag/v1.0.0
