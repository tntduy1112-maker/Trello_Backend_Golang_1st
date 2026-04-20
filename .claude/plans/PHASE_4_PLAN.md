# Phase 4: Card Details — Comments, Checklists, Attachments (2 Weeks)

> Detailed execution plan for advanced card features: comments, checklists, attachments, and activity logging.

---

## Overview

```
Week 7: Core Features          Week 8: Attachments & Activity
┌──────────────────────┐       ┌──────────────────────┐
│ 4.1 Comments         │       │ 4.3 Attachments      │
│ 4.2 Checklists       │──────▶│ 4.4 Activity Logs    │
│     + Items          │       │ 4.5 Full Card Detail │
└──────────────────────┘       └──────────────────────┘
```

---

## Database Migrations (Sprint 4)

### Migration 00014: comments

```sql
-- migrations/00014_create_comments.sql

-- +goose Up
CREATE TABLE comments (
    id          VARCHAR(25) PRIMARY KEY,
    card_id     VARCHAR(25) NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    author_id   VARCHAR(25) NOT NULL REFERENCES users(id),
    content     TEXT NOT NULL,
    parent_id   VARCHAR(25) REFERENCES comments(id) ON DELETE CASCADE,
    is_edited   BOOLEAN DEFAULT FALSE,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_comments_card ON comments(card_id);
CREATE INDEX idx_comments_author ON comments(author_id);
CREATE INDEX idx_comments_parent ON comments(parent_id);
CREATE INDEX idx_comments_deleted ON comments(deleted_at) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS comments;
```

### Migration 00015: checklists

```sql
-- migrations/00015_create_checklists.sql

-- +goose Up
CREATE TABLE checklists (
    id          VARCHAR(25) PRIMARY KEY,
    card_id     VARCHAR(25) NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    title       VARCHAR(255) NOT NULL,
    position    DOUBLE PRECISION NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_checklists_card ON checklists(card_id);
CREATE INDEX idx_checklists_position ON checklists(card_id, position);

-- +goose Down
DROP TABLE IF EXISTS checklists;
```

### Migration 00016: checklist_items

```sql
-- migrations/00016_create_checklist_items.sql

-- +goose Up
CREATE TABLE checklist_items (
    id              VARCHAR(25) PRIMARY KEY,
    checklist_id    VARCHAR(25) NOT NULL REFERENCES checklists(id) ON DELETE CASCADE,
    title           VARCHAR(500) NOT NULL,
    position        DOUBLE PRECISION NOT NULL,
    is_completed    BOOLEAN DEFAULT FALSE,
    completed_at    TIMESTAMPTZ,
    completed_by    VARCHAR(25) REFERENCES users(id),
    assignee_id     VARCHAR(25) REFERENCES users(id) ON DELETE SET NULL,
    due_date        TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_checklist_items_checklist ON checklist_items(checklist_id);
CREATE INDEX idx_checklist_items_position ON checklist_items(checklist_id, position);
CREATE INDEX idx_checklist_items_assignee ON checklist_items(assignee_id);

-- +goose Down
DROP TABLE IF EXISTS checklist_items;
```

### Migration 00017: attachments

```sql
-- migrations/00017_create_attachments.sql

-- +goose Up
CREATE TABLE attachments (
    id              VARCHAR(25) PRIMARY KEY,
    card_id         VARCHAR(25) NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    uploaded_by     VARCHAR(25) NOT NULL REFERENCES users(id),
    filename        VARCHAR(255) NOT NULL,
    original_name   VARCHAR(255) NOT NULL,
    mime_type       VARCHAR(100) NOT NULL,
    file_size       BIGINT NOT NULL,
    object_key      VARCHAR(500) NOT NULL,
    url             TEXT NOT NULL,
    is_cover        BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_attachments_card ON attachments(card_id);
CREATE INDEX idx_attachments_uploader ON attachments(uploaded_by);
CREATE INDEX idx_attachments_cover ON attachments(card_id, is_cover) WHERE is_cover = TRUE;

-- Add FK to cards for cover_attachment_id
ALTER TABLE cards 
    ADD CONSTRAINT fk_cards_cover_attachment 
    FOREIGN KEY (cover_attachment_id) REFERENCES attachments(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE cards DROP CONSTRAINT IF EXISTS fk_cards_cover_attachment;
DROP TABLE IF EXISTS attachments;
```

