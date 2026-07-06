# Configuration

Configuration is loaded by [spf13/viper](https://github.com/spf13/viper) at startup.

**Priority order (highest → lowest):**
1. OS environment variables
2. `.env` file in the working directory
3. Built-in defaults

Copy `.env.example` to `.env` to get started:

```bash
cp .env.example .env
```

---

## All Variables

### Server

| Variable | Default | Required | Description |
|---|---|:---:|---|
| `SERVER_PORT` | `8080` | ❌ | Port the HTTP server listens on |

### Database

| Variable | Default | Required | Description |
|---|---|:---:|---|
| `DB_URL` | _(none)_ | ✅ | Full PostgreSQL connection string |
| `DB_HOST` | `localhost` | ❌ | Host (informational, used by docker-compose) |
| `DB_PORT` | `5432` | ❌ | Port (informational, used by docker-compose) |
| `DB_USER` | `postgres` | ❌ | Username (informational, used by docker-compose) |
| `DB_PASSWORD` | `postgres` | ❌ | Password (informational, used by docker-compose) |
| `DB_NAME` | `oil_change_db` | ❌ | Database name (informational, used by docker-compose) |

> When running locally, only `DB_URL` matters for the Go app. The individual `DB_*` variables are used by Docker Compose to configure the postgres container.

**Connection string format:**

```
postgres://USER:PASSWORD@HOST:PORT/DBNAME?sslmode=disable
```

**Example:**

```env
DB_URL=postgres://postgres:postgres@localhost:5432/oil_change_db?sslmode=disable
```

### Authentication & Tokens

| Variable | Default | Required | Description |
|---|---|:---:|---|
| `JWT_SECRET` | _(none)_ | ✅ | HMAC-SHA256 signing key for JWT access tokens |
| `ACCESS_TOKEN_EXPIRY_MINUTES` | `15` | ❌ | Access token lifetime in minutes |
| `REFRESH_TOKEN_EXPIRY_DAYS` | `7` | ❌ | Refresh token lifetime in days |

> **Security**: `JWT_SECRET` must be a strong random string in production. Generate one with:
> ```bash
> openssl rand -hex 32
> ```

---

## Example `.env.example`

```env
# Server
SERVER_PORT=8080

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=oil_change_db
DB_URL=postgres://postgres:postgres@localhost:5432/oil_change_db?sslmode=disable

# JWT
JWT_SECRET=your-super-secret-jwt-key-change-in-production

# Token Expiry
ACCESS_TOKEN_EXPIRY_MINUTES=15
REFRESH_TOKEN_EXPIRY_DAYS=7
```

---

## Docker Compose Env Vars

When running with Docker Compose, the `api` service receives its config via the `environment:` block in `docker-compose.yml`. The `DB_URL` is automatically constructed from the individual `DB_*` variables pointing to the `postgres` service hostname:

```yaml
environment:
  DB_URL: postgres://${DB_USER}:${DB_PASSWORD}@postgres:5432/${DB_NAME}?sslmode=disable
```

Note that `@postgres` (not `@localhost`) is used because Docker Compose puts both services on the same network.

---

## Production Checklist

- [ ] Set a strong, random `JWT_SECRET` (at least 32 bytes of entropy)
- [ ] Use a strong `DB_PASSWORD`
- [ ] Never commit `.env` to version control
- [ ] Enable SSL for the database connection (`sslmode=require` or `sslmode=verify-full`)
- [ ] Consider shorter `ACCESS_TOKEN_EXPIRY_MINUTES` for higher-security environments
