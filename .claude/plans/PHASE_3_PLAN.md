# Phase 3: Lists, Cards & Labels (1.5 Weeks)

> Detailed execution plan for the Kanban board core: lists, cards, drag & drop, and labels.

---

## Overview

```
Week 5: Lists & Cards         Week 6: Labels & Polish
┌──────────────────┐          ┌──────────────────┐
│ 3.1 List Module  │          │ 3.3 Label Module │
│   - CRUD         │          │   - Board labels │
│   - Positioning  │─────────▶│   - Card labels  │
│ 3.2 Card Module  │          │ 3.4 Integration  │
│   - CRUD + Move  │          │   - Board Detail │
│   - Assign/Done  │          │   - Position Pkg │
└──────────────────┘          └──────────────────┘
```

---

## Prerequisites

Phase 3 depends on Phase 2 completion:
- [x] Board CRUD working
- [x] Board permission middleware functional
- [x] Organization membership verified

---

## Database Migrations (5 files)

| Migration | Table | Description |
|-----------|-------|-------------|
| `00009_create_lists.sql` | `lists` | Kanban columns |
| `00010_create_cards.sql` | `cards` | Tasks/tickets |
| `00011_create_card_members.sql` | `card_members` | Card watchers |
| `00012_create_labels.sql` | `labels` | Board labels |
| `00013_create_card_labels.sql` | `card_labels` | Card-label assignments |

---

## Position Strategy (Critical for Drag & Drop)

### Float-based Positioning

```go
// pkg/position/position.go

const (
    InitialGap    = 65536.0  // Gap between positions
    MinGap        = 1.0      // Trigger rebalance when gap < 1
    RebalanceGap  = 65536.0  // Gap after rebalancing
)

// When creating new item at end
func NextPosition(currentMax float64) float64 {
    if currentMax == 0 {
        return InitialGap
    }
    return currentMax + InitialGap
}

// When inserting between two items
func Between(before, after float64) float64 {
    return (before + after) / 2
}

// When inserting at start
func BeforeFirst(firstPosition float64) float64 {
    return firstPosition / 2
}

// Check if rebalance needed
func NeedsRebalance(before, after float64) bool {
    return math.Abs(after - before) < MinGap
}

// Rebalance all positions in sequence
func Rebalance(count int) []float64 {
    positions := make([]float64, count)
    for i := 0; i < count; i++ {
        positions[i] = float64(i+1) * RebalanceGap
    }
    return positions
}
```

### Position Examples

```
Initial state:
  List A: position = 65536
  List B: position = 131072
  List C: position = 196608

Insert between A and B:
  New List: position = (65536 + 131072) / 2 = 98304
  
Result:
  List A: 65536
  New:    98304
  List B: 131072
  List C: 196608

After many insertions (positions too close):
  List X: 65536.5
  List Y: 65536.75  ← gap < 1, needs rebalance
  
Rebalance triggers:
  List X: 65536
  List Y: 131072
  List Z: 196608
```

---

## Task 3.1: List Module

**Duration:** 2-3 days

### Files to Create

```
internal/
├── domain/
│   └── list.go
├── repository/
│   └── list_repository.go
├── service/
│   └── list_service.go
├── handler/
│   └── list_handler.go
└── dto/
    ├── request/
    │   └── list_request.go
    └── response/
        └── list_response.go

pkg/
└── position/
    └── position.go

migrations/
└── 00009_create_lists.sql
```

### Domain Model

```go
// internal/domain/list.go
type List struct {
    ID         string
    BoardID    string
    Title      string
    Position   float64
    IsArchived bool
    CreatedAt  time.Time
    UpdatedAt  time.Time
    ArchivedAt *time.Time
    
    // Aggregates (for response)
    Cards      []*Card
    CardsCount int
}
```

### Repository Interface

```go
type ListRepository interface {
    // CRUD
    Create(ctx, list *List) error
    FindByID(ctx, id string) (*List, error)
    FindByBoardID(ctx, boardID string, includeArchived bool) ([]*List, error)
    Update(ctx, list *List) error
    Archive(ctx, id string) error
    Restore(ctx, id string) error
    
    // Position
    GetMaxPosition(ctx, boardID string) (float64, error)
    UpdatePosition(ctx, id string, position float64) error
    RebalancePositions(ctx, boardID string) error
    
    // With cards
    FindByBoardIDWithCards(ctx, boardID string) ([]*List, error)
}
```

