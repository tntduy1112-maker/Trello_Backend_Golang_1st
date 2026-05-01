# TaskFlow — Backend Implementation Plan

> Kế hoạch triển khai Go Backend theo từng phase với chi tiết tasks, dependencies, và acceptance criteria.

---

## Tổng quan

| Metric | Value |
|--------|-------|
| **Total Phases** | 6 |
| **Total Duration** | ~10 weeks |
| **Total Modules** | 12 |
| **Total Handlers** | ~80 |
| **Architecture** | Clean Architecture (Layered) |

---

## Project Structure

```
taskflow-backend/
├── cmd/
│   └── server/
│       └── main.go                 # Entry point
│
├── internal/
│   ├── config/
│   │   └── config.go               # Viper configuration
│   │
│   ├── domain/                     # Domain models (entities)
│   │   ├── user.go
│   │   ├── organization.go
│   │   ├── board.go
│   │   ├── list.go
│   │   ├── card.go
│   │   └── ...
│   │
│   ├── repository/                 # Data access layer
│   │   ├── interfaces.go           # Repository interfaces
│   │   ├── user_repository.go
│   │   ├── board_repository.go
│   │   └── ...
│   │
│   ├── service/                    # Business logic layer
│   │   ├── auth_service.go
│   │   ├── user_service.go
│   │   ├── board_service.go
│   │   └── ...
│   │
│   ├── handler/                    # HTTP handlers (controllers)
│   │   ├── auth_handler.go
│   │   ├── user_handler.go
│   │   ├── board_handler.go
│   │   └── ...
│   │
│   ├── middleware/
│   │   ├── auth.go                 # JWT authentication
│   │   ├── ratelimit.go            # Rate limiting
│   │   ├── cors.go                 # CORS
│   │   ├── logger.go               # Request logging
│   │   └── error_handler.go        # Global error handler
│   │
│   ├── dto/                        # Data Transfer Objects
│   │   ├── request/
│   │   │   ├── auth_request.go
│   │   │   └── ...
│   │   └── response/
│   │       ├── auth_response.go
│   │       └── ...
│   │
│   └── worker/                     # Background jobs
│       ├── email_worker.go
│       └── notification_worker.go
│
├── pkg/                            # Shared packages
│   ├── apperror/                   # Custom errors
│   │   └── errors.go
│   ├── jwt/                        # JWT utilities
│   │   └── jwt.go
│   ├── hash/                       # Password hashing
│   │   └── bcrypt.go
│   ├── crypto/                     # AES encryption
│   │   └── aes.go
│   ├── cuid/                       # CUID generator
│   │   └── cuid.go
│   ├── validator/                  # Input validation
│   │   └── validator.go
│   ├── email/                      # Email service
│   │   └── email.go
│   ├── storage/                    # MinIO client
│   │   └── minio.go
│   └── position/                   # Position utilities
│       └── position.go
│
├── migrations/                     # Database migrations (Goose)
│   ├── 00001_create_users.sql
│   └── ...
│
├── docs/                           # Swagger documentation
│   └── swagger/
│
├── scripts/
│   └── init-db.sql
│
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
├── docker-compose.yml
└── .env.example
```

---

## Phase 1: Project Setup & Authentication (2 weeks)

### Week 1: Project Foundation

#### Task 1.1: Initialize Go Project
```bash
# Commands
go mod init github.com/tntduy1112-maker/taskflow-backend
```

**Files to create:**
- [ ] `go.mod`, `go.sum`
- [ ] `cmd/server/main.go`
- [ ] `internal/config/config.go`
- [ ] `Makefile` (already exists, update for Go)

**Dependencies:**
```go
// go.mod
require (
    github.com/gin-gonic/gin v1.9.1
    github.com/jackc/pgx/v5 v5.5.0
    github.com/redis/go-redis/v9 v9.3.0
    github.com/golang-jwt/jwt/v5 v5.2.0
    github.com/go-playground/validator/v10 v10.16.0
    github.com/spf13/viper v1.18.0
    github.com/rs/zerolog v1.31.0
    github.com/minio/minio-go/v7 v7.0.66
    github.com/nrednav/cuid2 v1.0.0
    golang.org/x/crypto v0.17.0
    github.com/pressly/goose/v3 v3.17.0
    github.com/hibiken/asynq v0.24.1
)
```

**Acceptance Criteria:**
- [ ] `go build ./...` compiles without errors
- [ ] `go test ./...` runs (even with no tests)
- [ ] Basic health endpoint responds at `/health`

