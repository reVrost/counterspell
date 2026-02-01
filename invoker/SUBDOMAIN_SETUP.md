# Subdomain Setup Guide

## Overview
This allows users to access their own Counterspell instance at `username.counterspell.io`.

## What Was Added

### 1. Middleware (`cmd/app/main.go`)
Extracts the subdomain from the request and adds it to the context.

### 2. Helper Function
`SubdomainFromContext(ctx)` - extracts subdomain from context in handlers.

### 3. Debug Endpoint
`GET /debug/subdomain` - returns the current subdomain for testing.

## Setup Steps

### 1. Update fly.toml
Already configured with services section for custom domains.

### 2. Deploy to Fly.io
```bash
fly deploy
```

### 3. Create SSL Certificates
```bash
fly certs create counterspell.io
fly certs create *.counterspell.io
```

### 4. Configure DNS
Add these records to `counterspell.io`:

| Type | Name | Value |
|------|------|-------|
| CNAME | `*` | `counterspell.fly.dev` |
| CNAME | `@` | `counterspell.fly.dev` (optional, for root domain) |

## Usage Examples

### In Handlers
```go
func (h *Handlers) SomeHandler(w http.ResponseWriter, r *http.Request) {
    subdomain := SubdomainFromContext(r.Context())
    // subdomain will be "john" for john.counterspell.io
    // subdomain will be "" for counterspell.io
}
```

### Testing
```bash
curl https://test.counterspell.io/debug/subdomain
# Returns: {"subdomain":"test"}

curl https://counterspell.io/debug/subdomain
# Returns: {"subdomain":""}
```

## Future Multi-Tenant Implementation

If you want separate data per subdomain:

```go
func (h *Handlers) HandleListTask(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    userID := auth.UserIDFromContext(ctx)
    subdomain := SubdomainFromContext(ctx)

    // Use subdomain as additional namespace for multi-tenancy
    tasks := h.taskService.GetTasks(ctx, userID, subdomain)
    // ...
}
```

## Notes

- Subdomain is extracted from the first part of the hostname
- `www.counterspell.io` is treated as no subdomain (empty string)
- Currently single-user - all subdomains point to same app/data
- Middleware is passive - won't break existing functionality
