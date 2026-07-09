# VPS Deployment Guide

This document covers how to deploy the backend to a VPS using GitHub Actions + GitHub Container Registry (GHCR) + Docker Compose.

---

## Overview

```
Developer pushes to master
        │
        ▼
GitHub Actions (.github/workflows/deploy.yml)
  1. Build Docker image from ./backend/Dockerfile
  2. Push image to ghcr.io/<your-repo>:latest + :sha-<commit>
  3. SSH into VPS → docker compose pull + up
        │
        ▼
VPS (docker-compose.yml)
  postgres container  ←→  oil-change-api container
        │
   Nginx / Caddy (reverse proxy, TLS)
        │
   Internet
```

---

## GitHub Secrets Required

Go to **GitHub → Repository → Settings → Secrets and variables → Actions** and add:

| Secret | Description | Example |
|---|---|---|
| `VPS_HOST` | VPS IP address or domain | `123.45.67.89` |
| `VPS_USER` | SSH login user | `deploy` |
| `VPS_SSH_KEY` | Full private SSH key (RSA or Ed25519) | `-----BEGIN OPENSSH PRIVATE KEY-----...` |
| `VPS_PORT` | SSH port | `22` |
| `VPS_APP_DIR` | Absolute path to your app dir on VPS | `/home/deploy/oil-change-app` |

> **`GITHUB_TOKEN`** is automatic — no action needed. It is used to authenticate docker push/pull to GHCR.

### Creating the SSH key pair

```bash
# On your local machine
ssh-keygen -t ed25519 -C "github-actions-deploy" -f ~/.ssh/deploy_key -N ""

# Copy the PUBLIC key to the VPS's authorized_keys
ssh-copy-id -i ~/.ssh/deploy_key.pub deploy@your-vps-ip

# Paste the PRIVATE key into the GitHub secret VPS_SSH_KEY
cat ~/.ssh/deploy_key
```

---

## First-Time VPS Setup

Run these commands **once** on the VPS:

```bash
# 1. Install Docker + Compose plugin
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER
newgrp docker

# 2. Create the app directory
mkdir -p /home/deploy/oil-change-app
cd /home/deploy/oil-change-app

# 3. Create the .env file (see "Environment Management" section below)
nano .env

# 4. Pull and start (GitHub Actions will handle this from now on,
#    but run it manually the first time to verify)
echo "<GITHUB_PAT>" | docker login ghcr.io -u <your-github-username> --password-stdin
docker pull ghcr.io/<your-github-username>/customer-oil-change-management/backend:latest
docker compose -f docker-compose.yml up -d
```

> **Note:** `docker-compose.prod.yml` itself is NOT committed to git (it is gitignored).
> Copy it to your VPS manually once and rename it to `docker-compose.yml` to match the deployment workflow:
> ```bash
> scp backend/docker-compose.prod.yml deploy@your-vps-ip:/home/deploy/oil-change-app/docker-compose.yml
> ```

---

## Environment Management on VPS

### The Golden Rule

> **Secrets live in one place: the `.env` file on the VPS.**  
> They are never in git, never in the Docker image, never in environment variables set in the CI workflow.

### The `.env` file on the VPS

Create `/home/deploy/oil-change-app/.env`:

```bash
# Server
APP_ENV=production
SERVER_PORT=8080

# Database — use strong passwords, not the defaults
DB_USER=oilchange_user
DB_PASSWORD=<strong-random-password-here>
DB_NAME=oil_change_db

# JWT — must be at least 32 random characters
JWT_SECRET=<generate-with: openssl rand -hex 32>

# Token expiry
ACCESS_TOKEN_EXPIRY_MINUTES=15
REFRESH_TOKEN_EXPIRY_DAYS=7

# Docker image repo (used by docker-compose.yml)
GITHUB_REPO=<your-github-username>/customer-oil-change-management
```

**Lock down file permissions:**

```bash
chmod 600 /home/deploy/oil-change-app/.env
chown deploy:deploy /home/deploy/oil-change-app/.env
```

### Generating secure values

```bash
# JWT secret (64 hex chars = 256-bit entropy)
openssl rand -hex 32

# Strong DB password
openssl rand -base64 24
```

### Updating secrets

```bash
# On the VPS, edit the file
nano /home/deploy/oil-change-app/.env

# Then restart the api container to pick up the new values
cd /home/deploy/oil-change-app
docker compose -f docker-compose.yml up -d oil-change-api
```

---

## Database Migrations in Production

When `APP_ENV` is set to `production`, the API service **does not** automatically apply SQL migrations on startup. This prevents table-locking or accidental data issues.

To run migrations manually in production:

```bash
# On the VPS
cd /home/deploy/oil-change-app

# Run a one-off container to apply the migrations and exit
docker compose -f docker-compose.yml run --rm oil-change-api -migrate
```

---

## Deployment Flow (After First-Time Setup)

Every `git push` to `master` (with any change inside `backend/`) automatically:

1. GitHub Actions builds the Docker image
2. Pushes it to `ghcr.io/<your-repo>:latest` and `ghcr.io/<your-repo>:sha-<shortcommit>`
3. SSHs into your VPS, pulls the new image, and restarts the `oil-change-api` container

**Expected downtime:** ~3 seconds (stop old container → start new one).

---

## Rollback

If a deploy breaks production, roll back instantly using the previous image tag:

```bash
# On the VPS
cd /home/deploy/oil-change-app

# List available image tags (sha-* tags are each individual commit)
docker images ghcr.io/<your-repo>

# Override the image tag and restart
IMAGE_TAG=sha-abc1234 \
  docker compose -f docker-compose.yml up -d oil-change-api
```

---

## Security Checklist

- [x] `.env` is gitignored (both `.env` and `.env.*` patterns covered)
- [x] Docker ports bound to `127.0.0.1` only (not `0.0.0.0`)
- [x] API container runs as non-root user (`appuser`)
- [x] PostgreSQL is not exposed to the internet (only `127.0.0.1:5432`)
- [ ] Place Nginx or Caddy in front of the API for TLS (HTTPS)
- [ ] Set up a firewall (`ufw allow 22,80,443`)
- [ ] Enable automatic security updates (`unattended-upgrades` on Ubuntu)
- [ ] Set up log rotation for Docker logs

---

## Reference

| File | Location | Purpose |
|---|---|---|
| `deploy.yml` | `.github/workflows/` in repo | GitHub Actions CI/CD pipeline |
| `docker-compose.prod.yml` | Local machine (not in git) | Production compose configuration (copied to VPS as `docker-compose.yml`) |
| `.env` | VPS only — created manually | All production secrets |
| `.env.example` | In repo | Template — safe to commit |
