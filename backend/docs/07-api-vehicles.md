# Vehicles API

Base URL: `http://localhost:8080`

All endpoints require a valid JWT access token in the `Authorization` header (`Bearer <token>`).  
All request/response bodies are JSON (`Content-Type: application/json`).

---

## Endpoints

### `GET /api/v1/vehicles`

List all vehicles, ordered by creation date descending.

**Example Request**

```bash
curl -s http://localhost:8080/api/v1/vehicles \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Success Response — `200 OK`**

```json
[
  {
    "id": "1b0a7a27-8b3b-475c-ac72-2ebca1a78d0a",
    "license_plate": "72A42914",
    "owner_name": "Nguyễn Văn A",
    "phone_number": "0981811837",
    "make": "Toyota",
    "model": "Vios",
    "year": 2022,
    "created_at": "2026-06-29T07:45:51Z"
  }
]
```

> Fields `make`, `model`, and `year` are omitted when `null`.

**Error Responses**

| Status | Reason |
|---|---|
| `401` | Missing or invalid access token |

---

### `POST /api/v1/vehicles`

Register a new vehicle.

**Request Body**

| Field | Type | Required | Rules |
|---|---|:---:|---|
| `license_plate` | string | ✅ | max 50 chars |
| `owner_name` | string | ✅ | max 255 chars |
| `phone_number` | string | ✅ | max 20 chars |
| `make` | string | ❌ | max 100 chars |
| `model` | string | ❌ | max 100 chars |
| `year` | integer | ❌ | — |

**Example Request**

```bash
curl -s -X POST http://localhost:8080/api/v1/vehicles \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "license_plate": "72A42914",
    "owner_name": "Nguyễn Văn A",
    "phone_number": "0981811837",
    "make": "Toyota",
    "model": "Vios",
    "year": 2022
  }'
```

**Success Response — `201 Created`**

```json
{
  "id": "1b0a7a27-8b3b-475c-ac72-2ebca1a78d0a",
  "license_plate": "72A42914",
  "owner_name": "Nguyễn Văn A",
  "phone_number": "0981811837",
  "make": "Toyota",
  "model": "Vios",
  "year": 2022,
  "created_at": "2026-06-29T07:45:51Z"
}
```

**Error Responses**

| Status | Reason |
|---|---|
| `400` | Malformed JSON body |
| `401` | Missing or invalid access token |
| `409` | License plate already registered |
| `422` | Validation failed (e.g. missing required field) |

---

### `GET /api/v1/vehicles/{vehicleID}`

Retrieve a vehicle by its UUID.

**Path Parameters**

| Parameter | Type | Description |
|---|---|---|
| `vehicleID` | UUID | The vehicle's unique ID |

**Example Request**

```bash
curl -s http://localhost:8080/api/v1/vehicles/1b0a7a27-8b3b-475c-ac72-2ebca1a78d0a \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Success Response — `200 OK`**

```json
{
  "id": "1b0a7a27-8b3b-475c-ac72-2ebca1a78d0a",
  "license_plate": "72A42914",
  "owner_name": "Nguyễn Văn A",
  "phone_number": "0981811837",
  "created_at": "2026-06-29T07:45:51Z"
}
```

**Error Responses**

| Status | Reason |
|---|---|
| `400` | Invalid UUID format |
| `401` | Missing or invalid access token |
| `404` | Vehicle not found |

---

### `PUT /api/v1/vehicles/{vehicleID}`

Update a vehicle's details.

**Path Parameters**

| Parameter | Type | Description |
|---|---|---|
| `vehicleID` | UUID | The vehicle's unique ID |

**Request Body** — same fields as `POST /api/v1/vehicles`.

**Example Request**

```bash
curl -s -X PUT http://localhost:8080/api/v1/vehicles/1b0a7a27-8b3b-475c-ac72-2ebca1a78d0a \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "license_plate": "72A42914",
    "owner_name": "Nguyễn Văn B",
    "phone_number": "0912345678"
  }'
```

**Success Response — `200 OK`**

Returns the updated vehicle object.

**Error Responses**

| Status | Reason |
|---|---|
| `400` | Invalid UUID or malformed JSON |
| `401` | Missing or invalid access token |
| `404` | Vehicle not found |
| `409` | License plate already taken by another vehicle |
| `422` | Validation failed |

---

### `DELETE /api/v1/vehicles/{vehicleID}`

Delete a vehicle and all its associated oil change records.

**Path Parameters**

| Parameter | Type | Description |
|---|---|---|
| `vehicleID` | UUID | The vehicle's unique ID |

**Example Request**

```bash
curl -s -X DELETE http://localhost:8080/api/v1/vehicles/1b0a7a27-8b3b-475c-ac72-2ebca1a78d0a \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Success Response — `204 No Content`**

_(empty body)_

**Error Responses**

| Status | Reason |
|---|---|
| `400` | Invalid UUID format |
| `401` | Missing or invalid access token |
| `404` | Vehicle not found |

---

## Error Response Format

All errors follow a consistent shape:

```json
{
  "error": "human-readable error message"
}
```