---

#### Task 1.2: Configuration & Environment
**Files:**
- [ ] `internal/config/config.go`
- [ ] `.env.example` (update)

```go
// internal/config/config.go
type Config struct {
    App      AppConfig
    Database DatabaseConfig
    Redis    RedisConfig
    JWT      JWTConfig
    MinIO    MinIOConfig
    SMTP     SMTPConfig
}
```

**Acceptance Criteria:**
- [ ] Config loads from `.env` file
- [ ] Config validates required fields
- [ ] Sensitive values are not logged

---

#### Task 1.3: Database Connection & Migrations
**Files:**
- [ ] `pkg/database/postgres.go`
- [ ] `migrations/00001_create_users.sql`
- [ ] `migrations/00002_create_refresh_tokens.sql`
- [ ] `migrations/00003_create_email_verifications.sql`

**Acceptance Criteria:**
- [ ] Database connects successfully
- [ ] Migrations run with `goose up`
- [ ] Migrations rollback with `goose down`

---

#### Task 1.4: Redis Connection
**Files:**
- [ ] `pkg/cache/redis.go`

**Acceptance Criteria:**
- [ ] Redis connects successfully
- [ ] Basic SET/GET operations work
- [ ] Connection pooling configured

---

#### Task 1.5: Shared Packages
**Files:**
- [ ] `pkg/apperror/errors.go` — Custom error types
- [ ] `pkg/cuid/cuid.go` — CUID generator
- [ ] `pkg/hash/bcrypt.go` — Password hashing
- [ ] `pkg/jwt/jwt.go` — JWT sign/verify
- [ ] `pkg/crypto/aes.go` — AES-256-GCM encryption
- [ ] `pkg/validator/validator.go` — Input validation

**Acceptance Criteria:**
- [ ] Unit tests pass for all packages
- [ ] Password hashing works correctly
- [ ] JWT generation and verification works
- [ ] AES encryption/decryption works

---

### Week 2: Authentication Module

#### Task 1.6: User Domain & Repository
**Files:**
- [ ] `internal/domain/user.go`
- [ ] `internal/repository/interfaces.go`
- [ ] `internal/repository/user_repository.go`

```go
// internal/domain/user.go
type User struct {
    ID               string
    Email            string
    PasswordHash     string
    FullName         string
    AvatarURL        *string
    IsVerified       bool
    IsActive         bool
    TokensValidAfter time.Time
    CreatedAt        time.Time
    UpdatedAt        time.Time
    DeletedAt        *time.Time
}
```

**Acceptance Criteria:**
- [ ] CRUD operations work
- [ ] Soft delete works
- [ ] FindByEmail works

---

#### Task 1.7: Auth Service
**Files:**
- [ ] `internal/service/auth_service.go`
- [ ] `internal/dto/request/auth_request.go`
- [ ] `internal/dto/response/auth_response.go`

**Methods:**
```go
type AuthService interface {
    Register(ctx, req RegisterRequest) (*User, error)
    Login(ctx, req LoginRequest) (*TokenPair, error)
    VerifyEmail(ctx, req VerifyEmailRequest) error
    ResendVerification(ctx, email string) error
    RefreshToken(ctx, refreshToken string) (*TokenPair, error)
    Logout(ctx, accessToken, refreshToken string) error
    LogoutAll(ctx, userID string) error
    ForgotPassword(ctx, email string) error
    ResetPassword(ctx, req ResetPasswordRequest) error
    GetCurrentUser(ctx, userID string) (*User, error)
    UpdateProfile(ctx, userID string, req UpdateProfileRequest) (*User, error)
}
```

**Acceptance Criteria:**
- [ ] Registration creates user + sends OTP
- [ ] Login returns access + refresh tokens
- [ ] Token refresh rotates refresh token
- [ ] Logout blacklists access token
- [ ] Password reset flow works

---

#### Task 1.8: Auth Handler & Routes
**Files:**
- [ ] `internal/handler/auth_handler.go`
- [ ] `internal/middleware/auth.go`
- [ ] `internal/middleware/ratelimit.go`

