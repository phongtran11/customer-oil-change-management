# Getting Started

## Prerequisites

| Tool | Minimum Version | Purpose |
|---|---|---|
| Go | 1.23 | Build & run the application |
| Docker + Compose | 24+ | Run PostgreSQL (and optionally the API) |
| `sqlc` | 1.27+ | Regenerate the repository layer after query changes |
| `goose` | 3.x | Run migrations manually (optional — app runs them on startup in development) |

---

## 1. Clone & Configure

```bash
# From the repo root
cd customer-oil-change-management/backend

# Copy the example env file
cp .env.example .env
```

Open `.env` and set at minimum:

```env
JWT_SECRET=replace-with-a-long-random-string
```

> **Security**: Never commit a real `.env` file. Only `.env.example` belongs in version control.

---

## 2. Run with Docker Compose (Recommended)

This starts **PostgreSQL** and the **API** together. The API waits for the database healthcheck before starting.

```bash
docker-compose up --build
```

| Service | URL / Port |
|---|---|
| API | `http://localhost:8080` |
| PostgreSQL | `localhost:5432` |

To stop:

```bash
docker-compose down
# Remove persistent volumes too:
docker-compose down -v
```

---

## 3. Run Locally (Without Docker)

You need a running PostgreSQL instance. Update `DB_URL` in `.env` to point to it.

```bash
# Install / download dependencies
go mod download

# Run the API (migrations are applied automatically on startup when APP_ENV=development)
go run ./cmd/api
```

---

## 4. Running Migrations Manually

Migrations run automatically when the app starts in development mode. To run them manually with the `goose` CLI:

```bash
goose -dir migrations postgres "$DB_URL" up
```

To rollback the last migration:

```bash
goose -dir migrations postgres "$DB_URL" down
```

---

## 5. Regenerating the Repository Layer (sqlc)

After modifying `sql/query.sql` or `sql/schema.sql`, regenerate the Go repository code:

```bash
sqlc generate
```

The generated files are written to `internal/db/sqlc/`.

---

## 6. Building the Binary

```bash
go build -o bin/api ./cmd/api
./bin/api
```

Or build the Docker image directly:

```bash
docker build -t oil-change-api .
```

---

## 7. Verify the Server is Running

```bash
curl -s http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}' | jq
```

Expected response (`201 Created`):

```json
{
  "id": "...",
  "email": "test@example.com"
}
```

---

## 8. Swagger API Documentation

Interactive Swagger documentation is auto-generated from code annotations. Once the server is running, you can access the Swagger UI directly in your browser:

- **Swagger UI URL:** `http://localhost:8080/api/swagger/index.html`

The Swagger UI automatically detects the environment and target URL, allowing you to test endpoints directly from the interface. When testing endpoints under **Protected Routes**, remember to authenticate first:
1. Obtain an access token by calling the `/api/v1/login` endpoint.
2. In Swagger UI, click the **Authorize** button at the top.
3. Enter your token in the format: `Bearer <your_token>`.
