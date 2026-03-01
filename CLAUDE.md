# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Status

**Archived.** The final release is `v1.0.6`. This project is superseded by the official [aws/rolesanywhere-credential-helper](https://github.com/aws/rolesanywhere-credential-helper), which natively supports writing `~/.aws/credentials` (`update` command), serving credentials over HTTP (`serve` command), and SDK integration (`credential-process` command).

## Build & Test Commands

```bash
# Build the binary
go build -o credential-helper .

# Run tests
go test ./...

# Run a single test
go test ./awsconfig -run TestGetCredentials

# Build Docker image
docker build --build-arg APP_VERSION=v1.0.6 -t credential-helper .
```

No Makefile, linter, or `go.sum` — the project uses only the Go standard library with zero external dependencies.

## Architecture

Kubernetes sidecar container that bridges AWS IAM Roles Anywhere credential delivery to applications expecting filesystem-based AWS credentials (`~/.aws/credentials` and `~/.aws/config`).

**Companion sidecar:** Designed to run alongside [iam-roles-anywhere-sidecar](https://github.com/josh23french/iam-roles-anywhere-sidecar) (also archived), which serves temporary AWS credentials over a local HTTP endpoint.

### Flow

1. `credential-helper.go` (main) — polls `$AWS_CONTAINER_CREDENTIALS_FULL_URI` on a configurable interval via `parseRefreshInterval`
2. `awsconfig/awsconfig.go` — fetches the JSON credential response, converts it to INI format, and atomically writes `~/.aws/credentials` and `~/.aws/config` via temp-file-then-rename
3. Main runs an HTTP health check server at `/healthz` on `$LISTEN_PORT`
4. `getHomeDir` is a package-level `var` (not a named function) to allow test injection of temp directories

### Credential JSON → INI mapping

The sidecar returns JSON with fields `AccessKeyId`, `SecretAccessKey`, `Token`, `Expiration`, and `RoleArn`. The `awsconfig` package converts this to INI `[default]` profile format for the credentials file. The config file only writes the `[default]` section header.

## CI/CD

GitHub Actions workflow (`.github/workflows/docker-publish.yml`):
- Triggers on pushes to `main` and semver tags (`v*.*.*`)
- Builds a multi-stage Docker image (Go build on Alpine 3.18, runtime on Alpine 3.23)
- Pushes to `ghcr.io` and signs with cosign v3.0.5 on non-PR events
- `APP_VERSION` build arg is set from `github.ref_name`
- Platform defaults to `linux/amd64` via `ARG PLATFORM`, overridable at build time

## Environment Variables

| Variable | Default | Notes |
|---|---|---|
| `AWS_CONTAINER_CREDENTIALS_FULL_URI` | `http://localhost:8080/creds` | Credential endpoint URL |
| `AWS_REFRESH_INTERVAL` | `300` (seconds) | Polling interval |
| `LISTEN_PORT` | `:3000` | Health check port (must include colon prefix) |
| `AWS_REGION` / `AWS_DEFAULT_REGION` | `us-east-2` | Set in Dockerfile defaults |
