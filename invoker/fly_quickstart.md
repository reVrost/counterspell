# Fly.io Quickstart Guide for Counterspell

## What is Fly.io?

Fly.io is a platform for running applications close to users globally. It provides:
- **Firecracker microVMs** - Isolated, lightweight virtual machines
- **Anycast networking** - Routes traffic to the nearest region automatically
- **Built-in auto-sleep** - Machines stop when idle to save costs
- **Volumes** - Persistent storage attached to machines
- **Programmatic API** - Create/stop/monitor machines via API

---

## Fly.io Primitives

### 1. Apps vs Machines

**Apps** (Recommended for Counterspell):
- Collection of machines running the same code
- Auto-scaling, rolling deployments, health checks
- Good for long-running services

**Machines** (Best for Counterspell's per-user VMs):
- Individual Firecracker microVMs
- Independent lifecycle (start/stop each one)
- Perfect for one VM per user
- Can attach persistent volumes

**Decision**: Use **Machines** for Counterspell because:
- Each user gets their own isolated environment
- Independent sleep/wake per user
- Cost-efficient (pay per running machine)
- Easier to manage per-user quotas

---

### 2. Machine Types

#### shared-cpu-1x (Best for Counterspell MVP)
- **Cost**: ~$5.70-7.12/month per 512MB RAM (varies by region)
- **vCPU**: 1 shared CPU (burstable)
- **RAM**: 512MB, 1GB, 2GB options
- **Cold Start**: ~5-10 seconds
- **Best For**: Web apps, API servers, light workloads
- **Auto-Sleep**: ✅ Built-in (after configurable inactivity)

#### dedicated-cpu-1x
- **Cost**: ~$23/month per 1 vCPU
- **vCPU**: 1 dedicated CPU
- **RAM**: 2GB, 4GB, 8GB options
- **Cold Start**: ~5-10 seconds
- **Best For**: Heavy workloads, AI inference, consistent performance

#### dedicated-cpu-2x+
- **Cost**: ~$46+/month per 2 vCPUs
- **vCPU**: 2-16 dedicated CPUs
- **RAM**: 4GB-64GB
- **Best For**: Large workloads, database servers

**Recommendation for Counterspell**: **shared-cpu-1x with 1GB RAM**
- Cost-effective for AI agent workloads (~$7-13/month per user)
- Auto-sleep saves ~80% when idle
- Cold starts fast enough for user experience
- Scale up to dedicated-cpu-1x later for power users

---

### 3. Volumes (Persistent Storage)

Fly.io Volumes provide persistent storage attached to machines:

```bash
# Create a volume
flyctl volumes create counterspell-data --region iad --size 1

# Attach to a machine
fly machines run --vm-size shared-cpu-1x --volume counterspell-data:/data
```

**For Counterspell**:
- Attach 1GB volume per user (`counterspell-{user_id}`)
- Store: SQLite databases, git worktrees, agent state
- Data persists across machine restarts
- Cost: ~$0.16/GB/month (negligible)

---

### 4. Auto-Sleep Configuration

Machines can automatically sleep when idle to save costs:

```toml
# fly.toml
[experimental]
auto_stop_machines = true
auto_start_machines = true
min_machines_running = 0
```

**For Counterspell**:
- Enable auto-sleep after 30 minutes of inactivity
- Machines auto-wake on incoming HTTP requests
- **Cost Savings**: ~80% for typical usage patterns

---

### 5. Regions

Fly.io has regions worldwide:
- **iad** (Virginia, USA) - Default, lowest cost
- **sjc** (San Jose, USA)
- **lax** (Los Angeles, USA)
- **dfw** (Dallas, USA)
- **cdg** (Paris, France)
- **ams** (Amsterdam, Netherlands)
- **fra** (Frankfurt, Germany)

**For Counterspell**:
- Default to **iad** (lowest cost)
- Option to let users choose region later
- Fly.io's anycast routing handles nearest routing automatically

---

## Fly.io API for Programmatic Control

Fly.io provides a REST API for managing machines programmatically.

### Authentication
```bash
# Get API token from flyctl
flyctl auth token

# Set as environment variable
export FLY_API_TOKEN=your_token_here
```

### Create a Machine

```bash
POST /v1/apps/{app_name}/machines

{
  "name": "counterspell-{user_id}",
  "region": "iad",
  "config": {
    "image": "registry.fly.io/counterspell:latest",
    "vm": {
      "size": "shared-cpu-1x",
      "memory_mb": 1024
    },
    "env": {
      "USER_ID": "{user_id}",
      "SUBDOMAIN": "{subdomain}"
    },
    "services": [{
      "protocol": "tcp",
      "internal_port": 8080,
      "ports": [{ "port": 80, "handlers": ["http"] }]
    }]
  },
  "mounts": [{
    "volume": "vol_{user_id}",
    "path": "/data"
  }]
}
```

**Response**:
```json
{
  "id": "abc123def456",
  "name": "counterspell-alice",
  "state": "starting",
  "region": "iad",
  "public_ip": ["203.0.113.1"],
  "config": { ... }
}
```

### Get Machine Status

```bash
GET /v1/apps/{app_name}/machines/{machine_id}

Response:
{
  "id": "abc123def456",
  "state": "running",
  "region": "iad",
  "public_ip": ["203.0.113.1"],
  "created_at": "2024-01-25T12:00:00Z",
  "updated_at": "2024-01-25T12:05:00Z"
}
```

**States**: `starting`, `running`, `stopped`, `stopping`, `error`

### Stop a Machine

```bash
POST /v1/apps/{app_name}/machines/{machine_id}/stop

Response:
{
  "id": "abc123def456",
  "state": "stopping"
}
```

### Start a Stopped Machine

```bash
POST /v1/apps/{app_name}/machines/{machine_id}/start

Response:
{
  "id": "abc123def456",
  "state": "starting"
}
```

---

## Cost Breakdown for Counterspell

### Per User (1GB RAM, 1GB Volume, iad Region)

| Resource | Cost/Month | Notes |
|----------|------------|-------|
| **shared-cpu-1x (1GB RAM)** | $5.70-7.12 | Varies by region |
| **1GB Volume** | ~$0.16 | Persistent storage |
| **Bandwidth** | Included | 100GB free/month |
| **Total (Always-On)** | **~$6-7/month** | 24/7 uptime |
| **Total (Auto-Sleep)** | **~$1.20-1.40/month** | 80% savings (typical usage) |

### 100 Active Users (20% concurrent)

| Scenario | Monthly Cost | Notes |
|----------|--------------|-------|
| **All Always-On** | ~$600-700/month | No auto-sleep |
| **Auto-Sleep (Typical)** | ~$120-140/month | 80% savings |
| **Power Users (50% Always-On)** | ~$360-400/month | Mixed usage |

---

## Recommended Setup for Counterspell

### Step 1: Create Fly.io App

```bash
# Login to Fly.io
flyctl auth signup

# Create app
flyctl apps create counterspell-data-plane --org personal
```

### Step 2: Deploy Base Image

```bash
# Build and deploy (will be template for user machines)
flyctl deploy --build-only --remote-only
# This creates Docker image at registry.fly.io/counterspell-data-plane:latest
```

### Step 3: Configure Auto-Sleep

```toml
# fly.toml
[experimental]
auto_stop_machines = true  # Stop machines after inactivity
auto_start_machines = true  # Auto-start on requests
min_machines_running = 0    # Don't keep any machines always-on
```

### Step 4: Create Volumes for Users

```bash
# Create volume for a user
flyctl volumes create counterspell-data-{user_id} \
  --region iad \
  --size 1 \
  --app counterspell-data-plane
```

### Step 5: Launch User Machines (Via API)

```bash
# Use Fly.io API to create machine per user
# See API examples above
```

---

## Firecracker MicroVMs vs Regular VMs

### Firecracker (Fly.io Uses This)
- **Startup Time**: 5-10 seconds
- **Isolation**: Hardware-level isolation (QEMU-based)
- **Cost**: Lower due to high density on host
- **Security**: Strong isolation between tenants
- **Perfect for**: Web apps, APIs, short-lived tasks

### Regular VMs (EC2, DigitalOcean Droplets)
- **Startup Time**: 30-90 seconds
- **Isolation**: OS-level or hardware
- **Cost**: Higher per instance
- **Security**: Varies by provider
- **Perfect for**: Long-running databases, heavy workloads

**Fly.io's Advantage**: Firecracker microVMs + auto-sleep = Fast cold starts + cost savings

---

## Best Practices for Counterspell

### 1. Machine Naming
- Use descriptive names: `counterspell-{username}`
- Easy to identify in Fly.io dashboard
- Consistent with subdomains

### 2. Health Checks
- Implement `/health` endpoint on data plane
- Fly.io auto-restarts unhealthy machines
- Control plane can also monitor via API

### 3. Volume Management
- One volume per user (attach to machine)
- Persist SQLite databases, git worktrees
- Volume survives machine recreation

### 4. Environment Variables
- Pass `USER_ID` to data plane
- Pass `SUBDOMAIN` for routing info
- Pass control plane API URL for health callbacks

### 5. Auto-Sleep Tuning
- Default: 30 minutes inactivity
- Let users adjust via settings later
- Power users can disable auto-sleep

### 6. Error Handling
- Retry machine creation on API errors
- Log all Fly.io API calls for debugging
- Alert on machine failures

---

## Monitoring Fly.io Machines

### Via Flyctl
```bash
# List all machines
flyctl machines list --app counterspell-data-plane

# View machine logs
flyctl logs --app counterspell-data-plane --machine abc123def456

# SSH into machine (debugging)
flyctl ssh --app counterspell-data-plane --machine abc123def456
```

### Via API
```bash
# Get machine status
GET /v1/apps/{app_name}/machines/{machine_id}

# List all machines
GET /v1/apps/{app_name}/machines
```

### Metrics to Track
- Machine uptime
- Cold start duration
- Auto-sleep frequency
- CPU/memory usage
- Health check failures

---

## Troubleshooting

### Machine Won't Start
1. Check logs: `flyctl logs --app counterspell-data-plane --machine {id}`
2. Verify image exists: `flyctl images list --app counterspell-data-plane`
3. Check resource limits: CPU, memory, disk
4. Validate environment variables

### Machine Stays in "Starting" State
1. Check for health endpoint: `GET /health` must return 200
2. Verify port configuration matches `fly.toml`
3. Check for application errors in logs

### Auto-Sleep Not Working
1. Verify `auto_stop_machines = true` in `fly.toml`
2. Check if there are long-running connections
3. Ensure health checks don't keep machine awake

### Volume Not Mounting
1. Check volume exists: `flyctl volumes list --app counterspell-data-plane`
2. Verify volume name matches mount config
3. Check region matches machine region

---

## Summary

**Fly.io Primitives for Counterspell**:
1. **Machines** - One per user (not Apps)
2. **shared-cpu-1x** - 1GB RAM for cost efficiency
3. **Volumes** - 1GB per user for persistence
4. **Auto-Sleep** - 30min timeout for 80% cost savings
5. **Firecracker** - Fast cold starts (5-10s)

**Cost**: ~$1-7/month per user depending on usage patterns

**Next Steps**:
1. Set up Fly.io account and API token
2. Create base app and deploy image
3. Implement Fly.io API client in invoker
4. Add machine creation to user registration flow
5. Test machine creation, sleep, and wake
