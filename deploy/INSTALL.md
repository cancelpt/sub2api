# Install with Docker Compose

This document tells another AI how to install and run this Docker Compose
deployment from scratch on a different machine.

## Source repository

This project is published from:

- `https://github.com/cancelpt/sub2api`

## What is special about this version

The defining feature of this version is OpenAI strict scheduling with optional
same-account retry.

- It supports the runtime `OpenAI Strict Primary Fallback` toggle in the admin
  panel.
- It keeps the default scheduler as `weighted_topk` until an admin explicitly
  enables strict mode.
- When strict scheduling is enabled, new OpenAI requests prefer higher-priority
  accounts first, and only fall back when the primary account cannot schedule.
- It also supports `Strict Same-Account Retry` and `Strict Retry Count` in the
  admin panel. When enabled, failover-eligible upstream errors that are not
  covered by the account `custom_error_codes` list will retry on the same
  account before switching to a fallback account.
- It also supports `Strict Retry Delay (ms)` in the admin panel, so the wait
  time between strict same-account retries is configurable instead of being
  fixed at `500ms`.
- Sticky routing for existing `previous_response_id` and `session_hash` flows is
  still preserved.
- This version does not require any Docker Compose env flag for strict mode.
  Strict behavior is controlled at runtime in the admin panel.
- This version also includes compatibility handling so older deployments and
  older settings clients do not accidentally reset the strict scheduler or retry
  settings.

## Prerequisites

Make sure the target machine already has:

- `git`
- `docker`
- `docker compose`

## Fresh install from zero

### 1. Clone the repository

```bash
git clone https://github.com/cancelpt/sub2api.git
cd sub2api
```

### 2. Make the machine-local override stay out of git

Run from the repository root:

```bash
mkdir -p .git/info
grep -qxF '/deploy/docker-compose.machine.local.yml' .git/info/exclude 2>/dev/null || \
  printf '/deploy/docker-compose.machine.local.yml\n' >> .git/info/exclude
```

This keeps the machine-local image pin file from appearing in normal git
status/commit flows on the target machine.

### 3. Prepare deployment directories

```bash
cd deploy
cp .env.example .env
mkdir -p data postgres_data redis_data
```

### 4. Edit `.env`

At minimum, set these values in `deploy/.env`:

- `POSTGRES_PASSWORD`
- `JWT_SECRET`
- `TOTP_ENCRYPTION_KEY`
- `ADMIN_EMAIL`
- `ADMIN_PASSWORD` if you do not want it auto-generated
- `SERVER_PORT` if you do not want the default host port
- `TZ` if the server timezone should not use the default

Recommended secret generation examples:

```bash
openssl rand -hex 32
```

### 5. Create the machine-local image override

Create this file:

`<repo-root>/deploy/docker-compose.machine.local.yml`

Contents:

```yaml
services:
  sub2api:
    image: sub2api:cancelpt-local
```

Do not edit the tracked `deploy/docker-compose.local.yml` just to pin the app
image. Keep image pinning in `docker-compose.machine.local.yml`.

### 6. Build the image for the current checkout

Run from the repository root:

```bash
cd <repo-root>
docker build \
  -t sub2api:cancelpt-local \
  --build-arg COMMIT="$(git rev-parse HEAD)" \
  --build-arg DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --build-arg GOPROXY=https://goproxy.cn,direct \
  --build-arg GOSUMDB=sum.golang.google.cn \
  .
```

### 7. Review the effective Compose config

Run from the deploy directory:

```bash
cd <repo-root>/deploy
docker compose \
  -f docker-compose.local.yml \
  -f docker-compose.machine.local.yml \
  config
```

Verify that the `sub2api` service resolves to:

```yaml
image: sub2api:cancelpt-local
```

### 8. Start the stack

```bash
cd <repo-root>/deploy
docker compose \
  -f docker-compose.local.yml \
  -f docker-compose.machine.local.yml \
  up -d
```

This will start:

- `sub2api`
- `postgres`
- `redis`

### 9. Enable strict scheduling in the admin panel if needed

Strict mode is a runtime setting now. After the stack is up:

1. Sign in to the admin panel.
2. Open `System Settings`.
3. Find `Gateway Scheduling Settings`.
4. Leave it unchanged if you want the default `weighted_topk` behavior.
5. Enable `OpenAI Strict Primary Fallback` only if you want strict priority
   scheduling for new OpenAI requests.
6. Optionally enable `Strict Same-Account Retry`, then set `Strict Retry Count`
   and `Strict Retry Delay (ms)` to control how many same-account retries
   happen before failover and how long each retry waits.
7. Save settings.

No Docker restart is required for these scheduler toggles.

## Post-install verification

### 1. Check container status

```bash
docker inspect sub2api \
  --format 'Image={{.Config.Image}} Running={{.State.Running}} Status={{.State.Status}} Health={{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}'
```

Expected:

- image is `sub2api:cancelpt-local`
- running is `true`
- health becomes `healthy`

### 2. Check health endpoint

Replace `<published-port>` with the host port you mapped through Compose.
On many machines this will match the `SERVER_PORT` value from `.env`.

```bash
curl -fsS http://127.0.0.1:<published-port>/health
```

Expected response:

```json
{"status":"ok"}
```

### 3. Check logs if needed

```bash
docker logs --tail 200 sub2api
```

Or:

```bash
cd <repo-root>/deploy
docker compose \
  -f docker-compose.local.yml \
  -f docker-compose.machine.local.yml \
  logs -f sub2api
```
