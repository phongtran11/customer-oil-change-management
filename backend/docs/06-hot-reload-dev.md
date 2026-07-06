# Hot Reload Development with Docker Compose

This project uses [air](https://github.com/air-verse/air) for Go hot reload inside Docker.  
When you save any `.go` or `.sql` file, air detects the change, recompiles, and restarts the binary — all inside the container.

---

## How It Works

| File                          | Purpose                                                                |
| ----------------------------- | ---------------------------------------------------------------------- |
| `Dockerfile.dev`              | Dev image: full Go toolchain + air installed                           |
| `.air.toml`                   | air config: what to watch, how to build, debounce delay                |
| `docker-compose.override.yml` | Swaps the `api` service to use `Dockerfile.dev` and mounts your source |

**Docker Compose auto-merge**: when you run `docker-compose up`, Compose automatically reads both `docker-compose.yml` (postgres + prod api config) **and** `docker-compose.override.yml` (dev overrides), merging them together. You don't need any extra flags.

---

## First-Time Setup

```bash
cd customer-oil-change-management/backend

# Copy env file (only needed once)
cp .env.example .env
# Edit .env — at minimum set JWT_SECRET

# Build the dev image and start everything
docker-compose up --build
```

The first `--build` takes ~1–2 min (downloads Go toolchain + air). Subsequent starts are fast because the Go module cache is stored in a named Docker volume (`go_module_cache`).

---

## Daily Workflow

```bash
# Start all services (postgres + api with air)
docker-compose up

# In another terminal — edit any .go file and save.
# air will print:
#   watching...
#   building...
#   running...
# The API is back up in < 1 second.
```

**You never need to restart docker-compose** — air handles rebuilds automatically.

---

## Stopping

```bash
# Stop all services (keeps DB data)
docker-compose down

# Stop AND wipe the database volume
docker-compose down -v
```

---

## Forcing a Rebuild of the Dev Image

If you change `go.mod` / `go.sum` (add a new dependency), rebuild the dev image to re-download modules:

```bash
docker-compose build api
docker-compose up
```

---

## What air Watches

Configured in [`.air.toml`](../.air.toml):

| Watched       | Extension       |
| ------------- | --------------- |
| `cmd/`        | `.go`           |
| `internal/`   | `.go`           |
| `migrations/` | `.sql`          |
| `sql/`        | `.sql`, `.toml` |

> Changes to `migrations/` trigger a rebuild. Because `main.go` runs `goose.Up()` on startup, migrations are re-applied (new ones only) every time the binary restarts.

---

## Switching Back to Production Mode

The override file is only active in development. To simulate the **production** build (multi-stage, minimal alpine image):

```bash
# Explicitly use only the prod compose file (skips the override)
docker-compose -f docker-compose.yml up --build
```

---

## Troubleshooting

| Problem                              | Fix                                                                                                          |
| ------------------------------------ | ------------------------------------------------------------------------------------------------------------ |
| `air: command not found`             | Rebuild the dev image: `docker-compose build api`                                                            |
| Changes not detected                 | Check that you're editing files **inside** the `backend/` directory (the bind-mounted folder)                |
| Port 8080 already in use             | Change `SERVER_PORT` in `.env`                                                                               |
| `go: module not found`               | Run `go mod tidy` on the host, then `docker-compose build api`                                               |
| DB connection refused on first start | Wait a few seconds — postgres is still starting. `docker-compose up` waits for the healthcheck automatically |