**Endpoints:**
| Method | Endpoint | Rate Limit | Auth |
|--------|----------|------------|------|
| POST | `/api/v1/auth/register` | 3/hour/IP | No |
| POST | `/api/v1/auth/verify-email` | 5 attempts/OTP | No |
| POST | `/api/v1/auth/resend-verification` | 3/hour/email | No |
| POST | `/api/v1/auth/login` | 5/15min/IP | No |
| POST | `/api/v1/auth/refresh` | 30/15min/user | No |
| POST | `/api/v1/auth/logout` | — | Optional |
| POST | `/api/v1/auth/logout-all` | — | Required |
| POST | `/api/v1/auth/forgot-password` | 5/15min/IP | No |
| POST | `/api/v1/auth/reset-password` | 5/15min/IP | No |
| GET | `/api/v1/auth/me` | — | Required |
| PUT | `/api/v1/auth/me` | — | Required |

**Acceptance Criteria:**
- [ ] All endpoints return correct status codes
- [ ] Rate limiting works
- [ ] JWT middleware validates tokens
- [ ] Redis blacklist blocks revoked tokens

---

#### Task 1.9: Email Service
**Files:**
- [ ] `pkg/email/email.go`
- [ ] `pkg/email/templates/` (optional)

**Methods:**
```go
type EmailService interface {
    SendVerificationEmail(to, otp string) error
    SendPasswordResetEmail(to, token string) error
    SendBoardInvitation(to, inviterName, boardName, token string) error
}
```

**Acceptance Criteria:**
- [ ] Emails send via SMTP
- [ ] Works with MailHog in development
- [ ] Non-blocking (fire-and-forget or queue)

---

#### Task 1.10: MinIO Storage
**Files:**
- [ ] `pkg/storage/minio.go`

**Methods:**
```go
type StorageService interface {
    Upload(ctx, folder, filename string, data io.Reader, size int64, contentType string) (string, error)
    Delete(ctx, objectKey string) error
    GetPublicURL(objectKey string) string
}
```

**Acceptance Criteria:**
- [ ] Upload works
- [ ] Delete works
- [ ] Public URLs are accessible

---

### Phase 1 Deliverables

- [ ] Go project compiles and runs
- [ ] Database migrations work
- [ ] All auth endpoints functional
- [ ] JWT + refresh token rotation works
- [ ] Rate limiting works
- [ ] Email sending works
- [ ] Avatar upload works
- [ ] Unit tests for auth service (80% coverage)

---

## Phase 2: Organization & Board (1.5 weeks)

### Task 2.1: Organization Module
**Files:**
- [ ] `internal/domain/organization.go`
- [ ] `internal/repository/organization_repository.go`
- [ ] `internal/service/organization_service.go`
- [ ] `internal/handler/organization_handler.go`
- [ ] `migrations/00004_create_organizations.sql`
- [ ] `migrations/00005_create_organization_members.sql`

**Endpoints:**
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/organizations` | List user's organizations |
| POST | `/api/v1/organizations` | Create organization |
| GET | `/api/v1/organizations/:slug` | Get organization details |
| PUT | `/api/v1/organizations/:slug` | Update organization |
| DELETE | `/api/v1/organizations/:slug` | Delete organization |
| GET | `/api/v1/organizations/:slug/members` | List members |
| POST | `/api/v1/organizations/:slug/members` | Invite member |
| PUT | `/api/v1/organizations/:slug/members/:userId` | Update member role |
| DELETE | `/api/v1/organizations/:slug/members/:userId` | Remove member |

**Acceptance Criteria:**
- [ ] CRUD operations work
- [ ] Slug is unique and URL-friendly
- [ ] Role-based permissions enforced
- [ ] Owner cannot be removed

---

### Task 2.2: Board Module
**Files:**
- [ ] `internal/domain/board.go`
- [ ] `internal/repository/board_repository.go`
- [ ] `internal/service/board_service.go`
- [ ] `internal/handler/board_handler.go`
- [ ] `migrations/00006_create_boards.sql`
- [ ] `migrations/00007_create_board_members.sql`
- [ ] `migrations/00008_create_board_invitations.sql`

**Endpoints:**
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/organizations/:slug/boards` | List boards in org |
| POST | `/api/v1/organizations/:slug/boards` | Create board |
| GET | `/api/v1/boards/:id` | Get board with lists & cards |
| PUT | `/api/v1/boards/:id` | Update board |
| DELETE | `/api/v1/boards/:id` | Delete board |
| POST | `/api/v1/boards/:id/close` | Close board |
| POST | `/api/v1/boards/:id/reopen` | Reopen board |
| GET | `/api/v1/boards/:id/members` | List board members |
| POST | `/api/v1/boards/:id/invite` | Invite to board |
| PUT | `/api/v1/boards/:id/members/:userId` | Update member role |
| DELETE | `/api/v1/boards/:id/members/:userId` | Remove member |

