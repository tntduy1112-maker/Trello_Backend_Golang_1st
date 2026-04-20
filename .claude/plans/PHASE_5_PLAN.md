# Phase 5: Notifications & Real-time (2 Weeks)

> Detailed execution plan for notifications system, SSE streaming, due reminders, and real-time updates.

---

## Overview

```
Week 9: Notifications Core     Week 10: Real-time & Polish
┌──────────────────────┐       ┌──────────────────────┐
│ 5.1 Notifications DB │       │ 5.3 SSE Streaming    │
│ 5.2 Notification Svc │──────▶│ 5.4 Due Reminders    │
│     + Triggers       │       │ 5.5 Integration      │
└──────────────────────┘       └──────────────────────┘
```

---

## Database Migration (Sprint 5)

### Migration 00019: notifications

```sql
-- migrations/00019_create_notifications.sql

-- +goose Up
CREATE TYPE notification_type AS ENUM (
    'card_assigned',
    'card_due_soon',
    'card_overdue',
    'comment_added',
    'comment_reply',
    'mentioned',
    'board_invitation',
    'checklist_item_assigned',
    'card_completed',
    'member_added_to_board'
);

CREATE TABLE notifications (
    id              VARCHAR(25) PRIMARY KEY,
    user_id         VARCHAR(25) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type            notification_type NOT NULL,
    title           VARCHAR(255) NOT NULL,
    message         TEXT,
    board_id        VARCHAR(25) REFERENCES boards(id) ON DELETE CASCADE,
    card_id         VARCHAR(25) REFERENCES cards(id) ON DELETE CASCADE,
    actor_id        VARCHAR(25) REFERENCES users(id) ON DELETE SET NULL,
    is_read         BOOLEAN DEFAULT FALSE,
    read_at         TIMESTAMPTZ,
    metadata        JSONB DEFAULT '{}',
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_notifications_user ON notifications(user_id);
CREATE INDEX idx_notifications_unread ON notifications(user_id, is_read) WHERE is_read = FALSE;
CREATE INDEX idx_notifications_created ON notifications(created_at DESC);
CREATE INDEX idx_notifications_type ON notifications(type);
CREATE INDEX idx_notifications_board ON notifications(board_id);
CREATE INDEX idx_notifications_card ON notifications(card_id);

-- +goose Down
DROP TABLE IF EXISTS notifications;
DROP TYPE IF EXISTS notification_type;
```

---

## Redis Keys for Real-time

| Key Pattern | Value | TTL | Purpose |
|-------------|-------|-----|---------|
| `sse:user:{user_id}` | JSON connection info | 30 min | Track active SSE connections |
| `sse:board:{board_id}:users` | SET of user_ids | 30 min | Users watching a board |
| `notification:pending:{user_id}` | LIST of notifications | 1 hour | Queue for offline users |
| `due_reminder:processed:{card_id}` | "1" | 25 hours | Prevent duplicate reminders |

---

## Task 5.1: Notifications Domain & Repository

### Domain Model

```go
// internal/domain/notification.go
package domain

import "time"

type NotificationType string

const (
    NotificationCardAssigned         NotificationType = "card_assigned"
    NotificationCardDueSoon          NotificationType = "card_due_soon"
    NotificationCardOverdue          NotificationType = "card_overdue"
    NotificationCommentAdded         NotificationType = "comment_added"
    NotificationCommentReply         NotificationType = "comment_reply"
    NotificationMentioned            NotificationType = "mentioned"
    NotificationBoardInvitation      NotificationType = "board_invitation"
    NotificationChecklistItemAssigned NotificationType = "checklist_item_assigned"
    NotificationCardCompleted        NotificationType = "card_completed"
    NotificationMemberAddedToBoard   NotificationType = "member_added_to_board"
)

type Notification struct {
    ID        string                 `json:"id"`
    UserID    string                 `json:"user_id"`
    Type      NotificationType       `json:"type"`
    Title     string                 `json:"title"`
    Message   string                 `json:"message,omitempty"`
    BoardID   *string                `json:"board_id,omitempty"`
    CardID    *string                `json:"card_id,omitempty"`
    ActorID   *string                `json:"actor_id,omitempty"`
    IsRead    bool                   `json:"is_read"`
    ReadAt    *time.Time             `json:"read_at,omitempty"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
    CreatedAt time.Time              `json:"created_at"`
    
    // Populated by joins
    Actor     *User                  `json:"actor,omitempty"`
    Board     *Board                 `json:"board,omitempty"`
    Card      *Card                  `json:"card,omitempty"`
}

