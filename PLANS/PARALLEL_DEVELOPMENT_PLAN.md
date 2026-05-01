# TaskFlow — Parallel Development Plan

> Kế hoạch phát triển song song Frontend + Backend với các milestone tích hợp để end-user có thể test từng feature.

---

## Tổng quan

```
Week    1    2    3    4    5    6    7    8    9    10
        ├────┴────┼────┴────┼────┴────┼────┴────┼────┴────┤
        │ Sprint 1│ Sprint 2│ Sprint 3│ Sprint 4│ Sprint 5│
        │         │         │         │         │         │
Backend │████ Auth│██ Org/Board █ List/Card █ Advanced█ Notif │
        │         │         │         │         │         │
Frontend│████ Auth│██ Workspace█ Kanban  █ Card Det█ Notif │
        │         │         │         │         │         │
        ▼         ▼         ▼         ▼         ▼         ▼
      Demo 1   Demo 2    Demo 3    Demo 4    Demo 5    Release
      (Auth)   (Boards)  (Kanban)  (Full)    (RT)      (Prod)
```

---

## Nguyên tắc phát triển song song

### 1. API Contract First
- Backend định nghĩa API contract (OpenAPI/Swagger) trước
- Frontend mock API responses để phát triển độc lập
- Integration khi cả hai hoàn thành

### 2. Vertical Slices
- Mỗi sprint hoàn thành một feature end-to-end
- User có thể test ngay sau mỗi sprint

### 3. Feature Flags
- Features mới có thể bật/tắt
- Cho phép deploy liên tục

---

## Sprint 1: Authentication (Week 1-2)

### Mục tiêu
User có thể đăng ký, xác thực email, đăng nhập, và xem profile.

### Backend Tasks

| Task | Priority | Endpoint | Description |
|------|:--------:|----------|-------------|
| B1.1 | P0 | — | Project setup, config, database |
| B1.2 | P0 | `POST /auth/register` | Đăng ký + gửi OTP |
| B1.3 | P0 | `POST /auth/verify-email` | Xác thực email |
| B1.4 | P0 | `POST /auth/login` | Đăng nhập |
| B1.5 | P0 | `POST /auth/refresh` | Refresh token |
| B1.6 | P0 | `POST /auth/logout` | Đăng xuất |
| B1.7 | P1 | `POST /auth/forgot-password` | Quên mật khẩu |
| B1.8 | P1 | `POST /auth/reset-password` | Đặt lại mật khẩu |
| B1.9 | P0 | `GET /auth/me` | Lấy thông tin user |
| B1.10 | P1 | `PUT /auth/me` | Cập nhật profile + avatar |

### Frontend Tasks

| Task | Priority | Page/Component | Description |
|------|:--------:|----------------|-------------|
| F1.1 | P0 | `axiosInstance.js` | Axios config + JWT interceptor |
| F1.2 | P0 | `authSlice.js` | Redux slice cho auth |
| F1.3 | P0 | `LoginPage.jsx` | Trang đăng nhập |
| F1.4 | P0 | `RegisterPage.jsx` | Trang đăng ký |
| F1.5 | P0 | `VerifyEmailPage.jsx` | Trang xác thực OTP |
| F1.6 | P1 | `ForgotPasswordPage.jsx` | Trang quên mật khẩu |
| F1.7 | P1 | `ResetPasswordPage.jsx` | Trang đặt lại mật khẩu |
| F1.8 | P0 | `AuthLayout.jsx` | Layout cho auth pages |
| F1.9 | P0 | `AppLayout.jsx` | Layout cho app (có Navbar) |
| F1.10 | P0 | Routing | Protected routes setup |

### API Contract (Sprint 1)

```yaml
# POST /api/v1/auth/register
Request:
  email: string (required, email format)
  password: string (required, min 6)
  full_name: string (required, 1-255)
Response 201:
  success: true
  data:
    id: string
    email: string
    full_name: string
    is_verified: false
  message: "Check email for verification code"

# POST /api/v1/auth/login
Request:
  email: string
  password: string
Response 200:
  success: true
  data:
    access_token: string
    user:
      id: string
      email: string
      full_name: string
      avatar_url: string | null
      is_verified: boolean
Set-Cookie: refreshToken=<encrypted>; HttpOnly; SameSite=Strict

# GET /api/v1/auth/me
Headers: Authorization: Bearer <token>
Response 200:
  success: true
  data:
    id: string
    email: string
    full_name: string
    avatar_url: string | null
    is_verified: boolean
    created_at: string
```

