# Phase 1: Project Setup & Authentication (2 Weeks)

> Detailed execution plan for the foundation and authentication module.

---

## Overview

```
Week 1: Foundation          Week 2: Auth Module
┌──────────────────┐        ┌──────────────────┐
│ 1.1 Go Project   │        │ 1.6 User Domain  │
│ 1.2 Config       │        │ 1.7 Auth Service │
│ 1.3 Database     │───────▶│ 1.8 Handlers     │
│ 1.4 Redis        │        │ 1.9 Email        │
│ 1.5 Packages     │        │ 1.10 Storage     │
└──────────────────┘        └──────────────────┘
```

---

## Week 1: Foundation (5 Tasks)

| Task | Files to Create | Priority | Est. Time |
|------|-----------------|:--------:|:---------:|
| **1.1** Initialize Go Project | `go.mod`, update `cmd/server/main.go` | P0 | 1h |
| **1.2** Configuration | `internal/config/config.go` | P0 | 2h |
| **1.3** Database + Migrations | `pkg/database/postgres.go`, 3 migration files | P0 | 3h |
| **1.4** Redis Connection | `pkg/cache/redis.go` | P0 | 1h |
| **1.5** Shared Packages | 6 packages in `pkg/` | P0 | 4h |

---

### Task 1.1: Initialize Go Project

```
Action: Install dependencies, setup entry point
Files:
  - go.mod (add dependencies)
  - cmd/server/main.go (Gin server with /health)
  
Dependencies to add:
  gin, pgx, go-redis, golang-jwt, validator, viper, 
  zerolog, minio-go, cuid2, bcrypt, goose, asynq

Verify: go build ./... && curl localhost:8080/health
```

---

### Task 1.2: Configuration

```
Action: Create config loader with Viper
Files:
  - internal/config/config.go

Structs needed:
  - AppConfig (name, env, port, url)
  - DatabaseConfig (host, port, user, password, name, pool)
  - RedisConfig (host, port, password)
  - JWTConfig (access_secret, refresh_secret, access_expiry, refresh_expiry)
  - MinIOConfig (endpoint, access_key, secret_key, bucket, use_ssl)
  - SMTPConfig (host, port, user, pass, from)

Verify: Config loads from .env, validates required fields
```

---

### Task 1.3: Database + Migrations

```
Action: Setup PostgreSQL connection pool + auth tables
Files:
  - pkg/database/postgres.go
  - migrations/00001_create_users.sql
  - migrations/00002_create_refresh_tokens.sql
  - migrations/00003_create_email_verifications.sql

Tables (from DATABASE_DESIGN.md):
  - users (id, email, password_hash, full_name, avatar_url, is_verified, is_active, tokens_valid_after, timestamps)
  - refresh_tokens (id, user_id, token_hash, device_info, ip_address, is_revoked, expires_at, timestamps)
  - email_verifications (id, user_id, token, type, expires_at, used_at, created_at)

Verify: goose up / goose down work
```

---

### Task 1.4: Redis Connection

```
Action: Setup Redis client with connection pool
Files:
  - pkg/cache/redis.go

Functions needed:
  - NewRedisClient(config) → *redis.Client
  - Set(key, value, ttl)
  - Get(key) → value
  - Delete(key)
  - Exists(key) → bool

Verify: SET/GET operations work
```

---

### Task 1.5: Shared Packages (6 packages)

```
pkg/
├── apperror/errors.go    ← Custom error types (from ERROR_CODES.md)
├── cuid/cuid.go          ← CUID generator wrapper
├── hash/bcrypt.go        ← HashPassword, ComparePassword
├── jwt/jwt.go            ← GenerateAccessToken, GenerateRefreshToken, Verify
├── crypto/aes.go         ← AES-256-GCM Encrypt/Decrypt (for cookies)
└── validator/validator.go ← Custom validation rules

Each package needs unit tests!
```

---

## Week 2: Authentication Module (5 Tasks)

| Task | Files to Create | Priority | Est. Time |
|------|-----------------|:--------:|:---------:|
| **1.6** User Domain & Repo | `internal/domain/`, `internal/repository/` | P0 | 3h |
| **1.7** Auth Service | `internal/service/auth_service.go`, DTOs | P0 | 6h |
| **1.8** Auth Handler & Routes | `internal/handler/`, `internal/middleware/` | P0 | 5h |
| **1.9** Email Service | `pkg/email/email.go` | P0 | 2h |
| **1.10** MinIO Storage | `pkg/storage/minio.go` | P1 | 2h |

---

### Task 1.6: User Domain & Repository

```
Action: Create User model and data access layer
Files:
  - internal/domain/user.go
  - internal/domain/refresh_token.go
  - internal/domain/email_verification.go
  - internal/repository/interfaces.go
  - internal/repository/user_repository.go
  - internal/repository/token_repository.go

Repository methods:
  UserRepository:
    - Create(user) → user
    - FindByID(id) → user
    - FindByEmail(email) → user
    - Update(user) → user
    - SoftDelete(id)
    
  TokenRepository:
    - CreateRefreshToken(token)
    - FindByHash(hash) → token
    - RevokeToken(hash)
    - RevokeAllUserTokens(userID)
    - CreateEmailVerification(verification)
    - FindVerification(token, type) → verification
    - MarkVerificationUsed(id)
```

---

### Task 1.7: Auth Service (Core Logic)

