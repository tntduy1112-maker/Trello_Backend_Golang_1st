# TaskFlow — Deployment Guide

> Hướng dẫn triển khai TaskFlow với Docker cho môi trường Development và Production.

---

## Tổng quan Infrastructure

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                           PRODUCTION ARCHITECTURE                             │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   Internet                                                                   │
│       │                                                                      │
│       ▼                                                                      │
│   ┌─────────┐    ┌─────────────────────────────────────────────────────┐    │
│   │  Nginx  │    │              Docker Network (internal)              │    │
│   │  :443   │    │  ┌─────────────────────────────────────────────┐   │    │
│   │  :80    │───▶│  │                                             │   │    │
│   └─────────┘    │  │    ┌─────────┐    ┌─────────┐               │   │    │
│                  │  │    │  API 1  │    │  API 2  │  (replicas)   │   │    │
│                  │  │    │  :8080  │    │  :8080  │               │   │    │
│                  │  │    └────┬────┘    └────┬────┘               │   │    │
│                  │  │         │              │                    │   │    │
│                  │  │         └──────┬───────┘                    │   │    │
│                  │  │                │                            │   │    │
│                  │  │    ┌───────────┼───────────┐                │   │    │
│                  │  │    ▼           ▼           ▼                │   │    │
│                  │  │ ┌──────┐  ┌─────────┐  ┌───────┐            │   │    │
│                  │  │ │Postgres│ │  Redis  │  │ MinIO │            │   │    │
│                  │  │ │ :5432 │  │  :6379  │  │ :9000 │            │   │    │
│                  │  │ └──────┘  └─────────┘  └───────┘            │   │    │
│                  │  │                                             │   │    │
│                  │  └─────────────────────────────────────────────┘   │    │
│                  └─────────────────────────────────────────────────────┘    │
│                                                                              │
│   Volumes: postgres_data, redis_data, minio_data                            │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## 1. Development Setup

### 1.1 Prerequisites

```bash
# Required
docker --version          # Docker 24+
docker compose version    # Docker Compose v2+

# Optional (for local Go development)
go version                # Go 1.23+
```

### 1.2 Quick Start

```bash
# 1. Clone repository
git clone git@github.com:tntduy1112-maker/Trello_Backend_Golang_1st.git
cd Trello_Backend_Golang_1st

# 2. Copy environment file
cp .env.example .env

# 3. Start all services
make dev
# hoặc
docker compose up

# 4. View logs
make logs
```

### 1.3 Service URLs (Development)

| Service | URL | Credentials |
|---------|-----|-------------|
| **API** | http://localhost:8080 | — |
| **API Health** | http://localhost:8080/health | — |
| **API Docs** | http://localhost:8080/swagger/index.html | — |
| **PostgreSQL** | localhost:5432 | taskflow / taskflow_secret |
| **Redis** | localhost:6379 | redis_secret |
| **MinIO Console** | http://localhost:9001 | minioadmin / minioadmin123 |
| **MinIO API** | http://localhost:9000 | minioadmin / minioadmin123 |
| **MailHog UI** | http://localhost:8025 | — |

### 1.4 Useful Commands

```bash
# Start/Stop
make up                   # Start in background
make down                 # Stop all services
make restart              # Restart all

# Logs
make logs                 # All services
make logs-api             # API only

# Database
make db-shell             # PostgreSQL shell
make migrate              # Run migrations
make migrate-down         # Rollback last migration

# Redis
make redis-cli            # Redis CLI

# Cleanup
make clean                # Remove containers + volumes
```

---

## 2. Docker Services

### 2.1 docker-compose.yml (Development)

```yaml
services:
  postgres:     # PostgreSQL 16 - Primary database
  redis:        # Redis 7 - Cache, rate limit, JWT blacklist
  minio:        # MinIO - S3-compatible file storage
  minio-setup:  # One-time bucket creation
  api:          # Go API server
  mailhog:      # Development email testing
```

### 2.2 Service Configuration

#### PostgreSQL

