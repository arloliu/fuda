# Dotenv Example

Demonstrates loading environment variables from `.env` files using [godotenv](https://github.com/joho/godotenv).

## Features

- **Single file**: `WithDotEnv(".env")`
- **Overlays**: `WithDotEnvFiles([]string{".env", ".env.local"})`
- **Override mode**: `WithDotEnv(".env", fuda.DotEnvOverride())`

## Files

| File              | Purpose                                  |
| ----------------- | ---------------------------------------- |
| `.env`            | Base environment variables               |
| `.env.local`      | Local development overrides (not in git) |
| `.env.production` | Production-specific values               |

## Run

```bash
go run main.go
```

## Key Points

1. Dotenv files are loaded **before** `env` tag processing
2. By default, existing env vars take precedence over dotenv values
3. Use `DotEnvOverride()` to force dotenv values to win
4. Missing files are silently ignored (safe for optional `.env.local`)