### Demo 1 Checklist
- [ ] User đăng ký thành công
- [ ] User nhận email OTP
- [ ] User xác thực email
- [ ] User đăng nhập
- [ ] Token refresh tự động khi hết hạn
- [ ] User đăng xuất
- [ ] Protected routes redirect đến login

---

## Sprint 2: Workspace & Board (Week 3-4)

### Mục tiêu
User có thể tạo workspace, mời members, tạo boards.

### Backend Tasks

| Task | Priority | Endpoint | Description |
|------|:--------:|----------|-------------|
| B2.1 | P0 | `GET /organizations` | List workspaces |
| B2.2 | P0 | `POST /organizations` | Tạo workspace |
| B2.3 | P0 | `GET /organizations/:slug` | Chi tiết workspace |
| B2.4 | P1 | `PUT /organizations/:slug` | Cập nhật workspace |
| B2.5 | P1 | `DELETE /organizations/:slug` | Xóa workspace |
| B2.6 | P0 | `GET /organizations/:slug/members` | List members |
| B2.7 | P1 | `POST /organizations/:slug/members` | Mời member |
| B2.8 | P0 | `GET /organizations/:slug/boards` | List boards |
| B2.9 | P0 | `POST /organizations/:slug/boards` | Tạo board |
| B2.10 | P0 | `GET /boards/:id` | Chi tiết board |
| B2.11 | P1 | `PUT /boards/:id` | Cập nhật board |

### Frontend Tasks

| Task | Priority | Page/Component | Description |
|------|:--------:|----------------|-------------|
| F2.1 | P0 | `workspaceSlice.js` | Redux slice |
| F2.2 | P0 | `boardSlice.js` | Redux slice |
| F2.3 | P0 | `WorkspacesPage.jsx` | List workspaces |
| F2.4 | P0 | `CreateWorkspacePage.jsx` | Tạo workspace |
| F2.5 | P0 | `BoardListPage.jsx` | List boards trong workspace |
| F2.6 | P1 | `WorkspaceSettingsPage.jsx` | Cài đặt workspace |
| F2.7 | P0 | `CreateBoardModal.jsx` | Modal tạo board |
| F2.8 | P0 | `Sidebar.jsx` | Sidebar navigation |
| F2.9 | P1 | `InviteMemberModal.jsx` | Modal mời member |

### API Contract (Sprint 2)

```yaml
# GET /api/v1/organizations
Response 200:
  success: true
  data:
    - id: string
      name: string
      slug: string
      logo_url: string | null
      role: "owner" | "admin" | "member"
      boards_count: number

# POST /api/v1/organizations
Request:
  name: string (required)
  description: string (optional)
Response 201:
  success: true
  data:
    id: string
    name: string
    slug: string (auto-generated)

# GET /api/v1/organizations/:slug/boards
Response 200:
  success: true
  data:
    - id: string
      title: string
      background_color: string
      visibility: "private" | "workspace" | "public"
      is_closed: boolean
      lists_count: number
      cards_count: number

# GET /api/v1/boards/:id
Response 200:
  success: true
  data:
    id: string
    title: string
    description: string | null
    background_color: string
    visibility: string
    lists: []  # Empty for now, Sprint 3
    labels: [] # Empty for now
    members:
      - id: string
        full_name: string
        avatar_url: string | null
        role: "owner" | "admin" | "member" | "viewer"
```

### Demo 2 Checklist
- [ ] User tạo workspace
- [ ] User thấy list workspaces
- [ ] User tạo board trong workspace
- [ ] User thấy list boards
- [ ] User mở board (empty state)
- [ ] Sidebar navigation hoạt động

---

## Sprint 3: Kanban Board (Week 5-6)

### Mục tiêu
User có thể tạo lists, cards, và drag & drop.

### Backend Tasks