```yaml
postgres:
  image: postgres:16-alpine
  environment:
    POSTGRES_USER: taskflow
    POSTGRES_PASSWORD: taskflow_secret
    POSTGRES_DB: taskflow
  ports:
    - "5432:5432"
  volumes:
    - postgres_data:/var/lib/postgresql/data
    - ./scripts/init-db.sql:/docker-entrypoint-initdb.d/init.sql
```

#### Redis

```yaml
redis:
  image: redis:7-alpine
  command: >
    redis-server
    --requirepass redis_secret
    --maxmemory 128mb
    --maxmemory-policy allkeys-lru
    --appendonly yes
  ports:
    - "6379:6379"
```

#### MinIO

```yaml
minio:
  image: minio/minio:latest
  command: server /data --console-address ":9001"
  environment:
    MINIO_ROOT_USER: minioadmin
    MINIO_ROOT_PASSWORD: minioadmin123
  ports:
    - "9000:9000"   # API
    - "9001:9001"   # Console
```

---

## 3. Production Deployment

### 3.1 Prerequisites

- Linux server (Ubuntu 22.04+ recommended)
- Docker + Docker Compose
- Domain name + SSL certificate
- Strong passwords (generate with `openssl rand -base64 32`)

### 3.2 Production Environment Variables

```bash
# .env.production (NEVER commit this file!)

# App
APP_ENV=production
APP_URL=https://api.taskflow.example.com
LOG_LEVEL=info

# Database
DB_USER=taskflow_prod
DB_PASSWORD=<strong-password-here>
DB_NAME=taskflow_production

# Redis
REDIS_PASSWORD=<strong-password-here>

# JWT (generate with: openssl rand -base64 32)
JWT_ACCESS_SECRET=<64-char-random-string>
JWT_REFRESH_SECRET=<64-char-random-string>

# AES (exactly 32 bytes)
AES_KEY=<32-byte-random-string>

# MinIO
MINIO_ROOT_USER=taskflow_minio
MINIO_ROOT_PASSWORD=<strong-password-here>
MINIO_BUCKET=taskflow-prod

# SMTP (production email service)
SMTP_HOST=smtp.resend.com
SMTP_PORT=587
SMTP_USER=resend
SMTP_PASS=re_xxxx
SMTP_FROM=noreply@taskflow.example.com

# Frontend
FRONTEND_URL=https://taskflow.example.com
CORS_ORIGINS=https://taskflow.example.com
```

### 3.3 Deploy Steps

```bash
# 1. SSH into server
ssh user@your-server

# 2. Clone repository
git clone <repo-url>
cd taskflow

# 3. Create production env file
cp .env.example .env.production
nano .env.production  # Edit with production values

# 4. Create SSL certificates directory
mkdir -p nginx/ssl
# Copy your SSL certificates:
# - nginx/ssl/fullchain.pem
# - nginx/ssl/privkey.pem

# 5. Build and start production stack
docker compose -f docker-compose.prod.yml --env-file .env.production up -d

# 6. Run database migrations
docker compose -f docker-compose.prod.yml exec api /app/server migrate up

# 7. Verify deployment
curl https://api.taskflow.example.com/health
```

### 3.4 Production docker-compose.prod.yml

Key differences from development:

| Aspect | Development | Production |
|--------|-------------|------------|
| **Restart policy** | `unless-stopped` | `always` |
| **Replicas** | 1 | 2+ (load balanced) |
| **Networks** | Single network | Internal + External |
| **Ports** | All exposed | Only Nginx 80/443 |
| **Logging** | Debug | Info |
| **SSL** | None | Required |
| **Secrets** | In .env | Docker secrets / Vault |

### 3.5 Nginx Configuration

```nginx
# Key features in nginx/nginx.conf:

# Rate limiting
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;
limit_req_zone $binary_remote_addr zone=auth_limit:10m rate=5r/m;

# Load balancing
upstream api_servers {
    least_conn;
    server api:8080;
    keepalive 32;
}

# SSE support (no buffering)
location /api/v1/notifications/stream {
    proxy_buffering off;
    proxy_cache off;
    proxy_read_timeout 86400s;
}

# SSL configuration (TLS 1.2+)
ssl_protocols TLSv1.2 TLSv1.3;
ssl_prefer_server_ciphers off;
```

