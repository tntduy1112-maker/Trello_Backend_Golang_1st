# TaskFlow — Trello-Style Kanban Board

<div align="center">
  <img src="https://res.cloudinary.com/ecommerce2021/image/upload/v1768626951/dev_efjbzw.jpg" alt="Code Web Khong Kho" width="120" style="border-radius: 50%"/>

  <h3>Full-Stack Kanban Board Application</h3>
  <p>Built with Go + React | Multi-tenant Workspaces | Real-time Collaboration</p>

  ![Version](https://img.shields.io/badge/version-1.0.0-blue?style=flat-square)
  ![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)
  ![React](https://img.shields.io/badge/React-18-61DAFB?logo=react&logoColor=white)
  ![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?logo=postgresql&logoColor=white)
  ![Redis](https://img.shields.io/badge/Redis-7-DC382D?logo=redis&logoColor=white)

  [![Facebook](https://img.shields.io/badge/Facebook-Code%20Web%20Khong%20Kho-1877F2?logo=facebook)](https://www.facebook.com/codewebkhongkho)
  [![TikTok](https://img.shields.io/badge/TikTok-@code.web.khng.kh-000000?logo=tiktok)](https://www.tiktok.com/@code.web.khng.kh)
  [![Website](https://img.shields.io/badge/Website-codewebkhongkho.com-FF6B35?logo=google-chrome)](https://codewebkhongkho.com/portfolios)
</div>

---

## Overview

**TaskFlow** is a collaborative project management application inspired by Trello. It enables teams to organize work using boards, lists, and cards with real-time collaboration features.

### Key Features

- **Multi-tenant Workspaces** — Organize teams and projects
- **Kanban Boards** — Visual project management with drag & drop
- **Real-time Collaboration** — Instant updates via Server-Sent Events (SSE)
- **Role-based Access** — Owner, Admin, Member, Viewer permissions
- **Rich Card Details** — Labels, checklists, attachments, comments, due dates
- **Activity Tracking** — Full audit log of all actions
- **Invitation System** — Secure team onboarding via email invitations

---

## Tech Stack

### Backend (Go)

| Technology | Purpose |
|------------|---------|
| **Go 1.25** | Runtime |
| **Gin** | HTTP framework |
| **pgx/v5** | PostgreSQL driver |
| **go-redis** | Redis client |
| **minio-go** | Object storage |
| **golang-jwt** | JWT authentication |
| **zerolog** | Structured logging |
| **validator/v10** | Request validation |

### Frontend (React)

| Technology | Purpose |
|------------|---------|
| **React 18** | UI framework |
| **Vite** | Build tool |
| **Redux Toolkit** | State management |
| **React Router v7** | Routing |
| **@dnd-kit** | Drag & drop |
| **Tailwind CSS** | Styling |
| **Axios** | HTTP client |
| **Lucide React** | Icons |

### Infrastructure

| Technology | Purpose |
|------------|---------|
| **PostgreSQL 16** | Primary database |
| **Redis** | Cache, rate limiting, SSE tracking |
| **MinIO** | S3-compatible file storage |
| **Docker** | Containerization |

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Browser (React SPA :5173)                    │
│               REST/JSON + SSE (Real-time Events)                │
└───────────────────────────────┬─────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Go API Server (:8080)                       │
├─────────────────────────────────────────────────────────────────┤
│  Middleware: CORS | JWT Auth | Rate Limit | Error Handler       │
├─────────────────────────────────────────────────────────────────┤
│  12 Modules: Auth | Organizations | Boards | Lists | Cards      │
│              Labels | Comments | Checklists | Attachments        │
│              Activity | Notifications | Invitations             │
├─────────────────────────────────────────────────────────────────┤
│  Architecture: Handler → Service → Repository                   │
└───────────────────────────────┬─────────────────────────────────┘
                                │
        ┌───────────────────────┼───────────────────────┐
        ▼                       ▼                       ▼
┌──────────────┐       ┌──────────────┐       ┌──────────────┐
│ PostgreSQL   │       │    Redis     │       │    MinIO     │
│   (Data)     │       │   (Cache)    │       │   (Files)    │
└──────────────┘       └──────────────┘       └──────────────┘
```

---

## Data Model

```
Users
└── Organizations (Workspaces)
    ├── organization_members (owner | admin | member)
    └── Boards
        ├── board_members (owner | admin | member | viewer)
        ├── board_invitations
        ├── Labels
        └── Lists
            └── Cards
                ├── card_labels
                ├── Checklists → checklist_items
                ├── Comments
                ├── Attachments (MinIO)
                └── Activity Logs
└── Notifications (SSE real-time)
```

---

## Implemented Features

### Authentication & Security
- [x] JWT access token (15 min) + refresh token (7 days)
- [x] Token rotation with reuse detection
- [x] Redis-based rate limiting
- [x] Password hashing (bcrypt)
- [x] Email verification (OTP)
- [x] Forgot/reset password flow

### Workspaces & Boards
- [x] Multi-tenant workspaces with unique slugs
- [x] Board creation with background colors
- [x] Board visibility (private/workspace/public)
- [x] 4-level permission system
- [x] Email invitation flow

### Lists & Cards
- [x] CRUD operations
- [x] Drag & drop reordering (@dnd-kit)
- [x] Float-based positioning (O(1) insert)
- [x] Archive/restore functionality

### Card Features
- [x] Rich text description
- [x] Due dates with date picker
- [x] Priority levels (none/low/medium/high)
- [x] Single assignee
- [x] Completion status
- [x] Cover image from attachments

### Labels & Organization
- [x] Custom colored labels per board
- [x] Multiple labels per card
- [x] Color picker UI

### Checklists
- [x] Multiple checklists per card
- [x] Checklist items with progress bar
- [x] Check/uncheck functionality

### Comments & Collaboration
- [x] Threaded comments (1-level)
- [x] @mentions with notification
- [x] Edit/delete own comments

### Attachments
- [x] File upload to MinIO
- [x] Image preview
- [x] Set as card cover
- [x] Download functionality

### Real-time & Notifications
- [x] Server-Sent Events (SSE)
- [x] Notification bell with unread count
- [x] Mark as read (single/all)
- [x] Click to navigate

### Activity Tracking
- [x] Card-level activity feed
- [x] Board-level activity log
- [x] JSONB metadata storage

---

## API Endpoints (84 Total)

| Module | Endpoints | Route Prefix |
|--------|:---------:|--------------|
| Auth | 10 | `/api/v1/auth` |
| Organizations | 11 | `/api/v1/organizations` |
| Boards | 12 | `/api/v1/boards` |
| Lists | 6 | `/api/v1/lists` |
| Cards | 14 | `/api/v1/cards` |
| Labels | 5 | `/api/v1/labels` |
| Comments | 3 | `/api/v1/comments` |
| Checklists | 7 | `/api/v1/checklists` |
| Checklist Items | 3 | `/api/v1/checklist-items` |
| Attachments | 4 | `/api/v1/attachments` |
| Notifications | 6 | `/api/v1/notifications` |
| Invitations | 4 | `/api/v1/invitations` |

**API Documentation:** `http://localhost:8080/api/docs`

---

## Quick Start

### Prerequisites

- Go 1.25+
- Node.js 20+
- Docker & Docker Compose
- PostgreSQL 16
- Redis 7
- MinIO

### 1. Clone Repository

```bash
git clone https://github.com/tntduy1112-maker/Trello_Backend_Golang_1st.git
cd Trello_Backend_Golang_1st
```

### 2. Start Infrastructure

```bash
docker-compose up -d postgres redis minio
```

### 3. Backend Setup

```bash
cd Backend

# Copy environment file
cp .env.example .env

# Run migrations
goose -dir migrations postgres "postgres://postgres:secret@localhost:5432/taskflow?sslmode=disable" up

# Start server (with hot reload)
air

# Or without hot reload
go run cmd/server/main.go
```

### 4. Frontend Setup

```bash
cd Frontend

# Install dependencies
npm install

# Start dev server
npm run dev
```

### 5. Access Application

- **Frontend:** http://localhost:5173
- **Backend API:** http://localhost:8080/api/v1
- **API Docs:** http://localhost:8080/api/docs
- **MinIO Console:** http://localhost:9001

---

## Environment Variables

### Backend (.env)

```bash
# App
APP_ENV=development
APP_PORT=8080
FRONTEND_URL=http://localhost:5173

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=taskflow
DB_USER=postgres
DB_PASSWORD=secret

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# JWT
JWT_ACCESS_SECRET=your-access-secret
JWT_REFRESH_SECRET=your-refresh-secret
JWT_ACCESS_EXPIRES_IN=15m
JWT_REFRESH_EXPIRES_IN=168h

# MinIO
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=taskflow

# SMTP
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASS=your-app-password
```

### Frontend (.env)

```bash
VITE_API_URL=http://localhost:8080/api/v1
```

---

## Project Structure

```
.
├── Backend/
│   ├── cmd/server/          # Entry point
│   ├── internal/
│   │   ├── config/          # Configuration
│   │   ├── domain/          # Entity models
│   │   ├── dto/             # Request/Response DTOs
│   │   ├── handler/         # HTTP handlers
│   │   ├── middleware/      # Auth, rate limit, error
│   │   ├── repository/      # Data access layer
│   │   └── service/         # Business logic
│   ├── pkg/                 # Reusable packages
│   ├── migrations/          # SQL migrations (Goose)
│   └── docs/                # Swagger/OpenAPI
│
├── Frontend/
│   ├── src/
│   │   ├── api/             # Axios instance
│   │   ├── components/      # UI components
│   │   ├── hooks/           # Custom hooks
│   │   ├── pages/           # Route pages
│   │   ├── redux/           # Store & slices
│   │   ├── services/        # API services
│   │   └── utils/           # Helpers
│   └── public/
│
├── PLANS/                   # Project documentation
│   ├── PROJECT_DESCRIPTION.md
│   ├── BACKEND_DESCRIPTION.md
│   ├── FRONTEND_DESCRIPTION.md
│   ├── DATABASE_DESIGN.md
│   ├── AUTH_FLOW.md
│   ├── API_SPEC.md
│   └── DELIVERY_NOTE.md
│
└── .claude/                 # AI Agent configuration
    ├── CLAUDE.md
    ├── commands/
    ├── agents/
    ├── rules/
    └── skills/
```

---

## Documentation

| Document | Description |
|----------|-------------|
| [PROJECT_DESCRIPTION.md](PLANS/PROJECT_DESCRIPTION.md) | Project overview & tech decisions |
| [BACKEND_DESCRIPTION.md](PLANS/BACKEND_DESCRIPTION.md) | Go backend architecture |
| [FRONTEND_DESCRIPTION.md](PLANS/FRONTEND_DESCRIPTION.md) | React frontend structure |
| [DATABASE_DESIGN.md](PLANS/DATABASE_DESIGN.md) | PostgreSQL schema (19 tables) |
| [AUTH_FLOW.md](PLANS/AUTH_FLOW.md) | JWT authentication flow |
| [API_SPEC.md](PLANS/API_SPEC.md) | OpenAPI specification |
| [DELIVERY_NOTE.md](PLANS/DELIVERY_NOTE.md) | User guide & feature list |

---

## Development Workflow

This project uses Claude AI agents with structured workflows:

```
/spec  →  /plan  →  /build  →  /test  →  /review  →  /deploy
```

See [.claude/CLAUDE.md](.claude/CLAUDE.md) for AI agent configuration.

---

## Contributing

1. Follow the development workflow
2. Ensure all tests pass
3. Follow conventional commit format
4. Run code review before submitting PR

---

## License

MIT

---

## Author

<div align="center">
  <img src="https://res.cloudinary.com/ecommerce2021/image/upload/v1768626951/dev_efjbzw.jpg" alt="Code Web Khong Kho" width="80" style="border-radius: 50%"/>

  **Code Web Khong Kho**

  | Platform | Link |
  |----------|------|
  | Facebook | [facebook.com/codewebkhongkho](https://www.facebook.com/codewebkhongkho) |
  | TikTok | [@code.web.khng.kh](https://www.tiktok.com/@code.web.khng.kh) |
  | Website | [codewebkhongkho.com](https://codewebkhongkho.com/portfolios) |
</div>

---

<div align="center">
  <sub>Made with care by <a href="https://www.facebook.com/codewebkhongkho">Code Web Khong Kho</a></sub>
</div>