| Task | Priority | Endpoint | Description |
|------|:--------:|----------|-------------|
| B3.1 | P0 | `POST /boards/:id/lists` | Tạo list |
| B3.2 | P0 | `PUT /lists/:id` | Cập nhật list |
| B3.3 | P0 | `PUT /lists/:id/move` | Di chuyển list |
| B3.4 | P0 | `DELETE /lists/:id` | Archive list |
| B3.5 | P0 | `POST /lists/:id/cards` | Tạo card |
| B3.6 | P0 | `GET /cards/:id` | Chi tiết card |
| B3.7 | P0 | `PUT /cards/:id` | Cập nhật card |
| B3.8 | P0 | `PUT /cards/:id/move` | Di chuyển card |
| B3.9 | P0 | `DELETE /cards/:id` | Archive card |
| B3.10 | P0 | `GET /boards/:id/labels` | List labels |
| B3.11 | P0 | `POST /boards/:id/labels` | Tạo label |
| B3.12 | P0 | `POST /cards/:id/labels/:labelId` | Gán label |

### Frontend Tasks

| Task | Priority | Page/Component | Description |
|------|:--------:|----------------|-------------|
| F3.1 | P0 | `BoardPage.jsx` | Trang Kanban board |
| F3.2 | P0 | `ListColumn.jsx` | Component list |
| F3.3 | P0 | `CardItem.jsx` | Component card |
| F3.4 | P0 | DnD Setup | @dnd-kit configuration |
| F3.5 | P0 | List DnD | Drag & drop lists |
| F3.6 | P0 | Card DnD | Drag & drop cards |
| F3.7 | P0 | Add List | Inline add list |
| F3.8 | P0 | Add Card | Inline add card |
| F3.9 | P1 | `CardDetailModal.jsx` | Modal chi tiết card (basic) |
| F3.10 | P1 | Labels UI | Hiển thị labels trên card |

### API Contract (Sprint 3)

```yaml
# GET /api/v1/boards/:id (updated)
Response 200:
  success: true
  data:
    id: string
    title: string
    lists:
      - id: string
        title: string
        position: number
        cards:
          - id: string
            title: string
            position: number
            description: string | null
            priority: "none" | "low" | "medium" | "high"
            due_date: string | null
            is_completed: boolean
            assignee:
              id: string
              full_name: string
              avatar_url: string | null
            labels:
              - id: string
                name: string
                color: string
            comments_count: number
            attachments_count: number
            checklists_progress:
              completed: number
              total: number
    labels:
      - id: string
        name: string
        color: string

# PUT /api/v1/cards/:id/move
Request:
  list_id: string
  position: number
Response 200:
  success: true
  data:
    id: string
    list_id: string
    position: number

# PUT /api/v1/lists/:id/move
Request:
  position: number
Response 200:
  success: true
  data:
    id: string
    position: number
```

### Demo 3 Checklist
- [ ] User tạo list
- [ ] User tạo card
- [ ] User drag & drop card giữa lists
- [ ] User drag & drop list
- [ ] Position persist sau refresh
- [ ] User mở card detail modal (basic info)
- [ ] User tạo/gán labels

---

## Sprint 4: Card Details (Week 7-8)

### Mục tiêu
User có thể edit card details, comments, checklists, attachments.

### Backend Tasks

| Task | Priority | Endpoint | Description |
|------|:--------:|----------|-------------|
| B4.1 | P0 | `POST /cards/:id/assign` | Assign card |
| B4.2 | P0 | `POST /cards/:id/complete` | Toggle complete |
| B4.3 | P0 | `GET /cards/:id/comments` | List comments |
| B4.4 | P0 | `POST /cards/:id/comments` | Tạo comment |
| B4.5 | P1 | `PUT /comments/:id` | Edit comment |
| B4.6 | P1 | `DELETE /comments/:id` | Delete comment |
| B4.7 | P0 | `POST /cards/:id/checklists` | Tạo checklist |
| B4.8 | P0 | `POST /checklists/:id/items` | Tạo item |
| B4.9 | P0 | `POST /checklist-items/:id/complete` | Toggle item |
| B4.10 | P0 | `POST /cards/:id/attachments` | Upload file |
| B4.11 | P0 | `POST /attachments/:id/cover` | Set cover |
| B4.12 | P0 | `GET /cards/:id/activity` | Activity log |

### Frontend Tasks