### Migration 00018: activity_logs

```sql
-- migrations/00018_create_activity_logs.sql

-- +goose Up
CREATE TYPE activity_action AS ENUM (
    'card_created', 'card_updated', 'card_moved', 'card_archived', 'card_deleted',
    'card_assigned', 'card_unassigned', 'card_completed', 'card_reopened',
    'card_due_date_set', 'card_due_date_removed',
    'list_created', 'list_renamed', 'list_moved', 'list_archived',
    'label_added', 'label_removed',
    'checklist_created', 'checklist_deleted',
    'checklist_item_completed', 'checklist_item_uncompleted',
    'comment_added', 'comment_edited', 'comment_deleted',
    'attachment_added', 'attachment_deleted', 'cover_set', 'cover_removed',
    'member_added', 'member_removed'
);

CREATE TABLE activity_logs (
    id          VARCHAR(25) PRIMARY KEY,
    board_id    VARCHAR(25) NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    card_id     VARCHAR(25) REFERENCES cards(id) ON DELETE SET NULL,
    list_id     VARCHAR(25) REFERENCES lists(id) ON DELETE SET NULL,
    user_id     VARCHAR(25) NOT NULL REFERENCES users(id),
    action      activity_action NOT NULL,
    metadata    JSONB DEFAULT '{}',
    description TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_activity_logs_board ON activity_logs(board_id);
CREATE INDEX idx_activity_logs_card ON activity_logs(card_id);
CREATE INDEX idx_activity_logs_user ON activity_logs(user_id);
CREATE INDEX idx_activity_logs_created ON activity_logs(created_at DESC);
CREATE INDEX idx_activity_logs_metadata ON activity_logs USING GIN (metadata);

-- +goose Down
DROP TABLE IF EXISTS activity_logs;
DROP TYPE IF EXISTS activity_action;
```

---

## Task 4.1: Comments Module

### Domain Model

```go
// internal/domain/comment.go
package domain

import "time"

type Comment struct {
    ID        string     `json:"id"`
    CardID    string     `json:"card_id"`
    AuthorID  string     `json:"author_id"`
    Content   string     `json:"content"`
    ParentID  *string    `json:"parent_id,omitempty"`
    IsEdited  bool       `json:"is_edited"`
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
    DeletedAt *time.Time `json:"-"`
    
    // Populated by joins
    Author    *User      `json:"author,omitempty"`
    Replies   []Comment  `json:"replies,omitempty"`
}
```

### Repository Interface

```go
// internal/repository/comment_repository.go
type CommentRepository interface {
    Create(ctx context.Context, comment *domain.Comment) error
    FindByID(ctx context.Context, id string) (*domain.Comment, error)
    FindByCardID(ctx context.Context, cardID string) ([]domain.Comment, error)
    Update(ctx context.Context, comment *domain.Comment) error
    SoftDelete(ctx context.Context, id string) error
    
    // Validation helpers
    IsAuthor(ctx context.Context, commentID, userID string) (bool, error)
}
```

### Service Methods

```go
// internal/service/comment_service.go
type CommentService interface {
    // List comments for a card (with author info, replies nested)
    ListByCard(ctx context.Context, cardID string) ([]CommentResponse, error)
    
    // Create comment (logs activity, triggers notification)
    Create(ctx context.Context, req CreateCommentRequest) (*CommentResponse, error)
    
    // Update comment (only author, marks is_edited=true)
    Update(ctx context.Context, commentID string, req UpdateCommentRequest) (*CommentResponse, error)
    
    // Soft delete (only author or board admin+)
    Delete(ctx context.Context, commentID string) error
}
```

### DTOs

```go
// internal/dto/request/comment_request.go
type CreateCommentRequest struct {
    CardID   string  `json:"card_id" validate:"required,cuid"`
    Content  string  `json:"content" validate:"required,min=1,max=10000"`
    ParentID *string `json:"parent_id,omitempty" validate:"omitempty,cuid"`
}

type UpdateCommentRequest struct {
    Content string `json:"content" validate:"required,min=1,max=10000"`
}

// internal/dto/response/comment_response.go
type CommentResponse struct {
    ID        string            `json:"id"`
    Content   string            `json:"content"`
    IsEdited  bool              `json:"is_edited"`
    CreatedAt time.Time         `json:"created_at"`
    Author    UserBriefResponse `json:"author"`
    Replies   []CommentResponse `json:"replies,omitempty"`
}
```