```
Action: Implement all auth business logic
Files:
  - internal/service/auth_service.go
  - internal/dto/request/auth_request.go
  - internal/dto/response/auth_response.go

11 Methods (from AUTH_FLOW.md):
  1. Register(email, password, fullName) → user + send OTP
  2. VerifyEmail(email, otp) → mark verified
  3. ResendVerification(email) → new OTP
  4. Login(email, password) → accessToken + refreshToken cookie
  5. RefreshToken(encryptedCookie) → new tokens (rotation!)
  6. Logout(accessToken, refreshToken) → blacklist + revoke
  7. LogoutAll(userID) → revoke all + invalidate
  8. ForgotPassword(email) → send reset link
  9. ResetPassword(token, password) → update + revoke all
  10. GetCurrentUser(userID) → user
  11. UpdateProfile(userID, name, avatar) → updated user

Key flows:
  - Token rotation on refresh
  - Reuse detection → nuke all sessions
  - Redis blacklist for immediate AT revocation
  - tokens_valid_after for mass invalidation
```

---

### Task 1.8: Auth Handler & Routes

```
Action: HTTP handlers + middleware
Files:
  - internal/handler/auth_handler.go
  - internal/middleware/auth.go (JWT validation)
  - internal/middleware/ratelimit.go
  - internal/middleware/error_handler.go

11 Endpoints:
  POST /auth/register         [3/hour/IP]
  POST /auth/verify-email     [5 attempts/OTP]
  POST /auth/resend-verification [3/hour/email]
  POST /auth/login            [5/15min/IP]
  POST /auth/refresh          [30/15min/user]
  POST /auth/logout           [no limit, optional auth]
  POST /auth/logout-all       [requires auth]
  POST /auth/forgot-password  [5/15min/IP]
  POST /auth/reset-password   [5/15min/IP]
  GET  /auth/me               [requires auth]
  PUT  /auth/me               [requires auth, multipart]

Middleware chain:
  Request → RateLimit → [Auth] → Handler → ErrorHandler → Response
```

---

### Task 1.9: Email Service

```
Action: SMTP email sender for OTP and reset links
Files:
  - pkg/email/email.go
  - pkg/email/templates/ (optional HTML templates)

Methods:
  - SendVerificationEmail(to, otp) → sends 6-digit OTP
  - SendPasswordResetEmail(to, token) → sends reset link
  - SendBoardInvitation(to, inviter, board, token) → (for later)

Config: Uses MailHog in dev (localhost:1025)
```

---

### Task 1.10: MinIO Storage

```
Action: S3-compatible file storage for avatars
Files:
  - pkg/storage/minio.go

Methods:
  - Upload(folder, filename, reader, size, contentType) → objectKey
  - Delete(objectKey)
  - GetPublicURL(objectKey) → URL string

Used by: PUT /auth/me for avatar upload
```

---

## Dependency Graph

```
Task 1.1 (Go Project)
    │
    ├──▶ Task 1.2 (Config)
    │        │
    │        ├──▶ Task 1.3 (Database) ──▶ Task 1.6 (User Repo)
    │        │                                   │
    │        ├──▶ Task 1.4 (Redis) ─────────────┤
    │        │                                   │
    │        └──▶ Task 1.5 (Packages) ──────────┤
    │                                            │
    │                                            ▼
    │                                     Task 1.7 (Auth Service)
    │                                            │
    ├──▶ Task 1.9 (Email) ──────────────────────┤
    │                                            │
    ├──▶ Task 1.10 (MinIO) ─────────────────────┤
    │                                            │
    │                                            ▼
    └───────────────────────────────────▶ Task 1.8 (Handlers)
```

---

## Recommended Execution Order

| Day | Tasks | Focus |
|:---:|-------|-------|
| 1 | 1.1, 1.2 | Project setup + config |
| 2 | 1.3 | Database + migrations |
| 3 | 1.4, 1.5 (partial) | Redis + apperror, cuid, hash |
| 4 | 1.5 (complete) | jwt, crypto, validator + tests |
| 5 | 1.6 | User domain + repository |
| 6-7 | 1.7 | Auth service (core logic) |
| 8 | 1.9, 1.10 | Email + MinIO |
| 9 | 1.8 | Handlers + middleware |
| 10 | Integration | End-to-end testing |

---

## Phase 1 Deliverables Checklist

### Infrastructure
- [ ] Go project compiles (`go build ./...`)
- [ ] Docker services start (`make dev`)
- [ ] Database migrations work (`goose up/down`)
- [ ] Redis connection works
- [ ] MinIO connection works

### Auth Endpoints (all working)
- [ ] `POST /auth/register` → 201 + OTP email
- [ ] `POST /auth/verify-email` → 200
- [ ] `POST /auth/login` → 200 + AT + RT cookie
- [ ] `POST /auth/refresh` → 200 + new tokens (rotation)
- [ ] `POST /auth/logout` → 200 + blacklist AT
- [ ] `POST /auth/logout-all` → 200 + invalidate all
- [ ] `POST /auth/forgot-password` → 200 + email
- [ ] `POST /auth/reset-password` → 200
- [ ] `GET /auth/me` → 200 + user
- [ ] `PUT /auth/me` → 200 + updated user

### Security
- [ ] Rate limiting works on all endpoints
- [ ] JWT middleware validates correctly
- [ ] Redis blacklist blocks revoked tokens
- [ ] Token rotation on refresh
- [ ] Reuse detection works

### Tests
- [ ] Unit tests for `pkg/*` (80% coverage)
- [ ] Integration tests for auth service