**Acceptance Criteria:**
- [ ] Board visibility works (private/workspace/public)
- [ ] 4-level permissions enforced
- [ ] Board invitation flow works
- [ ] Soft delete (close) works

---

### Task 2.3: Invitation Module
**Files:**
- [ ] `internal/handler/invitation_handler.go`

**Endpoints:**
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/invitations/:token` | Get invitation details |
| POST | `/api/v1/invitations/:token/accept` | Accept invitation |
| POST | `/api/v1/invitations/:token/decline` | Decline invitation |

**Acceptance Criteria:**
- [ ] Token-based invitation works
- [ ] Expired invitations rejected
- [ ] New user can register via invitation

---

### Phase 2 Deliverables

- [ ] Organization CRUD works
- [ ] Board CRUD works
- [ ] Role-based permissions work
- [ ] Invitation flow works
- [ ] Unit tests (80% coverage)

---

## Phase 3: Lists, Cards, Labels (1.5 weeks)

### Task 3.1: List Module
**Files:**
- [ ] `internal/domain/list.go`
- [ ] `internal/repository/list_repository.go`
- [ ] `internal/service/list_service.go`
- [ ] `internal/handler/list_handler.go`
- [ ] `migrations/00009_create_lists.sql`

**Endpoints:**
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/boards/:boardId/lists` | Create list |
| PUT | `/api/v1/lists/:id` | Update list title |
| DELETE | `/api/v1/lists/:id` | Archive list |
| PUT | `/api/v1/lists/:id/move` | Reorder list |
| POST | `/api/v1/lists/:id/copy` | Copy list |

**Acceptance Criteria:**
- [ ] Float position ordering works
- [ ] Drag & drop reorder works
- [ ] Position rebalancing works

---

### Task 3.2: Card Module
**Files:**
- [ ] `internal/domain/card.go`
- [ ] `internal/repository/card_repository.go`
- [ ] `internal/service/card_service.go`
- [ ] `internal/handler/card_handler.go`
- [ ] `migrations/00010_create_cards.sql`
- [ ] `migrations/00011_create_card_members.sql`

**Endpoints:**
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/lists/:listId/cards` | Create card |
| GET | `/api/v1/cards/:id` | Get card details |
| PUT | `/api/v1/cards/:id` | Update card |
| DELETE | `/api/v1/cards/:id` | Archive card |
| PUT | `/api/v1/cards/:id/move` | Move card (list + position) |
| POST | `/api/v1/cards/:id/assign` | Assign card |
| DELETE | `/api/v1/cards/:id/assign` | Unassign card |
| POST | `/api/v1/cards/:id/complete` | Mark complete |
| DELETE | `/api/v1/cards/:id/complete` | Mark incomplete |

**Acceptance Criteria:**
- [ ] Card CRUD works
- [ ] Move between lists works
- [ ] Position ordering works
- [ ] Due date works
- [ ] Priority works

---

### Task 3.3: Label Module
**Files:**
- [ ] `internal/domain/label.go`
- [ ] `internal/repository/label_repository.go`
- [ ] `internal/service/label_service.go`
- [ ] `internal/handler/label_handler.go`
- [ ] `migrations/00012_create_labels.sql`
- [ ] `migrations/00013_create_card_labels.sql`

**Endpoints:**
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/boards/:boardId/labels` | List board labels |
| POST | `/api/v1/boards/:boardId/labels` | Create label |
| PUT | `/api/v1/labels/:id` | Update label |
| DELETE | `/api/v1/labels/:id` | Delete label |
| POST | `/api/v1/cards/:cardId/labels/:labelId` | Assign label |
| DELETE | `/api/v1/cards/:cardId/labels/:labelId` | Remove label |

**Acceptance Criteria:**
- [ ] Labels per board work
- [ ] Assign/unassign works
- [ ] Color picker works

---

### Phase 3 Deliverables

- [ ] List CRUD + reorder works
- [ ] Card CRUD + move works
- [ ] Label CRUD + assign works
- [ ] Drag & drop positions persist
- [ ] Unit tests (80% coverage)

---

## Phase 4: Advanced Card Features (2 weeks)