### API Endpoints

| Method | Endpoint | Description | Auth | Rate Limit |
|--------|----------|-------------|:----:|:----------:|
| GET | `/cards/:id/comments` | List card comments | Required | 60/min |
| POST | `/cards/:id/comments` | Create comment | Required | 30/min |
| PUT | `/comments/:id` | Update comment | Required (author) | 30/min |
| DELETE | `/comments/:id` | Delete comment | Required (author/admin) | 30/min |

---

## Task 4.2: Checklists Module

### Domain Models

```go
// internal/domain/checklist.go
type Checklist struct {
    ID        string          `json:"id"`
    CardID    string          `json:"card_id"`
    Title     string          `json:"title"`
    Position  float64         `json:"position"`
    CreatedAt time.Time       `json:"created_at"`
    UpdatedAt time.Time       `json:"updated_at"`
    
    Items     []ChecklistItem `json:"items,omitempty"`
}

type ChecklistItem struct {
    ID          string     `json:"id"`
    ChecklistID string     `json:"checklist_id"`
    Title       string     `json:"title"`
    Position    float64    `json:"position"`
    IsCompleted bool       `json:"is_completed"`
    CompletedAt *time.Time `json:"completed_at,omitempty"`
    CompletedBy *string    `json:"completed_by,omitempty"`
    AssigneeID  *string    `json:"assignee_id,omitempty"`
    DueDate     *time.Time `json:"due_date,omitempty"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
    
    Assignee    *User      `json:"assignee,omitempty"`
}
```

### Repository Interfaces

```go
// internal/repository/checklist_repository.go
type ChecklistRepository interface {
    Create(ctx context.Context, checklist *domain.Checklist) error
    FindByID(ctx context.Context, id string) (*domain.Checklist, error)
    FindByCardID(ctx context.Context, cardID string) ([]domain.Checklist, error)
    Update(ctx context.Context, checklist *domain.Checklist) error
    Delete(ctx context.Context, id string) error
    
    // Position management
    GetMaxPosition(ctx context.Context, cardID string) (float64, error)
    Rebalance(ctx context.Context, cardID string) error
}

type ChecklistItemRepository interface {
    Create(ctx context.Context, item *domain.ChecklistItem) error
    FindByID(ctx context.Context, id string) (*domain.ChecklistItem, error)
    FindByChecklistID(ctx context.Context, checklistID string) ([]domain.ChecklistItem, error)
    Update(ctx context.Context, item *domain.ChecklistItem) error
    Delete(ctx context.Context, id string) error
    
    // Toggle complete
    ToggleComplete(ctx context.Context, id, userID string) (*domain.ChecklistItem, error)
    
    // Position management
    GetMaxPosition(ctx context.Context, checklistID string) (float64, error)
    Rebalance(ctx context.Context, checklistID string) error
}
```

### Service Methods

```go
// internal/service/checklist_service.go
type ChecklistService interface {
    // Checklist CRUD
    Create(ctx context.Context, req CreateChecklistRequest) (*ChecklistResponse, error)
    Update(ctx context.Context, checklistID string, req UpdateChecklistRequest) (*ChecklistResponse, error)
    Delete(ctx context.Context, checklistID string) error
    
    // Checklist items
    CreateItem(ctx context.Context, req CreateChecklistItemRequest) (*ChecklistItemResponse, error)
    UpdateItem(ctx context.Context, itemID string, req UpdateChecklistItemRequest) (*ChecklistItemResponse, error)
    DeleteItem(ctx context.Context, itemID string) error
    
    // Toggle item completion (logs activity, triggers notification if assigned)
    ToggleItemComplete(ctx context.Context, itemID string) (*ChecklistItemResponse, error)
    
    // Convert item to card
    ConvertItemToCard(ctx context.Context, itemID string, listID string) (*CardResponse, error)
}
```

### DTOs

```go
// internal/dto/request/checklist_request.go
type CreateChecklistRequest struct {
    CardID string `json:"card_id" validate:"required,cuid"`
    Title  string `json:"title" validate:"required,min=1,max=255"`
}

type UpdateChecklistRequest struct {
    Title *string `json:"title,omitempty" validate:"omitempty,min=1,max=255"`
}

