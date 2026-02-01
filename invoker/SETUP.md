# Counterspell - Complete Deployment Setup Guide

This guide walks you through deploying the Invoker control plane to Fly.io with automatic GitHub Actions deployment and custom domain routing to counterspell.io.

## Prerequisites

- A GitHub account
- A Fly.io account (sign up at https://fly.io)
- Domain: counterspell.io
- Git installed locally
- Go 1.25+ installed locally (optional, for local testing)

---

## Step 1: Initial Local Setup (Optional)

If you want to test locally first:

```bash
# Navigate to project directory
cd /Users/revrost/Code/invoker

# Build the application
go build -o bin/invoker ./cmd/invoker

# Run locally (will serve landing page on http://localhost:8080)
./bin/invoker
```

You can skip this if you want to deploy directly.

---

## Step 2: Install Fly.io CLI

### macOS:
```bash
brew install flyctl
```

### Linux:
```bash
curl -L https://fly.io/install.sh | sh
```

### Windows:
```powershell
iwr https://fly.io/install.ps1 -useb | iex
```

---

## Step 3: Authenticate with Fly.io

```bash
# Login to Fly.io
flyctl auth login
```

This will open your browser for authentication.

---

## Step 4: Create Fly.io App

```bash
# Create the app (this creates a new app on Fly.io)
flyctl apps create counterspell-invoker

# If you want a different name, update fly.toml app field first
```

---

## Step 5: Configure Environment Secrets

Set up your Fly.io app secrets (these will be loaded from GitHub Actions later, but set them now if needed):

```bash
# Basic environment
flyctl secrets set PORT=8080 --app counterspell-invoker
flyctl secrets set APP_VERSION=0.1.0 --app counterspell-invoker
flyctl secrets set ENVIRONMENT=production --app counterspell-invoker

# Add any other required secrets when ready:
# flyctl secrets set SUPABASE_URL=your-value --app counterspell-invoker
# flyctl secrets set SUPABASE_ANON_KEY=your-value --app counterspell-invoker
# flyctl secrets set DATABASE_URL=your-value --app counterspell-invoker
```

---

## Step 6: Test Local Deployment (Optional)

```bash
# Deploy from your local machine to test
flyctl deploy
```

Visit your app: `https://counterspell-invoker.fly.dev`

---

## Step 7: Set Up GitHub Repository

### 7a. Initialize Git (if not already done)

```bash
cd /Users/revrost/Code/invoker
git init
git add .
git commit -m "Initial commit: Add landing page and deployment config"
```

### 7b. Create GitHub Repository

1. Go to https://github.com/new
2. Create a new repository (e.g., `invoker` or `counterspell`)
3. Don't initialize with README (we already have one)
4. Copy the repository URL

### 7c. Push to GitHub

```bash
git branch -M main
git remote add origin https://github.com/YOUR_USERNAME/YOUR_REPO.git
git push -u origin main
```

---

## Step 8: Configure GitHub Secrets for Deployment

### 8a. Get Your Fly.io API Token

```bash
# Generate a new Fly.io token
flyctl tokens create deploy
```

Copy the generated token (starts with `fo1_...`)

### 8b. Add Token to GitHub Secrets

1. Go to your repository on GitHub
2. Navigate to: **Settings** → **Secrets and variables** → **Actions**
3. Click **New repository secret**
4. Name: `FLY_API_TOKEN`
5. Value: Paste your Fly.io token from step 8a
6. Click **Add secret**

---

## Step 9: Enable GitHub Actions

1. Go to your repository on GitHub
2. Navigate to **Actions** tab
3. Click **I understand my workflows, go ahead and enable them**

---

## Step 10: Deploy via GitHub Actions

### 10a. Make a Simple Change to Trigger Deployment

```bash
# Make a small change to trigger deployment
echo "# Deployment setup complete" >> DEPLOYMENT.md
git add DEPLOYMENT.md
git commit -m "Add deployment notes"
git push
```

### 10b. Watch Deployment

1. Go to **Actions** tab in GitHub
2. Click on the "Deploy to Fly.io" workflow
3. Watch the deployment logs

Once complete, your app will be deployed to: `https://counterspell-invoker.fly.dev`

---

## Step 11: Add Custom Domain (counterspell.io)

### 11a. Add Domain to Fly.io

```bash
# Add your domain to Fly.io
flyctl certs add counterspell.io --app counterspell-invoker
```

Fly.io will show you DNS records to add.

### 11b. Configure DNS Records

Go to your DNS provider (where you bought counterspell.io) and add these records:

**For the root domain (@):**

| Type | Name | Value | TTL |
|------|------|-------|-----|
| A    | @    | (provided by Fly.io) | 3600 |

**Or for CNAME (if Fly.io provides one):**

| Type | Name | Value | TTL |
|------|------|-------|-----|
| CNAME| @    | counterspell-invaker.fly.dev | 3600 |

**For wildcard subdomains (.counterspell.io):**

| Type | Name | Value | TTL |
|------|------|-------|-----|
| CNAME| *    | counterspell-invaker.fly.dev | 3600 |

Note: Fly.io will show you the exact IP addresses or CNAME targets to use. The commands in Step 11a will output the exact values.

### 11c. Verify DNS

```bash
# Wait 5-10 minutes for DNS to propagate, then check:
dig counterspell.io
```

### 11d. Verify SSL Certificate

Fly.io will automatically provision SSL. Check status:

```bash
flyctl certs show counterspell.io --app counterspell-invoker
```

The certificate should show as "Ready" within a few minutes.

---

## Step 12: Test Your Deployment

### 12a. Test Landing Page

Visit any of these URLs:
- https://counterspell.io
- https://www.counterspell.io
- https://counterspell-invoker.fly.dev

You should see the "⚡ Counterspell" landing page.

### 12b. Test API Endpoints

```bash
# Health check
curl https://counterspell.io/health

# Readiness check
curl https://counterspell.io/ready
```

---

## Step 13: Set Up Automatic Deployments

Now that GitHub Actions is configured, every push to `main` will automatically deploy to Fly.io:

```bash
# Make any change
git add .
git commit -m "Update something"
git push

# Watch Actions tab for deployment
```

---

## Quick Reference Commands

### Fly.io Management

```bash
# View app logs
flyctl logs --app counterspell-invoker

# View app status
flyctl status --app counterspell-invoker

# Open app in browser
flyctl open --app counterspell-invoker

# Scale up (increase min_machines_running)
flyctl scale count 2 --app counterspell-invoker

# Update environment variables
flyctl secrets set DATABASE_URL="new-value" --app counterspell-invoker

# View all secrets
flyctl secrets list --app counterspell-invoker
```

### GitHub Actions

- **Trigger**: Push to `main` branch
- **View logs**: GitHub → Actions → "Deploy to Fly.io"
- **Manual deploy**: Go to Actions → "Deploy to Fly.io" → "Run workflow"

---

## Troubleshooting

### Issue: GitHub Actions fails with "app not found"
**Solution**: Make sure the app name in `fly.toml` matches your Fly.io app name exactly.

### Issue: DNS not working
**Solution**:
1. Check DNS propagation: `dig counterspell.io`
2. Verify you added the correct records
3. Wait 10-30 minutes for DNS to propagate globally
4. Check Fly.io certificate: `flyctl certs show counterspell.io`

### Issue: Deployment fails
**Solution**:
1. Check GitHub Actions logs for error details
2. Verify `FLY_API_TOKEN` is correct in GitHub secrets
3. Try deploying locally: `flyctl deploy`

### Issue: Landing page shows 404
**Solution**:
1. Verify `static/index.html` exists
2. Check build logs for file embedding issues
3. Test locally first

---

## Security Notes

- ✅ HTTPS is forced by default in fly.toml
- ✅ Fly.io handles SSL certificates automatically
- ✅ Secrets are encrypted in GitHub Actions
- ⚠️ Never commit `.env` files to git
- ⚠️ Rotate Fly.io API tokens periodically

---

## Step 14: Set Up Supabase

### 14a. Create a Supabase Project

1. Go to https://supabase.com and sign up/log in
2. Click **"New Project"**
3. Fill in project details:
   - **Name**: `counterspell` (or your preferred name)
   - **Database Password**: Generate a strong password (save this!)
   - **Region**: Choose closest to your Fly.io region (e.g., `US East` for `iad`)
4. Click **"Create new project"**
5. Wait for the project to be provisioned (2-3 minutes)

### 14b. Get Supabase Credentials

Once your project is ready:

1. Navigate to **Settings** → **API**
2. Copy the following values:
   - **Project URL**: `https://your-project-ref.supabase.co`
   - **anon public** key: Your anonymous/public key
   - **service_role** key: Your service role key (secret!)

⚠️ **Important**: Never commit the `service_role` key to git or expose it to clients!

### 14c. Configure Supabase OAuth Providers

The application currently supports email/password authentication. To add OAuth providers (Google, GitHub, etc.):

1. Navigate to **Authentication** → **Providers** in Supabase dashboard
2. Enable desired providers (Google, GitHub, etc.)
3. For each provider, you'll need to:
   - Create an OAuth app in the provider's developer console
   - Set the callback URL to: `https://counterspell.io/api/auth/callback/{provider}`
   - Copy the Client ID and Client Secret to Supabase

**Google OAuth Setup Example**:
1. Go to https://console.cloud.google.com/apis/credentials
2. Create OAuth 2.0 credentials
3. Authorized redirect URIs: `https://counterspell.io/api/auth/callback/google`
4. Copy Client ID and Client Secret to Supabase Google provider settings

**GitHub OAuth Setup Example**:
1. Go to https://github.com/settings/developers
2. Register a new OAuth App
3. Authorization callback URL: `https://counterspell.io/api/auth/callback/github`
4. Copy Client ID and Client Secret to Supabase GitHub provider settings

### 14d. Configure Redirect URLs in Supabase

1. Navigate to **Authentication** → **URL Configuration**
2. Add the following URLs:

**Site URL**:
```
https://counterspell.io
```

**Redirect URLs**:
```
https://counterspell.io/**
https://counterspell-invoker.fly.dev/**
http://localhost:8080/** (for local development)
```

### 14e. Create Database Schema

1. Navigate to the **SQL Editor** in Supabase dashboard
2. Execute the schema (from `schema.sql` in the project root):
```sql
-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    username VARCHAR(50) UNIQUE NOT NULL,
    tier VARCHAR(50) DEFAULT 'free',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
```

Or use the schema from `internal/db/schema/schema.sql` if it's more comprehensive.

### 14f. Get Database Connection String

1. Navigate to **Settings** → **Database**
2. Scroll to **Connection String** section
3. Select **URI** tab
4. Copy the connection string (it will look like `postgresql://postgres:[YOUR-PASSWORD]@db.[project-ref].supabase.co:5432/postgres`)
5. Replace `[YOUR-PASSWORD]` with the database password you set in step 14a

### 14g. Set Supabase Environment Variables in Fly.io

```bash
# Set Supabase credentials
flyctl secrets set SUPABASE_URL="https://your-project-ref.supabase.co" --app counterspell-invoker
flyctl secrets set SUPABASE_ANON_KEY="your-anon-key-here" --app counterspell-invoker
flyctl secrets set SUPABASE_SERVICE_ROLE_KEY="your-service-role-key-here" --app counterspell-invoker

# Set database connection string
flyctl secrets set DATABASE_URL="postgresql://postgres:[YOUR-PASSWORD]@db.[project-ref].supabase.co:5432/postgres" --app counterspell-invoker

# Set JWT secret (use a strong random string)
flyctl secrets set JWT_SECRET="your-strong-random-jwt-secret-min-32-chars" --app counterspell-invoker
```

⚠️ **Important**: Keep the `SUPABASE_SERVICE_ROLE_KEY` and `SUPABASE_JWT_SECRET` secure. Never share these publicly.

### 14h. Test Supabase Connection

Deploy again to ensure Supabase integration works:

```bash
# Trigger a deployment (you can push a small change or use flyctl deploy)
flyctl deploy

# Check logs to verify Supabase auth is initialized
flyctl logs --app counterspell-invoker
```

You should see a log line like:
```
Supabase auth initialized
```

### 14i. Verify JWKS Endpoint

The application automatically fetches JWKS (JSON Web Key Set) from Supabase to validate JWT tokens. You can verify this works by checking:

```bash
# Test the JWKS endpoint (replace with your project URL)
curl https://your-project-ref.supabase.co/.well-known/jwks.json
```

You should receive a JSON response with public keys.

---

## Step 15: Add Supabase Credentials to GitHub Secrets

To enable local development and ensure secrets are available in GitHub Actions:

1. Go to your repository on GitHub
2. Navigate to: **Settings** → **Secrets and variables** → **Actions**
3. Add the following secrets:

| Secret Name | Value Source |
|-------------|--------------|
| `SUPABASE_URL` | From Supabase Settings → API |
| `SUPABASE_ANON_KEY` | From Supabase Settings → API |
| `SUPABASE_SERVICE_ROLE_KEY` | From Supabase Settings → API |
| `DATABASE_URL` | Your PostgreSQL connection string |
| `SUPABASE_JWT_SECRET` | Generate a strong random string (min 32 chars) |

4. Click **Add secret** for each one

---

## Step 16: Set Up Local Development (Optional)

### 16a. Create `.env` File

```bash
# Copy the example file
cp .env.example .env

# Edit with your actual values
nano .env
```

### 16b. Update `.env` with Your Values

```bash
# Supabase Configuration
SUPABASE_URL=https://your-project-ref.supabase.co
SUPABASE_ANON_KEY=your-actual-anon-key
SUPABASE_SERVICE_ROLE_KEY=your-actual-service-role-key

# Server Configuration
PORT=8080
APP_VERSION=0.1.0
ENVIRONMENT=development

# Database Configuration
DATABASE_URL=postgresql://postgres:[YOUR-PASSWORD]@db.[project-ref].supabase.co:5432/postgres

# Security
JWT_SECRET=your-strong-random-jwt-secret-min-32-chars
```

### 16c. Run Locally

```bash
# Build and run
go build -o bin/invoker ./cmd/invoker
./bin/invoker
```

Visit: http://localhost:8080

---

## Authentication Flow Overview

### How Supabase Auth Works in This Project

1. **Registration/Login**: User signs up/logs in via `/api/auth/register` or `/api/auth/login`
   - Currently implemented as placeholder (email/password flow)
   - Will integrate with Supabase Auth API

2. **Token Validation**: When a user makes authenticated requests:
   - Frontend includes Supabase JWT in `Authorization: Bearer <token>` header
   - Backend validates token using `SupabaseAuth.ValidateToken()` (see `internal/auth/supabase.go:117-141`)
   - Backend extracts user ID and email from validated claims

3. **JWKS Validation**:
   - On startup, backend fetches JWKS from `{SUPABASE_URL}/.well-known/jwks.json`
   - Public key from JWKS is used to verify JWT signatures
   - No secret storage needed - validates using public key only

4. **OAuth Integration** (future):
   - Redirect users to `{SUPABASE_URL}/auth/v1/authorize?provider=google`
   - Supabase handles OAuth flow with provider
   - Callback redirects to your configured URL with session token
   - Frontend stores token for subsequent requests

---

## Next Steps

1. **Add Supabase**: ✅ Set up Supabase and add credentials to Fly.io secrets (completed above)
2. **Add Database**: ✅ Configure DATABASE_URL secret (completed above)
3. **Implement OAuth Endpoints**: Add `/api/auth/callback` endpoint to handle OAuth callbacks
4. **Implement Features**: Build out the VM provisioning and billing features
5. **Monitor**: Set up monitoring and alerts (Fly.io provides built-in metrics)

---

## Architecture Overview

```
GitHub Repo (main branch)
    ↓ push
GitHub Actions (deploy.yml)
    ↓ deploy
Fly.io (counterspell-invoker)
    ↓ serve
counterspell.io
    + /health
    + /ready
    + /api/auth/*
    + /api/vm/*
    + (landing page)
```

---

## Questions?

- Fly.io docs: https://fly.io/docs/
- GitHub Actions docs: https://docs.github.com/en/actions
- Domain: https://counterspell.io

Happy deploying! 🚀