// NotificationPreferences (future: user settings)
type NotificationPreferences struct {
    UserID              string `json:"user_id"`
    EmailCardAssigned   bool   `json:"email_card_assigned"`
    EmailComments       bool   `json:"email_comments"`
    EmailDueReminders   bool   `json:"email_due_reminders"`
    PushEnabled         bool   `json:"push_enabled"`
}
```

### Repository Interface

```go
// internal/repository/notification_repository.go
type NotificationRepository interface {
    // CRUD
    Create(ctx context.Context, notification *domain.Notification) error
    CreateBatch(ctx context.Context, notifications []domain.Notification) error
    FindByID(ctx context.Context, id string) (*domain.Notification, error)
    
    // Queries
    FindByUser(ctx context.Context, userID string, opts NotificationQueryOpts) ([]domain.Notification, int, error)
    GetUnreadCount(ctx context.Context, userID string) (int, error)
    
    // Status updates
    MarkAsRead(ctx context.Context, id string) error
    MarkAllAsRead(ctx context.Context, userID string) error
    MarkMultipleAsRead(ctx context.Context, ids []string) error
    
    // Cleanup
    DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)
    DeleteByUser(ctx context.Context, userID string) error
}

type NotificationQueryOpts struct {
    Page      int
    Limit     int
    UnreadOnly bool
    Types     []domain.NotificationType
}
```

### Implementation

```go
// internal/repository/postgres/notification_repository.go
func (r *notificationRepository) FindByUser(
    ctx context.Context,
    userID string,
    opts NotificationQueryOpts,
) ([]domain.Notification, int, error) {
    query := `
        SELECT 
            n.id, n.user_id, n.type, n.title, n.message,
            n.board_id, n.card_id, n.actor_id, n.is_read,
            n.read_at, n.metadata, n.created_at,
            -- Actor info
            u.id as actor_id, u.full_name as actor_name, u.avatar_url as actor_avatar,
            -- Board info
            b.id as board_id, b.title as board_title,
            -- Card info
            c.id as card_id, c.title as card_title
        FROM notifications n
        LEFT JOIN users u ON n.actor_id = u.id
        LEFT JOIN boards b ON n.board_id = b.id
        LEFT JOIN cards c ON n.card_id = c.id
        WHERE n.user_id = $1
    `
    
    args := []interface{}{userID}
    argIndex := 2
    
    if opts.UnreadOnly {
        query += fmt.Sprintf(" AND n.is_read = FALSE")
    }
    
    if len(opts.Types) > 0 {
        query += fmt.Sprintf(" AND n.type = ANY($%d)", argIndex)
        args = append(args, opts.Types)
        argIndex++
    }
    
    query += " ORDER BY n.created_at DESC"
    query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
    args = append(args, opts.Limit, (opts.Page-1)*opts.Limit)
    
    // Execute and scan...
}
```

---

## Task 5.2: Notification Service & Triggers

### Service Interface

```go
// internal/service/notification_service.go
type NotificationService interface {
    // Query
    List(ctx context.Context, userID string, opts ListNotificationOpts) (*PaginatedNotificationsResponse, error)
    GetUnreadCount(ctx context.Context, userID string) (int, error)
    
    // Status
    MarkAsRead(ctx context.Context, notificationID string) error
    MarkAllAsRead(ctx context.Context, userID string) error
    
    // Create notifications (called by other services)
    NotifyCardAssigned(ctx context.Context, card *domain.Card, assignee *domain.User, actor *domain.User) error
    NotifyCommentAdded(ctx context.Context, comment *domain.Comment, card *domain.Card, actor *domain.User) error
    NotifyMentioned(ctx context.Context, mentionedUserIDs []string, card *domain.Card, actor *domain.User) error
    NotifyBoardInvitation(ctx context.Context, invitation *domain.BoardInvitation) error
    NotifyChecklistItemAssigned(ctx context.Context, item *domain.ChecklistItem, card *domain.Card, actor *domain.User) error
    NotifyCardDueSoon(ctx context.Context, card *domain.Card) error
    NotifyCardOverdue(ctx context.Context, card *domain.Card) error
    
    // Broadcast to SSE (internal)
    BroadcastToUser(ctx context.Context, userID string, notification *domain.Notification) error
}
```

### Notification Templates

```go
// internal/service/notification_templates.go
var notificationTemplates = map[domain.NotificationType]struct {
    TitleTemplate   string
    MessageTemplate string
}{
    domain.NotificationCardAssigned: {
        TitleTemplate:   "You were assigned to a card",
        MessageTemplate: "%s assigned you to \"%s\"",
    },
    domain.NotificationCommentAdded: {
        TitleTemplate:   "New comment on your card",
        MessageTemplate: "%s commented on \"%s\"",
    },
    domain.NotificationCommentReply: {
        TitleTemplate:   "Reply to your comment",
        MessageTemplate: "%s replied to your comment on \"%s\"",
    },
    domain.NotificationMentioned: {
        TitleTemplate:   "You were mentioned",
        MessageTemplate: "%s mentioned you in \"%s\"",
    },
    domain.NotificationCardDueSoon: {
        TitleTemplate:   "Card due soon",
        MessageTemplate: "\"%s\" is due in %s",
    },
    domain.NotificationCardOverdue: {
        TitleTemplate:   "Card overdue",
        MessageTemplate: "\"%s\" is overdue",
    },
    domain.NotificationBoardInvitation: {
        TitleTemplate:   "Board invitation",
        MessageTemplate: "%s invited you to join \"%s\"",
    },
    domain.NotificationChecklistItemAssigned: {
        TitleTemplate:   "Checklist item assigned",
        MessageTemplate: "%s assigned you to \"%s\" on \"%s\"",
    },
    domain.NotificationCardCompleted: {
        TitleTemplate:   "Card completed",
        MessageTemplate: "%s marked \"%s\" as complete",
    },
    domain.NotificationMemberAddedToBoard: {
        TitleTemplate:   "Added to board",
        MessageTemplate: "%s added you to \"%s\"",
    },
}
```

### Implementation

```go
// internal/service/notification_service_impl.go
func (s *notificationService) NotifyCardAssigned(
    ctx context.Context,
    card *domain.Card,
    assignee *domain.User,
    actor *domain.User,
) error {
    // Don't notify if self-assigned
    if assignee.ID == actor.ID {
        return nil
    }
    
    notification := &domain.Notification{
        ID:       cuid.New(),
        UserID:   assignee.ID,
        Type:     domain.NotificationCardAssigned,
        Title:    "You were assigned to a card",
        Message:  fmt.Sprintf("%s assigned you to \"%s\"", actor.FullName, card.Title),
        BoardID:  &card.BoardID,
        CardID:   &card.ID,
        ActorID:  &actor.ID,
        Metadata: map[string]interface{}{
            "card_title":  card.Title,
            "actor_name":  actor.FullName,
            "board_title": card.BoardTitle, // denormalized for display
        },
        CreatedAt: time.Now(),
    }
    
    if err := s.repo.Create(ctx, notification); err != nil {
        return err
    }
    
    // Broadcast via SSE
    return s.BroadcastToUser(ctx, assignee.ID, notification)
}