### Service Methods

```go
type ListService interface {
    Create(ctx, userID, boardID string, req CreateListRequest) (*List, error)
    Update(ctx, userID, listID string, req UpdateListRequest) (*List, error)
    Archive(ctx, userID, listID string) error
    Restore(ctx, userID, listID string) error
    Move(ctx, userID, listID string, req MoveListRequest) (*List, error)
    Copy(ctx, userID, listID string, req CopyListRequest) (*List, error)
}
```

### API Endpoints (5 endpoints)

| Method | Endpoint | Auth | Permission | Description |
|--------|----------|:----:|------------|-------------|
| POST | `/boards/:boardId/lists` | Yes | board member+ | Create list |
| PUT | `/lists/:id` | Yes | board member+ | Update list title |
| DELETE | `/lists/:id` | Yes | board member+ | Archive list |
| PUT | `/lists/:id/move` | Yes | board member+ | Reorder list |
| POST | `/lists/:id/copy` | Yes | board member+ | Copy list with cards |

### Request/Response DTOs

```go
// Request
type CreateListRequest struct {
    Title string `json:"title" validate:"required,min=1,max=255"`
}

type UpdateListRequest struct {
    Title string `json:"title" validate:"required,min=1,max=255"`
}

type MoveListRequest struct {
    Position float64 `json:"position" validate:"required,gt=0"`
}

type CopyListRequest struct {
    Title string `json:"title" validate:"required,min=1,max=255"`
}

// Response
type ListResponse struct {
    ID         string    `json:"id"`
    Title      string    `json:"title"`
    Position   float64   `json:"position"`
    IsArchived bool      `json:"is_archived"`
    CardsCount int       `json:"cards_count"`
    CreatedAt  time.Time `json:"created_at"`
}

type ListWithCardsResponse struct {
    ID         string         `json:"id"`
    Title      string         `json:"title"`
    Position   float64        `json:"position"`
    Cards      []CardSummary  `json:"cards"`
}
```

### Acceptance Criteria

- [ ] Create list at end of board (max position + gap)
- [ ] Update list title
- [ ] Archive list (soft delete)
- [ ] Restore archived list
- [ ] Move list to new position (drag & drop)
- [ ] Position rebalancing when gap < 1
- [ ] Copy list with all cards

---

## Task 3.2: Card Module

**Duration:** 3-4 days

### Files to Create

```
internal/
├── domain/
│   ├── card.go
│   └── card_member.go
├── repository/
│   └── card_repository.go
├── service/
│   └── card_service.go
├── handler/
│   └── card_handler.go
└── dto/
    ├── request/
    │   └── card_request.go
    └── response/
        └── card_response.go

migrations/
├── 00010_create_cards.sql
└── 00011_create_card_members.sql
```

### Domain Model

```go
// internal/domain/card.go
type CardPriority string

const (
    PriorityNone   CardPriority = "none"
    PriorityLow    CardPriority = "low"
    PriorityMedium CardPriority = "medium"
    PriorityHigh   CardPriority = "high"
)

type Card struct {
    ID                string
    ListID            string
    Title             string
    Description       *string
    Position          float64
    AssigneeID        *string
    Priority          CardPriority
    DueDate           *time.Time
    IsCompleted       bool
    CompletedAt       *time.Time
    CoverAttachmentID *string
    IsArchived        bool
    CreatedAt         time.Time
    UpdatedAt         time.Time
    ArchivedAt        *time.Time
    CreatedBy         string
    
    // Aggregates (for response)
    Assignee          *User
    Labels            []*Label
    CommentsCount     int
    AttachmentsCount  int
    ChecklistProgress *ChecklistProgress
}

type ChecklistProgress struct {
    Completed int `json:"completed"`
    Total     int `json:"total"`
}

// internal/domain/card_member.go
type CardMember struct {
    ID      string
    CardID  string
    UserID  string
    AddedAt time.Time
    
    User    *User
}
```

