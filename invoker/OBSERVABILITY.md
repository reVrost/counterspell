# Observability & Testing

This directory contains tools for testing and debugging the authentication flow.

## Observability Features

### Backend Logging

The application now includes enhanced logging:

1. **Request Logging** (`internal/middleware/logging.go`):
   - Logs all incoming requests with method, path, query params, remote addr, and user agent
   - Logs response status, duration, and size
   - Logs whether Authorization header is present

2. **Debug Logging** (enabled with `DEBUG=true`):
   - Logs full request headers
   - Logs request body (for POST/PUT)
   - Logs full response body
   - Pretty-prints JSON responses

### Frontend Logging

The auth callback page (`ui/src/routes/auth/callback/+page.svelte`) now logs:

1. **Request details**:
   - URL being called
   - HTTP method
   - Token presence and length

2. **Response details**:
   - HTTP status code and status text
   - Response body (both success and error)
   - Response headers for errors

All logs are prefixed with `=== AUTH DEBUG ===` for easy filtering in browser console.

## Enabling Debug Logging

### Backend

Set the `DEBUG` environment variable before starting the server:

```bash
export DEBUG=true
go run cmd/invoker/main.go
# or
DEBUG=true go run cmd/invoker/main.go
```

### Frontend

Debug logging is always enabled in the callback page. Open browser DevTools Console and filter by `AUTH DEBUG`.

## Testing Tools

### Token Generator

Generate valid JWT tokens for testing:

```bash
# Generate a token with default values
SUPABASE_JWT_SECRET="your-secret-here" go run cmd/token-gen/main.go

# Generate token with custom email/user
SUPABASE_JWT_SECRET="your-secret" TEST_EMAIL="user@example.com" TEST_USER_ID="custom-id" go run cmd/token-gen/main.go
```

Output includes:
- The generated JWT token
- Test curl command to use the token
- Email, user ID, and expiration time

### Auth Endpoint Test Script

Test the `/api/auth/profiles` endpoint:

```bash
# Test with a token
./test-auth.sh <your-jwt-token>

# Test with custom API URL
API_URL=http://localhost:3000 ./test-auth.sh <your-jwt-token>
```

The script:
- Makes a POST request to `/api/auth/profiles`
- Shows status code, headers, and response body
- Uses colored output (green for success, red for errors)
- Validates response (expects 200 for success)

## Manual Testing Flow

1. **Start the server with debug logging**:
   ```bash
   source .envrc
   DEBUG=true go run cmd/invoker/main.go
   ```

2. **Generate a test token**:
   ```bash
   go run cmd/token-gen/main.go
   ```

3. **Test the endpoint**:
   ```bash
   ./test-auth.sh <token-from-step-2>
   ```

4. **Or test via browser**:
   - Open browser DevTools Console
   - Complete auth flow through your app
   - Filter console by `AUTH DEBUG` to see detailed logs

## Common Issues

### 401 Unauthorized - ES256 Token Validation Failed

If logs show `"failed to fetch JWKS: JWKS endpoint returned status 404"`, the JWKS endpoint isn't accessible:

**Diagnose:**
```bash
# Test JWKS endpoint directly
./test-jwks.sh https://your-project.supabase.co
```

**Check JWT signing method in Supabase:**
1. Go to https://supabase.com/dashboard/project/_/settings/api
2. Look for "JWT signing method" (RS256 vs HS256)
3. If HS256, set `SUPABASE_JWT_SECRET` and don't use RS256
4. If RS256, verify `SUPABASE_URL` is correct

**Possible solutions:**

1. **Use HS256 instead** (recommended for testing):
   - Set `SUPABASE_JWT_SECRET` from project settings
   - Enable JWT signing with HS256 in Supabase project settings

2. **Verify project URL:**
   ```bash
   echo $SUPABASE_URL
   # Should match: https://project-ref.supabase.co
   ```

3. **Check JWKS endpoint manually:**
   ```bash
   curl https://your-project.supabase.co/.well-known/jwks.json
   ```

### 401 Unauthorized

Check backend logs for:
- "missing Authorization header" → Token not sent from frontend
- "invalid Authorization header format" → Token doesn't start with "Bearer "
- "invalid token" or "SUPABASE_JWT_SECRET not configured" → Wrong or missing JWT_SECRET

### Check JWT_SECRET

The JWT secret must match between:
- Supabase project settings (`supabase project > settings > API > JWT Secret`)
- Backend environment (`SUPABASE_JWT_SECRET` in .envrc or JWT_SECRET)

Both are now supported by the config loader.

### Enable All Logging

For maximum visibility:

```bash
# Backend
export DEBUG=true
go run cmd/invoker/main.go

# Frontend (browser console)
# Filter by: "AUTH DEBUG"
```

## Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `DEBUG` | Enable detailed request/response logging | No | `false` |
| `SUPABASE_JWT_SECRET` | JWT secret for token validation | Yes | - |
| `SUPABASE_JWT_SECRET` | Alternative JWT secret (fallback) | No | - |
| `API_URL` | API base URL for test script | No | `http://localhost:8079` |
| `TEST_EMAIL` | Email for test token generator | No | `test@example.com` |
| `TEST_USER_ID` | User ID for test token generator | No | `test-user-id-123` |