---

## 4. Database Migrations

### 4.1 Migration Commands

```bash
# Development
make migrate              # Apply all pending migrations
make migrate-down         # Rollback last migration
make migrate-new NAME=add_users_table  # Create new migration

# Production
docker compose -f docker-compose.prod.yml exec api /app/server migrate up
docker compose -f docker-compose.prod.yml exec api /app/server migrate status
```

### 4.2 Migration Files

```
migrations/
├── 00001_create_users.sql
├── 00002_create_refresh_tokens.sql
├── 00003_create_email_verifications.sql
├── 00004_create_organizations.sql
├── 00005_create_organization_members.sql
├── 00006_create_boards.sql
├── 00007_create_board_members.sql
├── 00008_create_board_invitations.sql
├── 00009_create_lists.sql
├── 00010_create_cards.sql
├── ...
└── 00020_add_indexes.sql
```

### 4.3 Migration Best Practices

```sql
-- migrations/00001_create_users.sql

-- +goose Up
CREATE TABLE users (
    id VARCHAR(25) PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    ...
);

-- +goose Down
DROP TABLE IF EXISTS users;
```

---

## 5. Backup & Restore

### 5.1 Database Backup

```bash
# Manual backup
docker compose exec postgres pg_dump -U taskflow -d taskflow > backup_$(date +%Y%m%d).sql

# Compressed backup
docker compose exec postgres pg_dump -U taskflow -d taskflow | gzip > backup_$(date +%Y%m%d).sql.gz

# Automated backup script (add to crontab)
#!/bin/bash
BACKUP_DIR=/backups
DATE=$(date +%Y%m%d_%H%M%S)
docker compose exec -T postgres pg_dump -U taskflow -d taskflow | gzip > $BACKUP_DIR/taskflow_$DATE.sql.gz
# Keep only last 7 days
find $BACKUP_DIR -name "*.sql.gz" -mtime +7 -delete
```

### 5.2 Database Restore

```bash
# Restore from backup
gunzip -c backup_20240420.sql.gz | docker compose exec -T postgres psql -U taskflow -d taskflow

# Or for plain SQL
cat backup.sql | docker compose exec -T postgres psql -U taskflow -d taskflow
```

### 5.3 MinIO Backup

```bash
# Using mc (MinIO Client)
mc alias set prod http://minio:9000 minioadmin minioadmin123
mc mirror prod/taskflow ./minio-backup/

# Restore
mc mirror ./minio-backup/ prod/taskflow
```

---

## 6. Monitoring

### 6.1 Health Checks

```bash
# API health
curl http://localhost:8080/health

# Expected response
{
  "status": "healthy",
  "version": "1.0.0",
  "services": {
    "database": "connected",
    "redis": "connected",
    "minio": "connected"
  }
}
```

### 6.2 Docker Health Status

```bash
# Check all container health
docker compose ps

# View resource usage
docker stats

# View logs
docker compose logs -f --tail=100 api
```

### 6.3 Prometheus Metrics (Optional)

```yaml
# Add to docker-compose.prod.yml
prometheus:
  image: prom/prometheus:latest
  volumes:
    - ./prometheus.yml:/etc/prometheus/prometheus.yml
  ports:
    - "9090:9090"

grafana:
  image: grafana/grafana:latest
  ports:
    - "3000:3000"
  environment:
    GF_SECURITY_ADMIN_PASSWORD: admin
```

---

## 7. Scaling

### 7.1 Horizontal Scaling (API)

```yaml
# docker-compose.prod.yml
api:
  deploy:
    mode: replicated
    replicas: 3  # Scale to 3 instances
    update_config:
      parallelism: 1
      delay: 10s
      order: start-first  # Zero-downtime deployment
```

```bash
# Scale manually
docker compose -f docker-compose.prod.yml up -d --scale api=3
```

### 7.2 Database Scaling