### Repository Interface

```go
type CardRepository interface {
    // CRUD
    Create(ctx, card *Card) error
    FindByID(ctx, id string) (*Card, error)
    FindByIDWithDetails(ctx, id string) (*Card, error)  // With labels, assignee, counts
    FindByListID(ctx, listID string, includeArchived bool) ([]*Card, error)
    Update(ctx, card *Card) error
    Archive(ctx, id string) error
    Restore(ctx, id string) error
    
    // Position
    GetMaxPosition(ctx, listID string) (float64, error)
    UpdatePosition(ctx, id string, listID string, position float64) error
    RebalancePositions(ctx, listID string) error
    
    // Assignment
    Assign(ctx, cardID, userID string) error
    Unassign(ctx, cardID string) error
    
    // Completion
    MarkComplete(ctx, cardID string) error
    MarkIncomplete(ctx, cardID string) error
    
    // Card members (watchers)
    AddMember(ctx, cardID, userID string) error
    RemoveMember(ctx, cardID, userID string) error
    FindMembers(ctx, cardID string) ([]*CardMember, error)
    
    // Board context
    FindBoardIDByCardID(ctx, cardID string) (string, error)
}
```

### Service Methods

```go
type CardService interface {
    // CRUD
    Create(ctx, userID, listID string, req CreateCardRequest) (*Card, error)
    GetByID(ctx, userID, cardID string) (*CardDetail, error)
    Update(ctx, userID, cardID string, req UpdateCardRequest) (*Card, error)
    Archive(ctx, userID, cardID string) error
    Restore(ctx, userID, cardID string) error
    
    // Move (drag & drop)
    Move(ctx, userID, cardID string, req MoveCardRequest) (*Card, error)
    
    // Assignment
    Assign(ctx, userID, cardID string, assigneeID string) error
    Unassign(ctx, userID, cardID string) error
    
    // Completion
    MarkComplete(ctx, userID, cardID string) error
    MarkIncomplete(ctx, userID, cardID string) error
    
    // Watchers
    AddWatcher(ctx, userID, cardID string, watcherID string) error
    RemoveWatcher(ctx, userID, cardID string, watcherID string) error
}
```

### API Endpoints (9 endpoints)

| Method | Endpoint | Auth | Permission | Description |
|--------|----------|:----:|------------|-------------|
| POST | `/lists/:listId/cards` | Yes | board member+ | Create card |
| GET | `/cards/:id` | Yes | board viewer+ | Get card details |
| PUT | `/cards/:id` | Yes | board member+ | Update card |
| DELETE | `/cards/:id` | Yes | board member+ | Archive card |
| PUT | `/cards/:id/move` | Yes | board member+ | Move card |
| POST | `/cards/:id/assign` | Yes | board member+ | Assign card |
| DELETE | `/cards/:id/assign` | Yes | board member+ | Unassign card |
| POST | `/cards/:id/complete` | Yes | board member+ | Mark complete |
| DELETE | `/cards/:id/complete` | Yes | board member+ | Mark incomplete |

### Request/Response DTOs

