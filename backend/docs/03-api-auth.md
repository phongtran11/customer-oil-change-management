# Authentication API

Base URL: `http://localhost:8080`

All request and response bodies are JSON (`Content-Type: application/json`).

---

## Endpoints

### `POST /api/v1/register`

Create a new user account.

**Request Body**

| Field | Type | Required | Rules |
|---|---|:---:|---|
| `email` | string | ✅ | Valid email format, max 255 chars |
| `password` | string | ✅ | Min 8, max 128 characters |

**Example Request**

```bash
curl -s -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "password": "securePass123"
  }'
```

**Success Response — `201 Created`**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "alice@example.com"
}
```

**Error Responses**

| Status | Reason |
|---|---|
| `400` | Malformed JSON body |
| `409` | Email already registered |
| `422` | Validation failed (e.g. invalid email, short password) |

---

### `POST /api/v1/login`

Authenticate and obtain an access token + refresh token.

**Request Body**

| Field | Type | Required |
|---|---|:---:|
| `email` | string | ✅ |
| `password` | string | ✅ |

**Example Request**

```bash
curl -s -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "password": "securePass123"
  }'
```

**Success Response — `200 OK`**

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "a3f7c2e1d4b9...",
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Token Details**

| Token | Storage | Lifetime |
|---|---|---|
| `access_token` | Memory / Authorization header | 15 minutes (configurable) |
| `refresh_token` | Secure HttpOnly cookie or secure storage | 7 days (configurable) |

**Error Responses**

| Status | Reason |
|---|---|
| `401` | Invalid email or password |
| `422` | Missing required fields |

---

### `POST /api/v1/refresh`

Exchange a valid refresh token for a new access token. The old refresh token is **revoked** and a new one is issued (token rotation).

**Request Body**

| Field | Type | Required |
|---|---|:---:|
| `refresh_token` | string | ✅ |

**Example Request**

```bash
curl -s -X POST http://localhost:8080/api/v1/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "a3f7c2e1d4b9..."
  }'
```

**Success Response — `200 OK`**

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "b8e1a4f2c6d3..."
}
```

> ⚠️ **Token Rotation**: The old `refresh_token` is immediately revoked. Store and use only the new token returned in the response.

**Error Responses**

| Status | Reason |
|---|---|
| `401` | Token not found, revoked, or expired |
| `422` | Missing `refresh_token` field |

---

### `POST /api/v1/logout`

Revoke a refresh token. Requires a valid JWT access token.

**Headers**

| Header | Value |
|---|---|
| `Authorization` | `Bearer <access_token>` |

**Request Body**

| Field | Type | Required |
|---|---|:---:|
| `refresh_token` | string | ✅ |

**Example Request**

```bash
curl -s -X POST http://localhost:8080/api/v1/logout \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "refresh_token": "b8e1a4f2c6d3..."
  }'
```

**Success Response — `204 No Content`**

_(empty body)_

**Error Responses**

| Status | Reason |
|---|---|
| `401` | Missing / invalid / expired access token |
| `401` | Refresh token not found or does not belong to the user |
| `422` | Missing `refresh_token` field |

---

## Authentication Flow Diagram

```
Client                          Server
  │                               │
  │── POST /api/v1/register ─────>│ Hash password, store user
  │<─ 201 { id, email } ──────────│
  │                               │
  │── POST /api/v1/login ────────>│ Verify password, generate tokens
  │<─ 200 { access, refresh } ────│ Store refresh token in sessions
  │                               │
  │── GET /api/v1/protected ─────>│
  │   Authorization: Bearer ...   │ Validate JWT
  │<─ 200 { data } ───────────────│
  │                               │
  │   (access token expires)      │
  │                               │
  │── POST /api/v1/refresh ──────>│ Validate refresh token
  │<─ 200 { new_access, new_ref } │ Revoke old, issue new tokens
  │                               │
  │── POST /api/v1/logout ───────>│ Revoke refresh token
  │<─ 204 ────────────────────────│
```

---

## Error Response Format

All errors follow a consistent shape:

```json
{
  "error": "human-readable error message"
}
```