```yaml
# Read replica (add to docker-compose.prod.yml)
postgres-replica:
  image: postgres:16-alpine
  environment:
    POSTGRES_USER: taskflow
    POSTGRES_PASSWORD: ${DB_PASSWORD}
  command: |
    postgres -c wal_level=replica
             -c max_wal_senders=3
             -c hot_standby=on
```

### 7.3 Redis Cluster (High Availability)

```yaml
# For production, consider Redis Sentinel or Redis Cluster
redis-master:
  image: redis:7-alpine
  
redis-slave:
  image: redis:7-alpine
  command: redis-server --slaveof redis-master 6379
  
redis-sentinel:
  image: redis:7-alpine
  command: redis-sentinel /etc/redis/sentinel.conf
```

---

## 8. Troubleshooting

### 8.1 Common Issues

| Issue | Solution |
|-------|----------|
| `connection refused` to database | Wait for health check: `docker compose logs postgres` |
| `permission denied` on volumes | Fix ownership: `sudo chown -R 1000:1000 ./data` |
| MinIO buckets not created | Restart minio-setup: `docker compose restart minio-setup` |
| API can't connect to Redis | Check password in `.env` matches Redis config |
| SSL certificate errors | Verify cert paths in `nginx/ssl/` |

### 8.2 Debug Commands

```bash
# Enter container shell
docker compose exec api sh
docker compose exec postgres psql -U taskflow -d taskflow

# View container logs
docker compose logs -f api
docker compose logs --tail=100 postgres

# Check network connectivity
docker compose exec api ping postgres
docker compose exec api nc -zv redis 6379

# Check environment variables
docker compose exec api env | grep DB

# Inspect container
docker inspect taskflow-api
```

### 8.3 Reset Everything

```bash
# Full reset (WARNING: destroys all data)
make clean-all

# Or manually:
docker compose down -v --remove-orphans
docker system prune -af
docker volume prune -f
```

---

## 9. CI/CD Pipeline

### 9.1 GitHub Actions Example

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Build Docker image
        run: docker build -t ghcr.io/${{ github.repository }}/api:${{ github.sha }} .
      
      - name: Push to registry
        run: |
          echo ${{ secrets.GITHUB_TOKEN }} | docker login ghcr.io -u ${{ github.actor }} --password-stdin
          docker push ghcr.io/${{ github.repository }}/api:${{ github.sha }}
  
  deploy:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to server
        uses: appleboy/ssh-action@master
        with:
          host: ${{ secrets.SERVER_HOST }}
          username: ${{ secrets.SERVER_USER }}
          key: ${{ secrets.SSH_KEY }}
          script: |
            cd /app/taskflow
            docker compose -f docker-compose.prod.yml pull
            docker compose -f docker-compose.prod.yml up -d
```

### 9.2 Version Tagging

```bash
# Tag release
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0

# Build with version
docker build -t taskflow-api:v1.0.0 --build-arg VERSION=v1.0.0 .
```

---

## 10. Security Checklist

### Pre-Deployment

- [ ] Strong passwords for all services (32+ chars)
- [ ] JWT secrets are unique and strong
- [ ] AES key is exactly 32 bytes
- [ ] SSL certificates are valid
- [ ] `.env.production` is NOT in git
- [ ] Database is not exposed to internet
- [ ] Redis requires password
- [ ] MinIO has strong credentials

### Post-Deployment

- [ ] Health endpoints respond correctly
- [ ] Rate limiting is working
- [ ] SSL certificate is valid (check with SSL Labs)
- [ ] Backup cron job is configured
- [ ] Monitoring alerts are set up
- [ ] Firewall rules are configured

---

## Quick Reference

```bash
# Development
make dev                  # Start all
make logs                 # View logs
make db-shell             # PostgreSQL CLI
make clean                # Full cleanup

# Production
docker compose -f docker-compose.prod.yml up -d
docker compose -f docker-compose.prod.yml logs -f
docker compose -f docker-compose.prod.yml exec api /app/server migrate up

# Backup
docker compose exec postgres pg_dump -U taskflow -d taskflow > backup.sql

# Scale
docker compose -f docker-compose.prod.yml up -d --scale api=3
```