```go
// Request
type CreateCardRequest struct {
    Title string `json:"title" validate:"required,min=1,max=255"`
}

type UpdateCardRequest struct {
    Title       *string       `json:"title" validate:"omitempty,min=1,max=255"`
    Description *string       `json:"description" validate:"omitempty,max=10000"`
    Priority    *CardPriority `json:"priority" validate:"omitempty,oneof=none low medium high"`
    DueDate     *time.Time    `json:"due_date"`
}

type MoveCardRequest struct {
    ListID   string  `json:"list_id" validate:"required"`
    Position float64 `json:"position" validate:"required,gt=0"`
}

type AssignCardRequest struct {
    UserID string `json:"user_id" validate:"required"`
}

// Response
type CardSummary struct {
    ID                string             `json:"id"`
    Title             string             `json:"title"`
    Position          float64            `json:"position"`
    Description       *string            `json:"description"`
    Priority          CardPriority       `json:"priority"`
    DueDate           *time.Time         `json:"due_date"`
    IsCompleted       bool               `json:"is_completed"`
    Assignee          *UserSummary       `json:"assignee"`
    Labels            []LabelSummary     `json:"labels"`
    CommentsCount     int                `json:"comments_count"`
    AttachmentsCount  int                `json:"attachments_count"`
    ChecklistProgress *ChecklistProgress `json:"checklists_progress"`
}

type CardDetail struct {
    ID                string             `json:"id"`
    Title             string             `json:"title"`
    Description       *string            `json:"description"`
    Position          float64            `json:"position"`
    Priority          CardPriority       `json:"priority"`
    DueDate           *time.Time         `json:"due_date"`
    IsCompleted       bool               `json:"is_completed"`
    CoverURL          *string            `json:"cover_url"`
    List              ListSummary        `json:"list"`
    Assignee          *UserSummary       `json:"assignee"`
    Labels            []LabelSummary     `json:"labels"`
    // Phase 4 fields (empty for now)
    Checklists        []interface{}      `json:"checklists"`
    Comments          []interface{}      `json:"comments"`
    Attachments       []interface{}      `json:"attachments"`
    Activity          []interface{}      `json:"activity"`
    CreatedAt         time.Time          `json:"created_at"`
    UpdatedAt         time.Time          `json:"updated_at"`
}
```

### Acceptance Criteria

- [ ] Create card at end of list
- [ ] Get card with full details
- [ ] Update card fields (title, description, priority, due date)
- [ ] Archive/restore card
- [ ] Move card within same list (reorder)
- [ ] Move card to different list
- [ ] Position rebalancing when needed
- [ ] Assign single user to card
- [ ] Unassign card
- [ ] Mark card complete/incomplete

---

## Task 3.3: Label Module

**Duration:** 2 days

### Files to Create

```
internal/
├── domain/
│   ├── label.go
│   └── card_label.go
├── repository/
│   └── label_repository.go
├── service/
│   └── label_service.go
├── handler/
│   └── label_handler.go
└── dto/
    ├── request/
    │   └── label_request.go
    └── response/
        └── label_response.go

migrations/
├── 00012_create_labels.sql
└── 00013_create_card_labels.sql
```

### Domain Model

```go
// internal/domain/label.go
type Label struct {
    ID        string
    BoardID   string
    Name      *string   // Optional, can be color-only
    Color     string    // Hex color #XXXXXX
    CreatedAt time.Time
    UpdatedAt time.Time
}

// Default label colors (Trello-style)
var DefaultLabelColors = []string{
    "#61bd4f", // Green
    "#f2d600", // Yellow
    "#ff9f1a", // Orange
    "#eb5a46", // Red
    "#c377e0", // Purple
    "#0079bf", // Blue
    "#00c2e0", // Sky
    "#51e898", // Lime
    "#ff78cb", // Pink
    "#344563", // Dark
}

// internal/domain/card_label.go
type CardLabel struct {
    ID         string
    CardID     string
    LabelID    string
    AssignedAt time.Time
}
```

### Repository Interface

```go
type LabelRepository interface {
    // Board labels
    Create(ctx, label *Label) error
    FindByID(ctx, id string) (*Label, error)
    FindByBoardID(ctx, boardID string) ([]*Label, error)
    Update(ctx, label *Label) error
    Delete(ctx, id string) error
    
    // Card labels
    AssignToCard(ctx, cardID, labelID string) error
    RemoveFromCard(ctx, cardID, labelID string) error
    FindByCardID(ctx, cardID string) ([]*Label, error)
    IsAssignedToCard(ctx, cardID, labelID string) (bool, error)
    
    // Board context
    FindBoardIDByLabelID(ctx, labelID string) (string, error)
}
```

### Service Methods

```go
type LabelService interface {
    // Board labels
    Create(ctx, userID, boardID string, req CreateLabelRequest) (*Label, error)
    List(ctx, userID, boardID string) ([]*Label, error)
    Update(ctx, userID, labelID string, req UpdateLabelRequest) (*Label, error)
    Delete(ctx, userID, labelID string) error
    
    // Card labels
    AssignToCard(ctx, userID, cardID, labelID string) error
    RemoveFromCard(ctx, userID, cardID, labelID string) error
}
```