type CreateChecklistItemRequest struct {
    ChecklistID string  `json:"checklist_id" validate:"required,cuid"`
    Title       string  `json:"title" validate:"required,min=1,max=500"`
    AssigneeID  *string `json:"assignee_id,omitempty" validate:"omitempty,cuid"`
    DueDate     *string `json:"due_date,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
}

type UpdateChecklistItemRequest struct {
    Title      *string `json:"title,omitempty" validate:"omitempty,min=1,max=500"`
    AssigneeID *string `json:"assignee_id,omitempty" validate:"omitempty,cuid"`
    DueDate    *string `json:"due_date,omitempty"`
}

// internal/dto/response/checklist_response.go
type ChecklistResponse struct {
    ID        string                  `json:"id"`
    Title     string                  `json:"title"`
    Position  float64                 `json:"position"`
    Items     []ChecklistItemResponse `json:"items"`
    Progress  ChecklistProgress       `json:"progress"`
}

type ChecklistItemResponse struct {
    ID          string             `json:"id"`
    Title       string             `json:"title"`
    IsCompleted bool               `json:"is_completed"`
    CompletedAt *time.Time         `json:"completed_at,omitempty"`
    Assignee    *UserBriefResponse `json:"assignee,omitempty"`
    DueDate     *time.Time         `json:"due_date,omitempty"`
}

type ChecklistProgress struct {
    Completed int `json:"completed"`
    Total     int `json:"total"`
}
```

### API Endpoints

| Method | Endpoint | Description | Auth | Rate Limit |
|--------|----------|-------------|:----:|:----------:|
| POST | `/cards/:id/checklists` | Create checklist | Required (member+) | 30/min |
| PUT | `/checklists/:id` | Update checklist | Required (member+) | 30/min |
| DELETE | `/checklists/:id` | Delete checklist | Required (member+) | 30/min |
| POST | `/checklists/:id/items` | Create item | Required (member+) | 60/min |
| PUT | `/checklist-items/:id` | Update item | Required (member+) | 60/min |
| DELETE | `/checklist-items/:id` | Delete item | Required (member+) | 60/min |
| POST | `/checklist-items/:id/toggle` | Toggle complete | Required (member+) | 120/min |
| POST | `/checklist-items/:id/convert` | Convert to card | Required (member+) | 10/min |

---

## Task 4.3: Attachments Module

### Domain Model

```go
// internal/domain/attachment.go
type Attachment struct {
    ID           string     `json:"id"`
    CardID       string     `json:"card_id"`
    UploadedBy   string     `json:"uploaded_by"`
    Filename     string     `json:"filename"`
    OriginalName string     `json:"original_name"`
    MimeType     string     `json:"mime_type"`
    FileSize     int64      `json:"file_size"`
    ObjectKey    string     `json:"-"`
    URL          string     `json:"url"`
    IsCover      bool       `json:"is_cover"`
    CreatedAt    time.Time  `json:"created_at"`
    DeletedAt    *time.Time `json:"-"`
    
    Uploader     *User      `json:"uploader,omitempty"`
}

// Allowed MIME types and size limits
var (
    AllowedMimeTypes = map[string]bool{
        "image/jpeg":      true,
        "image/png":       true,
        "image/gif":       true,
        "image/webp":      true,
        "application/pdf": true,
        "text/plain":      true,
        "application/msword": true,
        "application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
        "application/vnd.ms-excel": true,
        "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
    }
    
    MaxFileSize      = int64(10 * 1024 * 1024) // 10MB
    ImageMimeTypes   = []string{"image/jpeg", "image/png", "image/gif", "image/webp"}
)
```

### Repository Interface

```go
// internal/repository/attachment_repository.go
type AttachmentRepository interface {
    Create(ctx context.Context, attachment *domain.Attachment) error
    FindByID(ctx context.Context, id string) (*domain.Attachment, error)
    FindByCardID(ctx context.Context, cardID string) ([]domain.Attachment, error)
    SoftDelete(ctx context.Context, id string) error
    
    // Cover management
    SetCover(ctx context.Context, cardID, attachmentID string) error
    RemoveCover(ctx context.Context, cardID string) error
    GetCover(ctx context.Context, cardID string) (*domain.Attachment, error)
}
```

### Service Methods

```go
// internal/service/attachment_service.go
type AttachmentService interface {
    // List attachments for a card
    ListByCard(ctx context.Context, cardID string) ([]AttachmentResponse, error)
    
    // Upload file (validates type/size, uploads to MinIO, creates record)
    Upload(ctx context.Context, req UploadAttachmentRequest) (*AttachmentResponse, error)
    
    // Delete attachment (soft delete record, optionally delete from MinIO)
    Delete(ctx context.Context, attachmentID string) error
    
    // Set/remove cover image
    SetCover(ctx context.Context, cardID, attachmentID string) error
    RemoveCover(ctx context.Context, cardID string) error
    
    // Generate presigned URL for download
    GetDownloadURL(ctx context.Context, attachmentID string) (string, error)
}
```

### DTOs

```go
// internal/dto/request/attachment_request.go
type UploadAttachmentRequest struct {
    CardID   string                `form:"card_id" validate:"required,cuid"`
    File     *multipart.FileHeader `form:"file" validate:"required"`
}

// internal/dto/response/attachment_response.go
type AttachmentResponse struct {
    ID           string            `json:"id"`
    Filename     string            `json:"filename"`
    OriginalName string            `json:"original_name"`
    MimeType     string            `json:"mime_type"`
    FileSize     int64             `json:"file_size"`
    URL          string            `json:"url"`
    IsCover      bool              `json:"is_cover"`
    CreatedAt    time.Time         `json:"created_at"`
    Uploader     UserBriefResponse `json:"uploader"`
}
```

### MinIO Integration

```go
// pkg/storage/minio.go
type StorageService interface {
    Upload(ctx context.Context, opts UploadOptions) (*UploadResult, error)
    Delete(ctx context.Context, objectKey string) error
    GetPresignedURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error)
    GetPublicURL(objectKey string) string
}

type UploadOptions struct {
    Folder      string
    Filename    string
    Reader      io.Reader
    Size        int64
    ContentType string
}

type UploadResult struct {
    ObjectKey string
    URL       string
}

// Object key format: attachments/{board_id}/{card_id}/{uuid}_{filename}
```

### API Endpoints

| Method | Endpoint | Description | Auth | Rate Limit |
|--------|----------|-------------|:----:|:----------:|
| GET | `/cards/:id/attachments` | List attachments | Required | 60/min |
| POST | `/cards/:id/attachments` | Upload file | Required (member+) | 10/min |
| DELETE | `/attachments/:id` | Delete attachment | Required (uploader/admin) | 30/min |
| POST | `/attachments/:id/cover` | Set as cover | Required (member+) | 30/min |
| DELETE | `/cards/:id/cover` | Remove cover | Required (member+) | 30/min |
| GET | `/attachments/:id/download` | Get download URL | Required | 60/min |

---

## Task 4.4: Activity Logs Module

### Domain Model

```go
// internal/domain/activity.go
type ActivityAction string

const (
    ActivityCardCreated           ActivityAction = "card_created"
    ActivityCardUpdated           ActivityAction = "card_updated"
    ActivityCardMoved             ActivityAction = "card_moved"
    ActivityCardArchived          ActivityAction = "card_archived"
    ActivityCardAssigned          ActivityAction = "card_assigned"
    ActivityCardUnassigned        ActivityAction = "card_unassigned"
    ActivityCardCompleted         ActivityAction = "card_completed"
    ActivityCardReopened          ActivityAction = "card_reopened"
    ActivityCardDueDateSet        ActivityAction = "card_due_date_set"
    ActivityCardDueDateRemoved    ActivityAction = "card_due_date_removed"
    ActivityLabelAdded            ActivityAction = "label_added"
    ActivityLabelRemoved          ActivityAction = "label_removed"
    ActivityChecklistCreated      ActivityAction = "checklist_created"
    ActivityChecklistDeleted      ActivityAction = "checklist_deleted"
    ActivityChecklistItemComplete ActivityAction = "checklist_item_completed"
    ActivityChecklistItemUncomplete ActivityAction = "checklist_item_uncompleted"
    ActivityCommentAdded          ActivityAction = "comment_added"
    ActivityCommentEdited         ActivityAction = "comment_edited"
    ActivityCommentDeleted        ActivityAction = "comment_deleted"
    ActivityAttachmentAdded       ActivityAction = "attachment_added"
    ActivityAttachmentDeleted     ActivityAction = "attachment_deleted"
    ActivityCoverSet              ActivityAction = "cover_set"
    ActivityCoverRemoved          ActivityAction = "cover_removed"
    ActivityMemberAdded           ActivityAction = "member_added"
    ActivityMemberRemoved         ActivityAction = "member_removed"
)

type ActivityLog struct {
    ID          string                 `json:"id"`
    BoardID     string                 `json:"board_id"`
    CardID      *string                `json:"card_id,omitempty"`
    ListID      *string                `json:"list_id,omitempty"`
    UserID      string                 `json:"user_id"`
    Action      ActivityAction         `json:"action"`
    Metadata    map[string]interface{} `json:"metadata"`
    Description string                 `json:"description"`
    CreatedAt   time.Time              `json:"created_at"`
    
    User        *User                  `json:"user,omitempty"`
}
```

### Repository Interface

```go
// internal/repository/activity_repository.go
type ActivityRepository interface {
    Create(ctx context.Context, activity *domain.ActivityLog) error
    
    // Query by scope
    FindByBoard(ctx context.Context, boardID string, opts PaginationOpts) ([]domain.ActivityLog, int, error)
    FindByCard(ctx context.Context, cardID string, opts PaginationOpts) ([]domain.ActivityLog, int, error)
    FindByUser(ctx context.Context, userID string, opts PaginationOpts) ([]domain.ActivityLog, int, error)
    
    // Cleanup old logs (for maintenance)
    DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)
}
```

### Service Methods

```go
// internal/service/activity_service.go
type ActivityService interface {
    // Log activity (called by other services)
    Log(ctx context.Context, req LogActivityRequest) error
    
    // Query activities
    ListByCard(ctx context.Context, cardID string, page, limit int) (*PaginatedActivityResponse, error)
    ListByBoard(ctx context.Context, boardID string, page, limit int) (*PaginatedActivityResponse, error)
    
    // Generate human-readable description
    GenerateDescription(action ActivityAction, metadata map[string]interface{}) string
}

type LogActivityRequest struct {
    BoardID     string
    CardID      *string
    ListID      *string
    UserID      string
    Action      ActivityAction
    Metadata    map[string]interface{}
}
```

### Activity Description Templates

```go
var activityTemplates = map[ActivityAction]string{
    ActivityCardCreated:    "%s created this card",
    ActivityCardMoved:      "%s moved this card from %s to %s",
    ActivityCardAssigned:   "%s assigned %s to this card",
    ActivityCardCompleted:  "%s marked this card as complete",
    ActivityCommentAdded:   "%s commented on this card",
    ActivityLabelAdded:     "%s added label %s",
    ActivityChecklistItemComplete: "%s completed %s on %s",
    // ...
}
```

### API Endpoints

| Method | Endpoint | Description | Auth | Rate Limit |
|--------|----------|-------------|:----:|:----------:|
| GET | `/cards/:id/activity` | Card activity log | Required | 60/min |
| GET | `/boards/:id/activity` | Board activity log | Required | 30/min |

---

## Task 4.5: Full Card Detail Endpoint

### Enhanced Card Response

```go
// internal/dto/response/card_detail_response.go
type CardDetailResponse struct {
    ID          string               `json:"id"`
    Title       string               `json:"title"`
    Description *string              `json:"description"`
    Position    float64              `json:"position"`
    Priority    string               `json:"priority"`
    DueDate     *time.Time           `json:"due_date"`
    IsCompleted bool                 `json:"is_completed"`
    CompletedAt *time.Time           `json:"completed_at,omitempty"`
    CoverURL    *string              `json:"cover_url"`
    CreatedAt   time.Time            `json:"created_at"`
    UpdatedAt   time.Time            `json:"updated_at"`
    
    // Related entities (fully populated)
    List        ListBriefResponse      `json:"list"`
    Assignee    *UserBriefResponse     `json:"assignee"`
    Labels      []LabelResponse        `json:"labels"`
    Members     []UserBriefResponse    `json:"members"`
    Checklists  []ChecklistResponse    `json:"checklists"`
    Comments    []CommentResponse      `json:"comments"`
    Attachments []AttachmentResponse   `json:"attachments"`
    Activity    []ActivityLogResponse  `json:"activity"`
    
    // Computed
    ChecklistsProgress ChecklistProgress `json:"checklists_progress"`
}

type ListBriefResponse struct {
    ID    string `json:"id"`
    Title string `json:"title"`
}
```

### Service Enhancement

```go
// internal/service/card_service.go
// Enhanced GetByID to include all related data
func (s *cardService) GetDetail(ctx context.Context, cardID string) (*CardDetailResponse, error) {
    // Parallel fetch for performance
    var (
        card        *domain.Card
        comments    []domain.Comment
        checklists  []domain.Checklist
        attachments []domain.Attachment
        activity    []domain.ActivityLog
    )
    
    g, ctx := errgroup.WithContext(ctx)
    
    g.Go(func() error {
        var err error
        card, err = s.cardRepo.FindByID(ctx, cardID)
        return err
    })
    
    g.Go(func() error {
        var err error
        comments, err = s.commentRepo.FindByCardID(ctx, cardID)
        return err
    })
    
    // ... other parallel fetches
    
    if err := g.Wait(); err != nil {
        return nil, err
    }
    
    // Assemble response
    return s.assembleCardDetail(card, comments, checklists, attachments, activity), nil
}
```

---

## Dependency Graph

```
Migration 00014 (comments)
    │
    ├──▶ Task 4.1 (Comments Module)
    │        │
Migration 00015 (checklists)          │
    │                                  │
    ├──▶ Task 4.2 (Checklists Module) ─┤
    │                                  │
Migration 00016 (checklist_items)     │
    │                                  │
    └────────────────────────────────▶│
                                       │
Migration 00017 (attachments)          │
    │                                  │
    ├──▶ Task 4.3 (Attachments Module)─┤
    │                                  │
Migration 00018 (activity_logs)        │
    │                                  │
    └──▶ Task 4.4 (Activity Module) ───┤
                                       │
                                       ▼
                              Task 4.5 (Card Detail)
```

---

## Recommended Execution Order

| Day | Tasks | Focus |
|:---:|-------|-------|
| 1 | Migrations 14-15 | Comments + Checklists tables |
| 2 | Task 4.1 | Comments domain, repo, service |
| 3 | Task 4.1 | Comments handlers, tests |
| 4 | Task 4.2 | Checklists domain, repo |
| 5 | Task 4.2 | Checklist items, toggle, convert |
| 6 | Migration 17, Task 4.3 | Attachments + MinIO integration |
| 7 | Task 4.3 | Upload, cover, download URLs |
| 8 | Migration 18, Task 4.4 | Activity logs |
| 9 | Task 4.5 | Full card detail endpoint |
| 10 | Integration | End-to-end testing |

---

## Phase 4 Deliverables Checklist

### Comments
- [ ] `POST /cards/:id/comments` → 201 + comment
- [ ] `GET /cards/:id/comments` → 200 + nested replies
- [ ] `PUT /comments/:id` → 200 + marks is_edited
- [ ] `DELETE /comments/:id` → 200 (soft delete)
- [ ] Activity logged on comment actions

### Checklists
- [ ] `POST /cards/:id/checklists` → 201
- [ ] `PUT /checklists/:id` → 200
- [ ] `DELETE /checklists/:id` → 200
- [ ] `POST /checklists/:id/items` → 201
- [ ] `PUT /checklist-items/:id` → 200
- [ ] `DELETE /checklist-items/:id` → 200
- [ ] `POST /checklist-items/:id/toggle` → 200 (logs activity)
- [ ] `POST /checklist-items/:id/convert` → 201 (creates card)
- [ ] Position ordering works
- [ ] Progress calculation correct

### Attachments
- [ ] `POST /cards/:id/attachments` → 201 (multipart)
- [ ] `GET /cards/:id/attachments` → 200
- [ ] `DELETE /attachments/:id` → 200
- [ ] `POST /attachments/:id/cover` → 200
- [ ] `DELETE /cards/:id/cover` → 200
- [ ] `GET /attachments/:id/download` → presigned URL
- [ ] File size limit enforced (10MB)
- [ ] MIME type validation works
- [ ] MinIO storage working

### Activity
- [ ] `GET /cards/:id/activity` → 200 + paginated
- [ ] `GET /boards/:id/activity` → 200 + paginated
- [ ] All actions logged correctly
- [ ] Human-readable descriptions

### Card Detail
- [ ] `GET /cards/:id` → 200 with all nested data
- [ ] Parallel fetching for performance
- [ ] Checklists progress computed

### Tests
- [ ] Unit tests for all services
- [ ] Integration tests for endpoints
- [ ] File upload tests