| Task | Priority | Page/Component | Description |
|------|:--------:|----------------|-------------|
| F4.1 | P0 | Card Modal | Full card detail modal |
| F4.2 | P0 | Title/Desc | Inline edit title, description |
| F4.3 | P0 | Due Date | Date picker |
| F4.4 | P0 | Priority | Priority selector |
| F4.5 | P0 | Assignee | Assignee picker |
| F4.6 | P0 | Labels | Label picker |
| F4.7 | P0 | Comments | Comment list + add |
| F4.8 | P0 | Checklists | Checklist UI |
| F4.9 | P0 | Attachments | Upload + list |
| F4.10 | P0 | Cover | Set cover image |
| F4.11 | P1 | Activity | Activity feed |

### API Contract (Sprint 4)

```yaml
# GET /api/v1/cards/:id (full detail)
Response 200:
  success: true
  data:
    id: string
    title: string
    description: string | null
    position: number
    priority: string
    due_date: string | null
    is_completed: boolean
    cover_url: string | null
    list:
      id: string
      title: string
    assignee:
      id: string
      full_name: string
      avatar_url: string | null
    labels: [...]
    checklists:
      - id: string
        title: string
        position: number
        items:
          - id: string
            title: string
            is_completed: boolean
            assignee: {...} | null
            due_date: string | null
    comments:
      - id: string
        content: string
        author:
          id: string
          full_name: string
          avatar_url: string | null
        created_at: string
        is_edited: boolean
    attachments:
      - id: string
        filename: string
        url: string
        mime_type: string
        file_size: number
        is_cover: boolean
        created_at: string
    activity:
      - id: string
        action: string
        description: string
        user:
          id: string
          full_name: string
        created_at: string
        metadata: object
```

### Demo 4 Checklist
- [ ] User edit title, description
- [ ] User set due date, priority
- [ ] User assign card to member
- [ ] User mark card complete
- [ ] User add/toggle checklists
- [ ] User add comments
- [ ] User upload attachments
- [ ] User set cover image
- [ ] User xem activity log

---

## Sprint 5: Notifications & Real-time (Week 9-10)

### Mục tiêu
User nhận notifications real-time, live activity updates.

### Backend Tasks

| Task | Priority | Endpoint | Description |
|------|:--------:|----------|-------------|
| B5.1 | P0 | `GET /notifications` | List notifications |
| B5.2 | P0 | `GET /notifications/unread-count` | Unread count |
| B5.3 | P0 | `POST /notifications/:id/read` | Mark read |
| B5.4 | P0 | `POST /notifications/read-all` | Mark all read |
| B5.5 | P0 | `GET /notifications/stream` | SSE stream |
| B5.6 | P0 | SSE Events | Broadcast on actions |
| B5.7 | P1 | Due Reminders | Background job |

### Frontend Tasks

| Task | Priority | Page/Component | Description |
|------|:--------:|----------------|-------------|
| F5.1 | P0 | `notificationSlice.js` | Redux slice |
| F5.2 | P0 | `NotificationDropdown.jsx` | Dropdown UI |
| F5.3 | P0 | `useNotificationStream.js` | SSE hook |
| F5.4 | P0 | Unread Badge | Badge trên bell icon |
| F5.5 | P0 | Live Activity | Live update trong card |
| F5.6 | P1 | Optimistic Updates | Optimistic UI |
| F5.7 | P1 | `ProfilePage.jsx` | Profile page |
| F5.8 | P1 | `HelpDrawer.jsx` | Help drawer |

### API Contract (Sprint 5)

```yaml
# GET /api/v1/notifications
Response 200:
  success: true
  data:
    - id: string
      type: "card_assigned" | "comment_added" | ...
      title: string
      message: string
      is_read: boolean
      created_at: string
      board:
        id: string
        title: string
      card:
        id: string
        title: string
      actor:
        id: string
        full_name: string
        avatar_url: string | null

# GET /api/v1/notifications/stream (SSE)
Content-Type: text/event-stream

event: notification
data: {"type": "card_assigned", "card_id": "...", ...}

event: card_updated
data: {"card_id": "...", "changes": {...}}

event: activity
data: {"card_id": "...", "activity": {...}}
```

### Demo 5 Checklist
- [ ] User nhận notification khi được assign
- [ ] User nhận notification khi có comment
- [ ] Notification badge hiển thị unread count
- [ ] User click notification → navigate đến card
- [ ] Activity log live update
- [ ] Card changes reflect real-time (multi-user)