### API Endpoints (6 endpoints)

| Method | Endpoint | Auth | Permission | Description |
|--------|----------|:----:|------------|-------------|
| GET | `/boards/:boardId/labels` | Yes | board viewer+ | List board labels |
| POST | `/boards/:boardId/labels` | Yes | board admin+ | Create label |
| PUT | `/labels/:id` | Yes | board admin+ | Update label |
| DELETE | `/labels/:id` | Yes | board admin+ | Delete label |
| POST | `/cards/:cardId/labels/:labelId` | Yes | board member+ | Assign label |
| DELETE | `/cards/:cardId/labels/:labelId` | Yes | board member+ | Remove label |

### Request/Response DTOs

```go
// Request
type CreateLabelRequest struct {
    Name  *string `json:"name" validate:"omitempty,max=100"`
    Color string  `json:"color" validate:"required,hexcolor"`
}

type UpdateLabelRequest struct {
    Name  *string `json:"name" validate:"omitempty,max=100"`
    Color *string `json:"color" validate:"omitempty,hexcolor"`
}

// Response
type LabelResponse struct {
    ID    string  `json:"id"`
    Name  *string `json:"name"`
    Color string  `json:"color"`
}

type LabelSummary struct {
    ID    string  `json:"id"`
    Name  *string `json:"name"`
    Color string  `json:"color"`
}
```

### Acceptance Criteria

- [ ] Create label with color (name optional)
- [ ] List all labels for a board
- [ ] Update label name/color
- [ ] Delete label (removes from all cards)
- [ ] Assign label to card
- [ ] Remove label from card
- [ ] Card can have multiple labels
- [ ] Label colors are hex format

---

## Task 3.4: Board Detail Integration

**Duration:** 1 day

### Update GET /boards/:id Response

The board detail endpoint needs to return lists with cards:

```go
// Updated BoardDetail response
type BoardDetail struct {
    ID              string              `json:"id"`
    Title           string              `json:"title"`
    Description     *string             `json:"description"`
    BackgroundColor string              `json:"background_color"`
    Visibility      BoardVisibility     `json:"visibility"`
    IsClosed        bool                `json:"is_closed"`
    MyRole          BoardRole           `json:"my_role"`
    Organization    OrgSummary          `json:"organization"`
    Lists           []ListWithCards     `json:"lists"`      // NOW POPULATED
    Labels          []LabelSummary      `json:"labels"`     // NOW POPULATED
    Members         []MemberSummary     `json:"members"`
    CreatedAt       time.Time           `json:"created_at"`
}

type ListWithCards struct {
    ID       string        `json:"id"`
    Title    string        `json:"title"`
    Position float64       `json:"position"`
    Cards    []CardSummary `json:"cards"`
}
```

### Query Optimization

```go
// Efficient board loading with all data
func (r *BoardRepository) FindByIDWithFullDetails(ctx context.Context, id string) (*BoardDetail, error) {
    // Single query with JOINs or batch queries
    // 1. Board + Organization
    // 2. Lists (ordered by position)
    // 3. Cards (ordered by list_id, position) with:
    //    - Assignee
    //    - Labels
    //    - Counts (comments, attachments, checklist progress)
    // 4. Board labels
    // 5. Board members
    
    // Use DataLoader pattern to avoid N+1
}
```

---

## Dependency Graph

```
Phase 2 (Boards) ───────────────────────────────────┐
                                                    │
Task 3.1 (Lists)                                    │
    │                                               │
    ├── migrations (00009)                          │
    ├── pkg/position/position.go  ◀─────────────────┤
    ├── domain/list.go                              │
    ├── repository/list_repository.go               │
    ├── service/list_service.go                     │
    └── handler/list_handler.go                     │
            │                                       │
            ▼                                       │
Task 3.2 (Cards) ◀──────────────────────────────────┘
    │
    ├── migrations (00010, 00011)
    ├── domain/card.go
    ├── repository/card_repository.go
    ├── service/card_service.go
    └── handler/card_handler.go
            │
            ▼
Task 3.3 (Labels)
    │
    ├── migrations (00012, 00013)
    ├── domain/label.go
    ├── repository/label_repository.go
    ├── service/label_service.go
    └── handler/label_handler.go
            │
            ▼
Task 3.4 (Integration)
    │
    └── Update BoardService.GetByID to include lists, cards, labels
```

