# Install Fuda

## What you will learn

You will add Fuda to an existing Go module and check that the dependency is available.

## Before you start

Install Go and open a terminal in the root of your Go project.
Your project should contain a `go.mod` file.

## Step 1: Add the dependency

Run:

```bash
go get github.com/arloliu/fuda
```

Go adds Fuda to your module dependencies.

## Step 2: Check your module file

Open `go.mod`.
It should include `github.com/arloliu/fuda` in a `require` block or as a direct requirement.

## What happened

Your project can now import `github.com/arloliu/fuda`.
You do not need a separate Fuda command-line tool to load configuration.

## Common mistakes

Run `go get` from the directory that contains your application's `go.mod`.
If Go says it cannot find a module, create one with `go mod init example.com/your-app` before you install Fuda.

## Next step

[Build your first configuration](first-configuration.md) with a YAML file, a default value, and an environment override.
