# Go Blog

A small blog with a public feed and password-protected admin panel. Built with Go, Gin, GORM, PostgreSQL, and server-rendered HTML templates.

## Requirements

- Go 1.25+ (local development)
- PostgreSQL
- Docker and Docker Compose (recommended for deployment)
- Node.js 20+ (Playwright E2E tests only)

## Configuration

Copy the example environment file and edit the values:

```bash
cp .env.example .env
```

| Variable | Description |
|----------|-------------|
| `GO_BLOG_HTTP_PORT` | HTTP listen port (default: `8083`) |
| `GIN_MODE` | Gin mode: `release` (default) or `debug` |
| `GO_BLOG_SESSION_SECRET` | Session signing key, at least 32 characters |
| `GO_BLOG_SESSION_SECURE` | Set to `1` or `true` when serving over HTTPS |
| `GO_BLOG_DATABASE_HOST` | PostgreSQL host |
| `GO_BLOG_DATABASE_PORT` | PostgreSQL port (default: `5432`) |
| `GO_BLOG_DATABASE_NAME` | Database name |
| `GO_BLOG_DATABASE_USER` | Database user |
| `GO_BLOG_DATABASE_PASSWORD` | Database password (required) |

Migrations are applied manually with `./blog migrate`. Concurrent runs are serialized with a PostgreSQL advisory lock.

## Local development

```bash
# Start PostgreSQL and point .env at it, then:
go run . migrate
go run . server
```

Create an admin user:

```bash
go run . users-create
```

Or seed a default `admin` / `admin` user (development only):

```bash
go run . users-seed
```

Other CLI commands:

```bash
go run . users-show
go run . posts-show
go run . posts-seed 10
```

## Deployment with Docker

The production `docker-compose.yml` expects:

1. A running PostgreSQL instance (default host: `postgresql`)
2. An external Docker network named `infra`

For a shared PostgreSQL 16 setup on a VPS (one instance, multiple Docker apps), see [docs/example-postgresql-docker-compose/README.md](docs/example-postgresql-docker-compose/README.md).

Create the network once if it does not exist (the shared Postgres compose also creates it):

```bash
docker network create infra
```

Ensure PostgreSQL is reachable from that network and matches the values in `.env`.

Build and start:

```bash
docker compose up -d --build
docker compose exec go-blog ./blog migrate
```

Create the first admin user inside the container:

```bash
docker compose exec go-blog ./blog users-create
```

The container runs as a non-root `appuser`, listens on port `8083`, and uses `GIN_MODE=release` by default.

### Production checklist

- Set strong values for `GO_BLOG_SESSION_SECRET` and `GO_BLOG_DATABASE_PASSWORD`
- Create an admin with `./blog users-create`; avoid `users-seed` in production
- Set `GO_BLOG_SESSION_SECURE=1` when TLS terminates at a reverse proxy
- Put HTTPS in front of the app (nginx, Caddy, Traefik, etc.)
- Use a managed PostgreSQL instance, or the shared VPS stack in [docs/example-postgresql-docker-compose](docs/example-postgresql-docker-compose/README.md), and back it up regularly

### Health check

Docker Compose probes `GET /health` (returns `{"status":"ok"}` when the database is reachable). The app also serves `GET /robots.txt` (disallows all crawlers by default).

## Tests

E2E tests use Playwright and a disposable Docker Compose stack:

```bash
npm ci
npx playwright install --with-deps chromium
npm test
```

To run tests against an already running server:

```bash
SKIP_DOCKER_SETUP=1 GO_BLOG_HTTP_PORT=8083 npm test
```

CI runs `gofmt`, `golangci-lint`, `go vet`, `go build`, Compose validation, a production Compose smoke test (shared Postgres from `docs/example-postgresql-docker-compose` + the app on the `infra` network), and the Playwright suite on every push and pull request (see `.github/workflows/ci.yml`). A separate workflow exercises the PostgreSQL VPS bootstrap script (see `.github/workflows/vps-setup.yml`). A separate advisory workflow runs `govulncheck` and Trivy on the image and filesystem (see `.github/workflows/security.yml`); findings are printed in the job log / step summary and do not fail CI. Dependabot opens weekly PRs for Go, Docker, npm, and GitHub Actions updates (see `.github/workflows/dependabot.yml`).

## Project layout

```
main.go              HTTP server entrypoint
cli/                 CLI commands (users, posts)
handlers/            Route handlers
middleware/          Auth middleware
models/              GORM models
database/            DB connection and migrations
migrations/          SQL migrations (goose)
templates/           HTML templates
tests/               Playwright E2E tests
docs/                Deployment examples (shared PostgreSQL)
```
