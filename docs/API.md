# Forge Backend API Documentation

This document covers the currently wired API routes in the backend.

## Base URL

The API prefix is configured via environment variables:

- `PORT`
- `API_VERSION`

Compose your base URL as:

`http://localhost:<PORT><API_VERSION>`

Example with common local defaults:

`http://localhost:8080/api/v1`

## Content Type

- Request body content type: `application/json`
- Response content type: `application/json`

## Error Format

All handled errors use this shape:

```json
{
  "error": "message"
}
```

## Endpoints

### 1) POST /signup

Creates a new user.

#### Request Body

```json
{
  "username": "alice",
  "email": "alice@example.com",
  "password": "strongpass123"
}
```

#### Validation Rules

- `username` is required.
- `email` is required and must be valid.
- `password` is required and must be at least 8 characters.

#### Success Response

- Status: `201 Created`

```json
{
  "id": 1,
  "uuid": "8f5d5e53-2b87-4f6c-a6c1-6e0f06d0a5f8",
  "username": "alice",
  "email": "alice@example.com",
  "token": "<jwt-access-token>"
}
```

The token is an HMAC-signed JWT and includes a UUID token ID (`jti`) claim.

#### Common Error Responses

- `400 Bad Request` for malformed JSON or invalid fields.
- `409 Conflict` if a user with the email already exists.
- `500 Internal Server Error` for server-side failures.

### 2) POST /login

Current status: route exists, handler is not implemented yet.

#### Current Response

- Status: `501 Not Implemented`

```json
{
  "error": "login is not implemented yet"
}
```

### 3) POST /logout

Current status: protected route exists, handler is not implemented yet.

#### Authentication

Requires bearer authentication middleware.

#### Current Response

- Status: `501 Not Implemented`

```json
{
  "error": "logout is not implemented yet"
}
```

## CORS / OPTIONS

The server handles `OPTIONS` requests globally and returns `200 OK`.