---

## Recommended Execution Order

| Day | Tasks | Focus |
|:---:|-------|-------|
| 1 | pkg/position | Position utilities + tests |
| 1 | Migrations | 00009, 00010, 00011, 00012, 00013 |
| 2 | List module | Domain, repository, service |
| 3 | List handler | Endpoints + move/rebalance |
| 4 | Card module | Domain, repository |
| 5 | Card service | CRUD, move, assign, complete |
| 6 | Card handler | 9 endpoints |
| 7 | Label module | Full CRUD + card assignment |
| 8 | Integration | Board detail with lists/cards |
| 9 | Testing | Unit + integration + drag & drop |

---

## API Response Examples

### GET /boards/:id (Full Board)

```json
{
  "success": true,
  "data": {
    "id": "clx_board_123",
    "title": "Project Alpha",
    "background_color": "#0079bf",
    "visibility": "workspace",
    "is_closed": false,
    "my_role": "admin",
    "organization": {
      "id": "clx_org_123",
      "name": "My Workspace",
      "slug": "my-workspace"
    },
    "lists": [
      {
        "id": "clx_list_1",
        "title": "To Do",
        "position": 65536,
        "cards": [
          {
            "id": "clx_card_1",
            "title": "Implement login",
            "position": 65536,
            "priority": "high",
            "due_date": "2026-04-30T00:00:00Z",
            "is_completed": false,
            "assignee": {
              "id": "clx_user_1",
              "full_name": "John Doe",
              "avatar_url": null
            },
            "labels": [
              {"id": "clx_label_1", "name": "Bug", "color": "#eb5a46"}
            ],
            "comments_count": 3,
            "attachments_count": 1,
            "checklists_progress": {"completed": 2, "total": 5}
          }
        ]
      },
      {
        "id": "clx_list_2",
        "title": "In Progress",
        "position": 131072,
        "cards": []
      }
    ],
    "labels": [
      {"id": "clx_label_1", "name": "Bug", "color": "#eb5a46"},
      {"id": "clx_label_2", "name": "Feature", "color": "#61bd4f"}
    ],
    "members": [
      {"id": "clx_user_1", "full_name": "John Doe", "avatar_url": null, "role": "owner"}
    ]
  }
}
```

### PUT /cards/:id/move (Drag & Drop)

```json
// Request
{
  "list_id": "clx_list_2",
  "position": 98304
}

// Response
{
  "success": true,
  "data": {
    "id": "clx_card_1",
    "list_id": "clx_list_2",
    "position": 98304
  }
}
```

---

## Phase 3 Deliverables Checklist

### Lists
- [ ] Create list at end of board
- [ ] Update list title
- [ ] Archive/restore list
- [ ] Move list (drag & drop)
- [ ] Copy list with cards
- [ ] Position rebalancing works

### Cards
- [ ] Create card in list
- [ ] Get card with full details
- [ ] Update card (title, description, priority, due date)
- [ ] Archive/restore card
- [ ] Move card within list
- [ ] Move card to different list
- [ ] Assign/unassign user
- [ ] Mark complete/incomplete
- [ ] Position rebalancing works

### Labels
- [ ] Create board label
- [ ] List board labels
- [ ] Update label
- [ ] Delete label
- [ ] Assign label to card
- [ ] Remove label from card
- [ ] Multiple labels per card

### Integration
- [ ] GET /boards/:id returns lists with cards
- [ ] Cards include labels, assignee, counts
- [ ] Positions persist after refresh
- [ ] N+1 queries prevented

### Tests
- [ ] Position utilities unit tests
- [ ] List service unit tests
- [ ] Card service unit tests
- [ ] Label service unit tests
- [ ] Integration tests for drag & drop