### Task 4.1: Checklist Module
**Files:**
- [ ] `internal/domain/checklist.go`
- [ ] `internal/repository/checklist_repository.go`
- [ ] `internal/service/checklist_service.go`
- [ ] `internal/handler/checklist_handler.go`
- [ ] `migrations/00014_create_checklists.sql`
- [ ] `migrations/00015_create_checklist_items.sql`

**Endpoints:**
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/cards/:cardId/checklists` | Create checklist |
| PUT | `/api/v1/checklists/:id` | Update checklist |
| DELETE | `/api/v1/checklists/:id` | Delete checklist |
| POST | `/api/v1/checklists/:id/items` | Add item |
| PUT | `/api/v1/checklist-items/:id` | Update item |
| DELETE | `/api/v1/checklist-items/:id` | Delete item |
| POST | `/api/v1/checklist-items/:id/complete` | Toggle complete |
| POST | `/api/v1/checklist-items/:id/assign` | Assign item |

**Acceptance Criteria:**
- [ ] Multiple checklists per card
- [ ] Item completion tracking
- [ ] Progress calculation works
- [ ] Item due dates work

---

### Task 4.2: Comment Module
**Files:**
- [ ] `internal/domain/comment.go`
- [ ] `internal/repository/comment_repository.go`
- [ ] `internal/service/comment_service.go`
- [ ] `internal/handler/comment_handler.go`
- [ ] `migrations/00016_create_comments.sql`

**Endpoints:**
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/cards/:cardId/comments` | List comments |
| POST | `/api/v1/cards/:cardId/comments` | Create comment |
| PUT | `/api/v1/comments/:id` | Edit comment |
| DELETE | `/api/v1/comments/:id` | Delete comment |
| POST | `/api/v1/comments/:id/reply` | Reply to comment |

**Acceptance Criteria:**
- [ ] 1-level threading works
- [ ] Edit tracking works
- [ ] Soft delete works

---

### Task 4.3: Attachment Module
**Files:**
- [ ] `internal/domain/attachment.go`
- [ ] `internal/repository/attachment_repository.go`
- [ ] `internal/service/attachment_service.go`
- [ ] `internal/handler/attachment_handler.go`
- [ ] `migrations/00017_create_attachments.sql`

**Endpoints:**
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/cards/:cardId/attachments` | List attachments |
| POST | `/api/v1/cards/:cardId/attachments` | Upload attachment |
| DELETE | `/api/v1/attachments/:id` | Delete attachment |
| GET | `/api/v1/attachments/:id/download` | Download file |
| POST | `/api/v1/attachments/:id/cover` | Set as cover |
| DELETE | `/api/v1/cards/:cardId/cover` | Remove cover |

**Acceptance Criteria:**
- [ ] File upload to MinIO works
- [ ] Download with signed URL works
- [ ] Cover image works
- [ ] File size limit enforced (10MB)

---

### Task 4.4: Activity Log Module
**Files:**
- [ ] `internal/domain/activity.go`
- [ ] `internal/repository/activity_repository.go`
- [ ] `internal/service/activity_service.go`
- [ ] `internal/handler/activity_handler.go`
- [ ] `migrations/00018_create_activity_logs.sql`

**Endpoints:**
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/boards/:boardId/activity` | Board activity |
| GET | `/api/v1/cards/:cardId/activity` | Card activity |
| GET | `/api/v1/activity/:id` | Activity detail |

**Acceptance Criteria:**
- [ ] Auto-logging on all actions
- [ ] JSONB metadata works
- [ ] Human-readable descriptions
- [ ] Pagination works

---

### Phase 4 Deliverables

- [ ] Checklists work
- [ ] Comments work
- [ ] Attachments work
- [ ] Activity logging works
- [ ] Unit tests (80% coverage)

---

## Phase 5: Notifications & Real-time (1 week)

### Task 5.1: Notification Module
**Files:**
- [ ] `internal/domain/notification.go`
- [ ] `internal/repository/notification_repository.go`
- [ ] `internal/service/notification_service.go`
- [ ] `internal/handler/notification_handler.go`
- [ ] `migrations/00019_create_notifications.sql`

