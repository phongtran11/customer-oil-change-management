# Oil Change Records API

Base URL: `http://localhost:8080`

All endpoints require a valid JWT access token in the `Authorization` header (`Bearer <token>`).  
All request/response bodies are JSON (`Content-Type: application/json`).

Oil change records are **nested under vehicles** — you always address them via `/api/v1/vehicles/{vehicleID}/oil-changes`.

---

## Endpoints

### `POST /api/v1/vehicles/{vehicleID}/oil-changes`

Log a new oil change service for a vehicle.

**Path Parameters**

| Parameter | Type | Description |
|---|---|---|
| `vehicleID` | UUID | The vehicle's unique ID |

**Request Body**

| Field | Type | Required | Rules |
|---|---|:---:|---|
| `service_date` | string (RFC3339) | ✅ | Date of the service |
| `current_mileage` | integer | ✅ | Must be ≥ 0 |
| `next_service_mileage` | integer | ❌ | Recommended mileage for next change |
| `next_service_date` | string (RFC3339) | ❌ | Recommended date for next change |
| `oil_type` | string | ❌ | e.g. `5W30`, `5W40` |
| `oil_filter` | string | ❌ | Oil filter part number or description |
| `next_oil_filter` | string | ❌ | Recommended filter for next change |

**Example Request**

```bash
curl -s -X POST \
  http://localhost:8080/api/v1/vehicles/b4925809-60f5-45f7-ad6e-1f0ca0922ea9/oil-changes \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "service_date": "2026-06-15T00:00:00Z",
    "current_mileage": 158397,
    "next_service_mileage": 166397,
    "oil_type": "5W40"
  }'
```

**Success Response — `201 Created`**

```json
{
  "id": "49b586b2-4eda-436c-9b93-33d8306e18f0",
  "vehicle_id": "b4925809-60f5-45f7-ad6e-1f0ca0922ea9",
  "service_date": "2026-06-15T00:00:00Z",
  "current_mileage": 158397,
  "next_service_mileage": 166397,
  "oil_type": "5W40",
  "created_at": "2026-06-15T07:47:45Z"
}
```

> Nullable fields (`next_service_date`, `oil_filter`, `next_oil_filter`) are omitted from the response when `null`.

**Error Responses**

| Status | Reason |
|---|---|
| `400` | Invalid UUID or malformed JSON |
| `401` | Missing or invalid access token |
| `404` | Vehicle not found |
| `422` | Validation failed (e.g. missing `service_date` or negative mileage) |

---

### `GET /api/v1/vehicles/{vehicleID}/oil-changes`

List all oil change records for a vehicle, ordered by service date descending.

**Path Parameters**

| Parameter | Type | Description |
|---|---|---|
| `vehicleID` | UUID | The vehicle's unique ID |

**Example Request**

```bash
curl -s \
  http://localhost:8080/api/v1/vehicles/b4925809-60f5-45f7-ad6e-1f0ca0922ea9/oil-changes \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Success Response — `200 OK`**

```json
[
  {
    "id": "49b586b2-4eda-436c-9b93-33d8306e18f0",
    "vehicle_id": "b4925809-60f5-45f7-ad6e-1f0ca0922ea9",
    "service_date": "2026-06-15T00:00:00Z",
    "current_mileage": 158397,
    "next_service_mileage": 166397,
    "oil_type": "5W40",
    "created_at": "2026-06-15T07:47:45Z"
  }
]
```

**Error Responses**

| Status | Reason |
|---|---|
| `400` | Invalid UUID format |
| `401` | Missing or invalid access token |
| `404` | Vehicle not found |

---

### `GET /api/v1/vehicles/{vehicleID}/oil-changes/latest`

Retrieve the most recent oil change record for a vehicle.

**Path Parameters**

| Parameter | Type | Description |
|---|---|---|
| `vehicleID` | UUID | The vehicle's unique ID |

**Example Request**

```bash
curl -s \
  http://localhost:8080/api/v1/vehicles/b4925809-60f5-45f7-ad6e-1f0ca0922ea9/oil-changes/latest \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Success Response — `200 OK`**

Returns the single most recent record (same shape as the list items above).

**Error Responses**

| Status | Reason |
|---|---|
| `400` | Invalid UUID format |
| `401` | Missing or invalid access token |
| `404` | Vehicle not found, or vehicle has no records yet |

---

### `GET /api/v1/vehicles/{vehicleID}/oil-changes/{recordID}`

Retrieve a specific oil change record by its UUID.

**Path Parameters**

| Parameter | Type | Description |
|---|---|---|
| `vehicleID` | UUID | The vehicle's unique ID |
| `recordID` | UUID | The oil change record's unique ID |

**Example Request**

```bash
curl -s \
  http://localhost:8080/api/v1/vehicles/b4925809-60f5-45f7-ad6e-1f0ca0922ea9/oil-changes/49b586b2-4eda-436c-9b93-33d8306e18f0 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Success Response — `200 OK`**

Returns the record object.

**Error Responses**

| Status | Reason |
|---|---|
| `400` | Invalid UUID format |
| `401` | Missing or invalid access token |
| `404` | Record not found |

---

### `DELETE /api/v1/vehicles/{vehicleID}/oil-changes/{recordID}`

Remove a specific oil change record.

**Path Parameters**

| Parameter | Type | Description |
|---|---|---|
| `vehicleID` | UUID | The vehicle's unique ID |
| `recordID` | UUID | The oil change record's unique ID |

**Example Request**

```bash
curl -s -X DELETE \
  http://localhost:8080/api/v1/vehicles/b4925809-60f5-45f7-ad6e-1f0ca0922ea9/oil-changes/49b586b2-4eda-436c-9b93-33d8306e18f0 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Success Response — `204 No Content`**

_(empty body)_

**Error Responses**

| Status | Reason |
|---|---|
| `400` | Invalid UUID format |
| `401` | Missing or invalid access token |
| `404` | Record not found |

---

## Error Response Format

All errors follow a consistent shape:

```json
{
  "error": "human-readable error message"
}
```
