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

Authenticates an existing user.

#### Request Body

```json
{
  "email": "alice@example.com",
  "password": "strongpass123"
}
```

#### Success Response

- Status: `200 OK`

```json
{
  "accessToken": "<jwt-access-token>",
  "refreshToken": "<jwt-refresh-token>",
  "uuid": "8f5d5e53-2b87-4f6c-a6c1-6e0f06d0a5f8"
}
```

#### Common Error Responses

- `400 Bad Request` for malformed JSON or invalid email format.
- `401 Unauthorized` for invalid email/password.
- `500 Internal Server Error` for token generation or dependency failures.

### 3) POST /refresh

Refreshes an expired/expiring access token using a valid refresh token.

#### Request Body

```json
{
  "refreshToken": "<jwt-refresh-token>"
}
```

#### Success Response

- Status: `200 OK`

```json
{
  "accessToken": "<jwt-access-token>",
  "refreshToken": "<jwt-refresh-token>",
  "uuid": "8f5d5e53-2b87-4f6c-a6c1-6e0f06d0a5f8"
}
```

#### Common Error Responses

- `400 Bad Request` for malformed JSON or missing `refreshToken`.
- `401 Unauthorized` for invalid, expired, or revoked refresh tokens.

### 4) POST /logout

Invalidates the current access token. If a `refreshToken` is included in the body, that refresh token is revoked as well.

#### Authentication

Requires bearer authentication middleware.

#### Request Body (optional)

```json
{
  "refreshToken": "<jwt-refresh-token>"
}
```

#### Success Response

- Status: `200 OK`

```json
{
  "message": "logged out successfully"
}
```

#### Common Error Responses

- `400 Bad Request` for malformed JSON body.
- `401 Unauthorized` for invalid/expired access token or refresh token.
- `403 Forbidden` if provided refresh token belongs to another user.

## CORS / OPTIONS

The server handles `OPTIONS` requests globally and returns `200 OK`.
