# Project Architecture

## Overview
Fuda (札) is a lightweight, struct-tag-first Go configuration library that supports built-in defaults, environment overrides, external secret references, and validation.

## Core Components
1. **Loader / Builder (`fuda.New(...)`)**: The primary entry point that uses the Builder pattern to construct a `Loader`. It takes configurations such as explicit files, dotenv files, environments, custom validation, and custom filesystems (afero).
2. **Tag Processing Engine**: Fuda configures objects based on Go struct tags:
   - `yaml`/`json`: Standard Unmarshaling.
   - `default`: Base fallback values.
   - `env`: Environment variable extraction with optional prefixes.
   - `ref` / `refFrom`: External secret or remote resolution.
   - `dsn`: Dynamic string composition for things like database URLs.
3. **External Resolvers (`RefResolver`)**: An interface for fetching secrets or text from external systems via `ref` and `refFrom` tags. Notable implementations include `fuda/vault` for HashiCorp Vault.
4. **Lifecycle Hooks (`Setter` & `Scanner`)**:
   - `Setter`: Interface for dynamic fallback defaults.
   - `Scanner`: Interface for custom string-to-value parsing.
5. **Watcher (`fuda/watcher`)**: Provides hot-reloading configurations utilizing `fsnotify` underneath to watch for local file changes.