**Endpoints:**
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/notifications` | List notifications |
| GET | `/api/v1/notifications/unread-count` | Unread count |
| POST | `/api/v1/notifications/:id/read` | Mark as read |
| POST | `/api/v1/notifications/read-all` | Mark all as read |
| DELETE | `/api/v1/notifications/:id` | Delete notification |
| GET | `/api/v1/notifications/stream` | SSE stream |

**Acceptance Criteria:**
- [ ] Notifications created on triggers
- [ ] Unread count works
- [ ] Mark read works

---

### Task 5.2: SSE (Server-Sent Events)
**Files:**
- [ ] `internal/handler/sse_handler.go`
- [ ] `pkg/sse/broker.go`

**Events:**
```go
type SSEEvent struct {
    Type    string      `json:"type"`
    Payload interface{} `json:"payload"`
}

// Event types:
// - notification:new
// - card:updated
// - card:moved
// - comment:added
// - activity:new
```

**Acceptance Criteria:**
- [ ] SSE connection works
- [ ] Events broadcast to correct users
- [ ] Connection tracking in Redis
- [ ] Reconnection works

---

### Task 5.3: Background Workers
**Files:**
- [ ] `internal/worker/email_worker.go`
- [ ] `internal/worker/notification_worker.go`
- [ ] `internal/worker/reminder_worker.go`

**Jobs:**
- [ ] Email sending (async)
- [ ] Notification creation
- [ ] Due date reminders (cron)

**Acceptance Criteria:**
- [ ] Jobs process correctly
- [ ] Retry on failure
- [ ] Dead letter queue works

---

### Phase 5 Deliverables

- [ ] Notifications work
- [ ] SSE real-time works
- [ ] Background jobs work
- [ ] Due date reminders work

---

## Phase 6: Polish & Production (1.5 weeks)

### Task 6.1: API Documentation
**Files:**
- [ ] `docs/swagger/swagger.json`
- [ ] Swagger annotations on all handlers

**Acceptance Criteria:**
- [ ] Swagger UI accessible
- [ ] All endpoints documented
- [ ] Request/response schemas complete

---

### Task 6.2: Testing
**Coverage Targets:**
- [ ] Unit tests: 80%
- [ ] Integration tests: critical paths
- [ ] E2E tests: happy paths

**Files:**
- [ ] `*_test.go` for all packages
- [ ] `tests/integration/` directory

---

### Task 6.3: Performance Optimization
- [ ] Database query optimization
- [ ] Redis caching strategy
- [ ] Connection pooling tuned
- [ ] N+1 query prevention

---

### Task 6.4: Security Hardening
- [ ] Security headers (Helmet equivalent)
- [ ] Input sanitization
- [ ] SQL injection prevention (parameterized)
- [ ] Rate limiting all endpoints
- [ ] Audit logging

---

### Task 6.5: Production Readiness
- [ ] Graceful shutdown
- [ ] Health check endpoint
- [ ] Readiness/liveness probes
- [ ] Prometheus metrics
- [ ] Structured logging (Zerolog)

---

### Phase 6 Deliverables

- [ ] API documentation complete
- [ ] Test coverage targets met
- [ ] Performance optimized
- [ ] Security hardened
- [ ] Production-ready Docker image

---

## Dependencies Matrix

```
Phase 1 (Auth)
    │
    ├── Phase 2 (Org/Board) ─── depends on auth
    │       │
    │       └── Phase 3 (List/Card/Label) ─── depends on board
    │               │
    │               └── Phase 4 (Checklist/Comment/Attachment) ─── depends on card
    │                       │
    │                       └── Phase 5 (Notifications/SSE) ─── depends on all
    │
    └── Phase 6 (Polish) ─── runs parallel with Phase 5
```

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| SSE scaling | Use Redis pub/sub for multi-instance |
| Database bottleneck | Read replicas, connection pooling |
| File storage | MinIO with CDN in production |
| Rate limit bypass | Redis-based distributed rate limiting |
| Token theft | Short expiry + rotation + blacklist |

---

## Quick Start Checklist

```bash
# 1. Initialize project
mkdir taskflow-backend && cd taskflow-backend
go mod init github.com/tntduy1112-maker/taskflow-backend

# 2. Install dependencies
go get github.com/gin-gonic/gin
go get github.com/jackc/pgx/v5
go get github.com/redis/go-redis/v9
# ... (see full list above)

# 3. Create project structure
mkdir -p cmd/server internal/{config,domain,repository,service,handler,middleware,dto,worker} pkg/{apperror,jwt,hash,crypto,cuid,validator,email,storage,position} migrations

# 4. Start infrastructure
docker compose up -d postgres redis minio mailhog

# 5. Run migrations
goose -dir migrations postgres "$DATABASE_URL" up

# 6. Start development
air -c .air.toml
```