// Mention parsing helper
func (s *notificationService) NotifyMentioned(
    ctx context.Context,
    mentionedUserIDs []string,
    card *domain.Card,
    actor *domain.User,
) error {
    var notifications []domain.Notification
    
    for _, userID := range mentionedUserIDs {
        if userID == actor.ID {
            continue // Don't notify self-mentions
        }
        
        notifications = append(notifications, domain.Notification{
            ID:       cuid.New(),
            UserID:   userID,
            Type:     domain.NotificationMentioned,
            Title:    "You were mentioned",
            Message:  fmt.Sprintf("%s mentioned you in \"%s\"", actor.FullName, card.Title),
            BoardID:  &card.BoardID,
            CardID:   &card.ID,
            ActorID:  &actor.ID,
            CreatedAt: time.Now(),
        })
    }
    
    if len(notifications) == 0 {
        return nil
    }
    
    if err := s.repo.CreateBatch(ctx, notifications); err != nil {
        return err
    }
    
    // Broadcast to each user
    for i := range notifications {
        _ = s.BroadcastToUser(ctx, notifications[i].UserID, &notifications[i])
    }
    
    return nil
}
```

### Integration Points

```go
// Integrate into existing services

// internal/service/card_service.go
func (s *cardService) Assign(ctx context.Context, cardID, assigneeID string) error {
    card, err := s.cardRepo.FindByID(ctx, cardID)
    if err != nil {
        return err
    }
    
    assignee, err := s.userRepo.FindByID(ctx, assigneeID)
    if err != nil {
        return err
    }
    
    // Update card
    card.AssigneeID = &assigneeID
    if err := s.cardRepo.Update(ctx, card); err != nil {
        return err
    }
    
    // Log activity
    s.activitySvc.Log(ctx, LogActivityRequest{
        BoardID:  card.BoardID,
        CardID:   &card.ID,
        UserID:   ctx.Value("user_id").(string),
        Action:   domain.ActivityCardAssigned,
        Metadata: map[string]interface{}{"assignee_name": assignee.FullName},
    })
    
    // Send notification
    actor := ctx.Value("user").(*domain.User)
    s.notificationSvc.NotifyCardAssigned(ctx, card, assignee, actor)
    
    return nil
}