---

## Integration Timeline

```
Week 1-2 (Sprint 1):
├── Day 1-3:  Backend project setup + DB
├── Day 4-7:  Backend auth endpoints
├── Day 8-10: Frontend auth pages
├── Day 11-14: Integration + Testing
└── Demo 1: Auth flow complete ✓

Week 3-4 (Sprint 2):
├── Day 1-4:  Backend org/board endpoints
├── Day 5-8:  Frontend workspace/board pages
├── Day 9-12: Integration + Testing
└── Demo 2: Workspace/Board flow ✓

Week 5-6 (Sprint 3):
├── Day 1-5:  Backend list/card/label endpoints
├── Day 6-10: Frontend Kanban board + DnD
├── Day 11-14: Integration + Testing
└── Demo 3: Kanban board working ✓

Week 7-8 (Sprint 4):
├── Day 1-6:  Backend advanced card features
├── Day 7-12: Frontend card detail modal
├── Day 13-14: Integration + Testing
└── Demo 4: Full card features ✓

Week 9-10 (Sprint 5):
├── Day 1-5:  Backend notifications + SSE
├── Day 6-10: Frontend notifications + real-time
├── Day 11-12: Integration + Testing
├── Day 13-14: Polish + Bug fixes
└── Demo 5: Real-time working ✓
```

---

## Mock API Strategy

Để Frontend có thể phát triển song song khi Backend chưa sẵn sàng:

### Option 1: MSW (Mock Service Worker)

```javascript
// src/mocks/handlers.js
import { rest } from 'msw';

export const handlers = [
  rest.post('/api/v1/auth/login', (req, res, ctx) => {
    return res(
      ctx.json({
        success: true,
        data: {
          access_token: 'mock-token',
          user: {
            id: 'user-1',
            email: 'test@example.com',
            full_name: 'Test User',
          },
        },
      })
    );
  }),
  // ... more handlers
];
```

### Option 2: JSON Server

```bash
# db.json with mock data
npx json-server --watch db.json --port 3001
```

### Option 3: Axios Interceptor Mock

```javascript
// src/api/mock.js
if (import.meta.env.VITE_USE_MOCK === 'true') {
  axiosInstance.interceptors.request.use((config) => {
    // Return mock response based on URL
  });
}
```

---

## Communication Protocol

### Daily Sync
- 15 min standup
- Backend shares: API changes, blockers
- Frontend shares: Integration issues, UI changes

### API Changes
1. Backend updates Swagger spec
2. Notify Frontend via Slack/Discord
3. Frontend updates API calls
4. Test integration

### Git Workflow
```
main
├── develop
│   ├── feature/be-auth
│   ├── feature/fe-auth
│   ├── feature/be-board
│   └── feature/fe-board
```

### PR Convention
```
feat(be): implement auth endpoints
feat(fe): implement login page
fix(be): token refresh rotation
fix(fe): axios interceptor retry
```

---

## Definition of Done

### Per Sprint
- [ ] All P0 tasks completed
- [ ] API contract documented
- [ ] Frontend-Backend integrated
- [ ] No critical bugs
- [ ] Demo recorded/presented

### Per Feature
- [ ] Backend endpoint working
- [ ] Frontend UI working
- [ ] Integration tested
- [ ] Error handling done
- [ ] Loading states done

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| API changes break FE | Version API, communicate early |
| FE blocked on BE | Use mock APIs |
| BE blocked on FE | Test with Postman/curl |
| Integration issues | Daily sync, shared staging |
| Scope creep | Strict P0/P1 prioritization |

---

## Quick Reference

### Backend Start
```bash
cd taskflow-backend
make dev
# API: http://localhost:8080
```

### Frontend Start
```bash
cd Frontend
npm run dev
# App: http://localhost:5173
```

### Full Stack
```bash
# Terminal 1: Infrastructure
docker compose up -d postgres redis minio mailhog

# Terminal 2: Backend
cd taskflow-backend && make watch

# Terminal 3: Frontend
cd Frontend && npm run dev
```

### Swagger UI
```
http://localhost:8080/swagger/index.html
```

### MailHog (dev emails)
```
http://localhost:8025
```