// internal/service/comment_service.go
func (s *commentService) Create(ctx context.Context, req CreateCommentRequest) (*CommentResponse, error) {
    // ... create comment ...
    
    // Parse @mentions from content
    mentionedUserIDs := s.parseMentions(req.Content)
    
    // Notify card watchers/assignee
    if card.AssigneeID != nil && *card.AssigneeID != actor.ID {
        s.notificationSvc.NotifyCommentAdded(ctx, comment, card, actor)
    }
    
    // Notify mentioned users
    if len(mentionedUserIDs) > 0 {
        s.notificationSvc.NotifyMentioned(ctx, mentionedUserIDs, card, actor)
    }
    
    return response, nil
}
```

### DTOs

```go
// internal/dto/request/notification_request.go
type ListNotificationOpts struct {
    Page       int                      `form:"page" validate:"min=1"`
    Limit      int                      `form:"limit" validate:"min=1,max=100"`
    UnreadOnly bool                     `form:"unread_only"`
    Types      []domain.NotificationType `form:"types"`
}

// internal/dto/response/notification_response.go
type NotificationResponse struct {
    ID        string                 `json:"id"`
    Type      string                 `json:"type"`
    Title     string                 `json:"title"`
    Message   string                 `json:"message,omitempty"`
    IsRead    bool                   `json:"is_read"`
    CreatedAt time.Time              `json:"created_at"`
    Actor     *UserBriefResponse     `json:"actor,omitempty"`
    Board     *BoardBriefResponse    `json:"board,omitempty"`
    Card      *CardBriefResponse     `json:"card,omitempty"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type PaginatedNotificationsResponse struct {
    Notifications []NotificationResponse `json:"notifications"`
    Pagination    PaginationMeta         `json:"pagination"`
    UnreadCount   int                    `json:"unread_count"`
}

type UnreadCountResponse struct {
    Count int `json:"count"`
}
```

### API Endpoints

| Method | Endpoint | Description | Auth | Rate Limit |
|--------|----------|-------------|:----:|:----------:|
| GET | `/notifications` | List notifications | Required | 60/min |
| GET | `/notifications/unread-count` | Get unread count | Required | 120/min |
| POST | `/notifications/:id/read` | Mark as read | Required | 120/min |
| POST | `/notifications/read-all` | Mark all as read | Required | 10/min |
| DELETE | `/notifications/:id` | Delete notification | Required | 60/min |

---

## Task 5.3: SSE Streaming

### SSE Handler

```go
// internal/handler/sse_handler.go
type SSEHandler struct {
    notificationSvc service.NotificationService
    boardSvc        service.BoardService
    redisClient     *redis.Client
    logger          *zerolog.Logger
}

func (h *SSEHandler) Stream(c *gin.Context) {
    userID := c.GetString("user_id")
    
    // Set SSE headers
    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")
    c.Header("X-Accel-Buffering", "no") // Disable nginx buffering
    
    // Create client channel
    clientChan := make(chan SSEEvent, 100)
    clientID := cuid.New()
    
    // Register client
    h.registerClient(c.Request.Context(), userID, clientID, clientChan)
    defer h.unregisterClient(userID, clientID)
    
    // Send initial connection event
    c.SSEvent("connected", map[string]string{"client_id": clientID})
    c.Writer.Flush()
    
    // Keep-alive ticker (every 30s)
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    ctx := c.Request.Context()
    
    for {
        select {
        case <-ctx.Done():
            return
            
        case event := <-clientChan:
            c.SSEvent(event.Type, event.Data)
            c.Writer.Flush()
            
        case <-ticker.C:
            c.SSEvent("ping", map[string]int64{"timestamp": time.Now().Unix()})
            c.Writer.Flush()
        }
    }
}

type SSEEvent struct {
    Type string      `json:"type"`
    Data interface{} `json:"data"`
}
```

### SSE Client Manager

```go
// internal/service/sse_manager.go
type SSEManager struct {
    clients     map[string]map[string]chan SSEEvent // userID -> clientID -> channel
    boardUsers  map[string]map[string]bool          // boardID -> userID -> true
    mu          sync.RWMutex
    redisClient *redis.Client
}

func NewSSEManager(redisClient *redis.Client) *SSEManager {
    return &SSEManager{
        clients:     make(map[string]map[string]chan SSEEvent),
        boardUsers:  make(map[string]map[string]bool),
        redisClient: redisClient,
    }
}

func (m *SSEManager) Register(userID, clientID string, ch chan SSEEvent) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    if m.clients[userID] == nil {
        m.clients[userID] = make(map[string]chan SSEEvent)
    }
    m.clients[userID][clientID] = ch
    
    // Track in Redis for distributed deployment
    m.redisClient.SAdd(context.Background(), 
        fmt.Sprintf("sse:user:%s:clients", userID), 
        clientID,
    )
    m.redisClient.Expire(context.Background(),
        fmt.Sprintf("sse:user:%s:clients", userID),
        30*time.Minute,
    )
}

func (m *SSEManager) Unregister(userID, clientID string) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    if clients, ok := m.clients[userID]; ok {
        if ch, ok := clients[clientID]; ok {
            close(ch)
            delete(clients, clientID)
        }
        if len(clients) == 0 {
            delete(m.clients, userID)
        }
    }
    
    m.redisClient.SRem(context.Background(),
        fmt.Sprintf("sse:user:%s:clients", userID),
        clientID,
    )
}

func (m *SSEManager) SendToUser(userID string, event SSEEvent) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    if clients, ok := m.clients[userID]; ok {
        for _, ch := range clients {
            select {
            case ch <- event:
            default:
                // Channel full, skip (client is slow)
            }
        }
    }
}

func (m *SSEManager) BroadcastToBoard(boardID string, event SSEEvent, excludeUserID string) {
    m.mu.RLock()
    users := m.boardUsers[boardID]
    m.mu.RUnlock()
    
    for userID := range users {
        if userID != excludeUserID {
            m.SendToUser(userID, event)
        }
    }
}

// Track users viewing a board
func (m *SSEManager) JoinBoard(userID, boardID string) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    if m.boardUsers[boardID] == nil {
        m.boardUsers[boardID] = make(map[string]bool)
    }
    m.boardUsers[boardID][userID] = true
    
    // Redis for distributed
    m.redisClient.SAdd(context.Background(),
        fmt.Sprintf("sse:board:%s:users", boardID),
        userID,
    )
}

func (m *SSEManager) LeaveBoard(userID, boardID string) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    if users, ok := m.boardUsers[boardID]; ok {
        delete(users, userID)
        if len(users) == 0 {
            delete(m.boardUsers, boardID)
        }
    }
    
    m.redisClient.SRem(context.Background(),
        fmt.Sprintf("sse:board:%s:users", boardID),
        userID,
    )
}
```

### SSE Event Types

```go
// internal/service/sse_events.go
const (
    // Notification events
    SSEEventNotification = "notification"
    
    // Board events (real-time collaboration)
    SSEEventCardCreated   = "card_created"
    SSEEventCardUpdated   = "card_updated"
    SSEEventCardMoved     = "card_moved"
    SSEEventCardDeleted   = "card_deleted"
    SSEEventListCreated   = "list_created"
    SSEEventListUpdated   = "list_updated"
    SSEEventListMoved     = "list_moved"
    SSEEventListDeleted   = "list_deleted"
    SSEEventCommentAdded  = "comment_added"
    SSEEventActivityAdded = "activity_added"
    
    // Presence events
    SSEEventUserJoined    = "user_joined"
    SSEEventUserLeft      = "user_left"
)

// Event payloads
type CardEventPayload struct {
    CardID   string      `json:"card_id"`
    ListID   string      `json:"list_id"`
    BoardID  string      `json:"board_id"`
    Card     interface{} `json:"card,omitempty"`
    Changes  interface{} `json:"changes,omitempty"`
    Actor    interface{} `json:"actor"`
}

type ListEventPayload struct {
    ListID  string      `json:"list_id"`
    BoardID string      `json:"board_id"`
    List    interface{} `json:"list,omitempty"`
    Changes interface{} `json:"changes,omitempty"`
    Actor   interface{} `json:"actor"`
}
```

### Real-time Updates Integration

```go
// internal/service/card_service.go
func (s *cardService) Move(ctx context.Context, cardID string, req MoveCardRequest) error {
    // ... move card logic ...
    
    // Broadcast to board viewers
    s.sseManager.BroadcastToBoard(card.BoardID, SSEEvent{
        Type: SSEEventCardMoved,
        Data: CardEventPayload{
            CardID:  cardID,
            ListID:  req.ListID,
            BoardID: card.BoardID,
            Changes: map[string]interface{}{
                "from_list_id": oldListID,
                "to_list_id":   req.ListID,
                "position":     req.Position,
            },
            Actor: UserBriefResponse{
                ID:       actor.ID,
                FullName: actor.FullName,
                AvatarURL: actor.AvatarURL,
            },
        },
    }, actor.ID) // Exclude the actor
    
    return nil
}
```

### API Endpoint

| Method | Endpoint | Description | Auth | Notes |
|--------|----------|-------------|:----:|:------|
| GET | `/notifications/stream` | SSE stream | Required | Long-lived connection |
| POST | `/boards/:id/join` | Join board for real-time | Required | Track presence |
| POST | `/boards/:id/leave` | Leave board | Required | Stop presence |

---

## Task 5.4: Due Reminders (Background Job)

### Background Worker

```go
// cmd/worker/main.go
package main

import (
    "context"
    "time"
    
    "github.com/hibiken/asynq"
)

func main() {
    // Redis connection for Asynq
    redisOpt := asynq.RedisClientOpt{Addr: cfg.Redis.Addr}
    
    srv := asynq.NewServer(redisOpt, asynq.Config{
        Concurrency: 10,
        Queues: map[string]int{
            "critical": 6,
            "default":  3,
            "low":      1,
        },
    })
    
    mux := asynq.NewServeMux()
    mux.HandleFunc(tasks.TypeDueReminder, tasks.HandleDueReminder)
    mux.HandleFunc(tasks.TypeOverdueCheck, tasks.HandleOverdueCheck)
    
    srv.Run(mux)
}
```

### Task Definitions

```go
// internal/tasks/due_reminder.go
package tasks

import (
    "context"
    "encoding/json"
    "time"
    
    "github.com/hibiken/asynq"
)

const (
    TypeDueReminder = "due:reminder"
    TypeOverdueCheck = "due:overdue"
)

type DueReminderPayload struct {
    CardID string `json:"card_id"`
}

func NewDueReminderTask(cardID string) (*asynq.Task, error) {
    payload, _ := json.Marshal(DueReminderPayload{CardID: cardID})
    return asynq.NewTask(TypeDueReminder, payload), nil
}

func HandleDueReminder(ctx context.Context, t *asynq.Task) error {
    var payload DueReminderPayload
    if err := json.Unmarshal(t.Payload(), &payload); err != nil {
        return err
    }
    
    // Get services from context
    cardRepo := ctx.Value("cardRepo").(repository.CardRepository)
    notificationSvc := ctx.Value("notificationSvc").(service.NotificationService)
    redisClient := ctx.Value("redis").(*redis.Client)
    
    // Check if already processed (idempotency)
    key := fmt.Sprintf("due_reminder:processed:%s", payload.CardID)
    exists, _ := redisClient.Exists(ctx, key).Result()
    if exists > 0 {
        return nil // Already sent reminder
    }
    
    card, err := cardRepo.FindByID(ctx, payload.CardID)
    if err != nil {
        return err
    }
    
    // Card may have been deleted, completed, or due date changed
    if card == nil || card.IsCompleted || card.DueDate == nil {
        return nil
    }
    
    // Check if still due soon (within 24 hours)
    if time.Until(*card.DueDate) > 24*time.Hour {
        return nil // Due date was extended
    }
    
    // Send notification to assignee
    if card.AssigneeID != nil {
        notificationSvc.NotifyCardDueSoon(ctx, card)
    }
    
    // Mark as processed
    redisClient.Set(ctx, key, "1", 25*time.Hour)
    
    return nil
}
```

### Scheduler

```go
// cmd/scheduler/main.go
package main

import (
    "time"
    
    "github.com/hibiken/asynq"
)

func main() {
    scheduler := asynq.NewScheduler(redisOpt, nil)
    
    // Check for cards due soon every hour
    scheduler.Register("0 * * * *", asynq.NewTask(tasks.TypeDueSoonScan, nil))
    
    // Check for overdue cards every 15 minutes
    scheduler.Register("*/15 * * * *", asynq.NewTask(tasks.TypeOverdueScan, nil))
    
    scheduler.Run()
}

// internal/tasks/due_scanner.go
func HandleDueSoonScan(ctx context.Context, t *asynq.Task) error {
    cardRepo := ctx.Value("cardRepo").(repository.CardRepository)
    client := ctx.Value("asynqClient").(*asynq.Client)
    
    // Find cards due in next 24 hours that haven't been reminded
    cutoff := time.Now().Add(24 * time.Hour)
    cards, err := cardRepo.FindDueBefore(ctx, cutoff)
    if err != nil {
        return err
    }
    
    for _, card := range cards {
        task, _ := NewDueReminderTask(card.ID)
        client.Enqueue(task, asynq.Queue("default"))
    }
    
    return nil
}

func HandleOverdueScan(ctx context.Context, t *asynq.Task) error {
    cardRepo := ctx.Value("cardRepo").(repository.CardRepository)
    notificationSvc := ctx.Value("notificationSvc").(service.NotificationService)
    
    // Find overdue cards
    cards, err := cardRepo.FindOverdue(ctx)
    if err != nil {
        return err
    }
    
    for _, card := range cards {
        if card.AssigneeID != nil {
            notificationSvc.NotifyCardOverdue(ctx, &card)
        }
    }
    
    return nil
}
```

### Repository Methods for Due Dates

```go
// internal/repository/card_repository.go
type CardRepository interface {
    // ... existing methods ...
    
    // Due date queries
    FindDueBefore(ctx context.Context, before time.Time) ([]domain.Card, error)
    FindOverdue(ctx context.Context) ([]domain.Card, error)
    FindDueBetween(ctx context.Context, start, end time.Time) ([]domain.Card, error)
}

// SQL queries
const (
    findDueSoonQuery = `
        SELECT * FROM cards
        WHERE due_date IS NOT NULL
          AND due_date <= $1
          AND due_date > NOW()
          AND is_completed = FALSE
          AND is_archived = FALSE
          AND assignee_id IS NOT NULL
    `
    
    findOverdueQuery = `
        SELECT * FROM cards
        WHERE due_date IS NOT NULL
          AND due_date < NOW()
          AND is_completed = FALSE
          AND is_archived = FALSE
          AND assignee_id IS NOT NULL
    `
)
```

---

## Task 5.5: Integration & Testing

### Handler Registration

```go
// internal/handler/routes.go
func SetupRoutes(r *gin.Engine, h *Handlers) {
    // ... existing routes ...
    
    // Notifications
    notifications := r.Group("/api/v1/notifications")
    notifications.Use(h.authMiddleware.RequireAuth())
    {
        notifications.GET("", h.notification.List)
        notifications.GET("/unread-count", h.notification.GetUnreadCount)
        notifications.GET("/stream", h.sse.Stream)
        notifications.POST("/:id/read", h.notification.MarkAsRead)
        notifications.POST("/read-all", h.notification.MarkAllAsRead)
        notifications.DELETE("/:id", h.notification.Delete)
    }
    
    // Board presence (for real-time)
    boards := r.Group("/api/v1/boards")
    boards.Use(h.authMiddleware.RequireAuth())
    {
        // ... existing board routes ...
        boards.POST("/:id/join", h.board.JoinRealtime)
        boards.POST("/:id/leave", h.board.LeaveRealtime)
    }
}
```

### Frontend SSE Integration Example

```javascript
// Frontend: useNotificationStream.js
import { useEffect, useRef, useCallback } from 'react';
import { useDispatch } from 'react-redux';
import { addNotification, incrementUnread } from '../store/notificationSlice';

export function useNotificationStream() {
    const eventSourceRef = useRef(null);
    const dispatch = useDispatch();
    
    const connect = useCallback(() => {
        const token = localStorage.getItem('accessToken');
        const url = `${API_URL}/notifications/stream`;
        
        eventSourceRef.current = new EventSource(url, {
            headers: { Authorization: `Bearer ${token}` }
        });
        
        eventSourceRef.current.addEventListener('notification', (event) => {
            const notification = JSON.parse(event.data);
            dispatch(addNotification(notification));
            dispatch(incrementUnread());
            
            // Show toast
            showToast(notification.title, notification.message);
        });
        
        eventSourceRef.current.addEventListener('card_updated', (event) => {
            const data = JSON.parse(event.data);
            // Update card in store
            dispatch(updateCard(data.card_id, data.changes));
        });
        
        eventSourceRef.current.addEventListener('ping', () => {
            // Keep-alive, no action needed
        });
        
        eventSourceRef.current.onerror = () => {
            eventSourceRef.current.close();
            // Reconnect after 5 seconds
            setTimeout(connect, 5000);
        };
    }, [dispatch]);
    
    useEffect(() => {
        connect();
        return () => eventSourceRef.current?.close();
    }, [connect]);
}
```

---

## Dependency Graph

```
Migration 00019 (notifications)
    │
    ├──▶ Task 5.1 (Notification Domain & Repo)
    │        │
    │        ▼
    └──▶ Task 5.2 (Notification Service)
              │
              ├── Integrate with Card Service
              ├── Integrate with Comment Service
              ├── Integrate with Checklist Service
              │
              ▼
         Task 5.3 (SSE Streaming)
              │
              ├── SSE Handler
              ├── SSE Manager (client tracking)
              └── Real-time board updates
              │
              ▼
         Task 5.4 (Due Reminders)
              │
              ├── Asynq Worker
              ├── Scheduler (cron)
              └── Due date scanner
              │
              ▼
         Task 5.5 (Integration)
              │
              ├── Route registration
              ├── End-to-end testing
              └── Frontend integration
```

---

## Recommended Execution Order

| Day | Tasks | Focus |
|:---:|-------|-------|
| 1 | Migration 19, Task 5.1 | Notifications table + domain |
| 2 | Task 5.1 | Repository implementation |
| 3 | Task 5.2 | Notification service + templates |
| 4 | Task 5.2 | Integration with Card/Comment services |
| 5 | Task 5.2 | Handlers + API endpoints |
| 6 | Task 5.3 | SSE handler + client manager |
| 7 | Task 5.3 | Real-time board updates |
| 8 | Task 5.4 | Asynq worker + scheduler |
| 9 | Task 5.4 | Due reminder tasks |
| 10 | Task 5.5 | Integration testing + polish |

---

## Phase 5 Deliverables Checklist

### Notifications API
- [ ] `GET /notifications` → 200 + paginated list
- [ ] `GET /notifications/unread-count` → 200 + count
- [ ] `POST /notifications/:id/read` → 200
- [ ] `POST /notifications/read-all` → 200
- [ ] `DELETE /notifications/:id` → 200

### Notification Triggers
- [ ] Card assigned → notification to assignee
- [ ] Comment added → notification to card watchers
- [ ] @mention → notification to mentioned users
- [ ] Board invitation → notification to invitee
- [ ] Checklist item assigned → notification
- [ ] Due soon (24h) → notification
- [ ] Overdue → notification

### SSE Streaming
- [ ] `GET /notifications/stream` → SSE connection
- [ ] Notifications pushed in real-time
- [ ] Keep-alive ping every 30s
- [ ] Reconnection handling
- [ ] Board join/leave tracking

### Real-time Updates
- [ ] Card created/updated/moved → broadcast to board
- [ ] List created/updated/moved → broadcast to board
- [ ] Comment added → broadcast to card viewers
- [ ] Activity logged → broadcast

### Background Jobs
- [ ] Due reminder scanner runs hourly
- [ ] Overdue checker runs every 15 min
- [ ] Idempotent (no duplicate reminders)
- [ ] Handles deleted/completed cards

### Performance
- [ ] SSE scales with multiple clients
- [ ] Redis used for distributed SSE
- [ ] Notification queries optimized (indexes)
- [ ] Background jobs don't block main app

### Tests
- [ ] Unit tests for notification service
- [ ] Integration tests for SSE
- [ ] Load test for multiple SSE clients
- [ ] End-to-end notification flow

---

## Production Considerations

### SSE Scaling

```yaml
# For production with multiple instances
# Use Redis Pub/Sub to broadcast across instances

# docker-compose.prod.yml
services:
  api-1:
    environment:
      - INSTANCE_ID=api-1
  api-2:
    environment:
      - INSTANCE_ID=api-2
  
  # Load balancer with sticky sessions for SSE
  nginx:
    upstream api {
      ip_hash;  # Sticky sessions
      server api-1:8080;
      server api-2:8080;
    }
```

### Notification Cleanup

```go
// Scheduled job: clean old notifications
func CleanupOldNotifications(ctx context.Context) error {
    cutoff := time.Now().AddDate(0, 0, -90) // 90 days
    deleted, err := notificationRepo.DeleteOlderThan(ctx, cutoff)
    log.Info().Int64("deleted", deleted).Msg("Cleaned old notifications")
    return err
}
```

### Rate Limiting SSE

```go
// Limit SSE connections per user
const maxSSEConnectionsPerUser = 5

func (m *SSEManager) Register(userID, clientID string, ch chan SSEEvent) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    if len(m.clients[userID]) >= maxSSEConnectionsPerUser {
        return ErrTooManyConnections
    }
    
    // ... register client
}
```
