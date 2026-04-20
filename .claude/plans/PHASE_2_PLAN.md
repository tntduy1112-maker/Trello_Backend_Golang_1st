# Phase 2: Organization & Board (1.5 Weeks)

> Detailed execution plan for workspaces, boards, and invitation system.

---

## Overview

```
Week 3: Organizations       Week 4: Boards + Invitations
┌──────────────────┐        ┌──────────────────┐
│ 2.1 Org Domain   │        │ 2.2 Board Module │
│ 2.1 Org Members  │───────▶│ 2.2 Board Members│
│ 2.1 Permissions  │        │ 2.3 Invitations  │
└──────────────────┘        └──────────────────┘
```

---

## Prerequisites

Phase 2 depends on Phase 1 completion:
- [x] User authentication working
- [x] JWT middleware functional
- [x] Database connection established
- [x] CUID, validator packages ready

---

## Database Migrations (5 files)

| Migration | Table | Description |
|-----------|-------|-------------|
| `00004_create_organizations.sql` | `organizations` | Workspaces |
| `00005_create_organization_members.sql` | `organization_members` | Org membership + roles |
| `00006_create_boards.sql` | `boards` | Project boards |
| `00007_create_board_members.sql` | `board_members` | Board membership + roles |
| `00008_create_board_invitations.sql` | `board_invitations` | Invite tokens |

---

## Task 2.1: Organization Module

**Duration:** 3-4 days

### Files to Create

```
internal/
├── domain/
│   ├── organization.go
│   └── organization_member.go
├── repository/
│   └── organization_repository.go
├── service/
│   └── organization_service.go
├── handler/
│   └── organization_handler.go
└── dto/
    ├── request/
    │   └── organization_request.go
    └── response/
        └── organization_response.go

migrations/
├── 00004_create_organizations.sql
└── 00005_create_organization_members.sql
```

### Domain Models

```go
// internal/domain/organization.go
type Organization struct {
    ID          string
    Name        string
    Slug        string     // URL-friendly unique identifier
    Description *string
    LogoURL     *string
    OwnerID     string
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   *time.Time
}

// internal/domain/organization_member.go
type OrgRole string

const (
    OrgRoleOwner  OrgRole = "owner"
    OrgRoleAdmin  OrgRole = "admin"
    OrgRoleMember OrgRole = "member"
)

type OrganizationMember struct {
    ID             string
    OrganizationID string
    UserID         string
    Role           OrgRole
    JoinedAt       time.Time
    
    // Joined data
    User           *User
}
```

### Repository Interface

```go
type OrganizationRepository interface {
    // Organizations
    Create(ctx, org *Organization) error
    FindByID(ctx, id string) (*Organization, error)
    FindBySlug(ctx, slug string) (*Organization, error)
    FindByUserID(ctx, userID string) ([]*Organization, error)
    Update(ctx, org *Organization) error
    SoftDelete(ctx, id string) error
    
    // Slug
    SlugExists(ctx, slug string) (bool, error)
    GenerateUniqueSlug(ctx, name string) (string, error)
    
    // Members
    AddMember(ctx, member *OrganizationMember) error
    FindMember(ctx, orgID, userID string) (*OrganizationMember, error)
    FindMembers(ctx, orgID string) ([]*OrganizationMember, error)
    UpdateMemberRole(ctx, orgID, userID string, role OrgRole) error
    RemoveMember(ctx, orgID, userID string) error
    CountMembers(ctx, orgID string) (int, error)
}
```

### Service Methods

```go
type OrganizationService interface {
    // Organizations
    Create(ctx, userID string, req CreateOrgRequest) (*Organization, error)
    GetBySlug(ctx, userID, slug string) (*OrganizationDetail, error)
    List(ctx, userID string) ([]*OrganizationSummary, error)
    Update(ctx, userID, slug string, req UpdateOrgRequest) (*Organization, error)
    Delete(ctx, userID, slug string) error
    
    // Members
    ListMembers(ctx, userID, slug string) ([]*MemberResponse, error)
    InviteMember(ctx, userID, slug string, req InviteMemberRequest) error
    UpdateMemberRole(ctx, userID, slug, targetUserID string, role OrgRole) error
    RemoveMember(ctx, userID, slug, targetUserID string) error
    LeaveOrganization(ctx, userID, slug string) error
}
```

### API Endpoints (9 endpoints)

| Method | Endpoint | Auth | Permission | Description |
|--------|----------|:----:|------------|-------------|
| GET | `/organizations` | Yes | — | List user's organizations |
| POST | `/organizations` | Yes | — | Create organization |
| GET | `/organizations/:slug` | Yes | member+ | Get organization details |
| PUT | `/organizations/:slug` | Yes | admin+ | Update organization |
| DELETE | `/organizations/:slug` | Yes | owner | Delete organization |
| GET | `/organizations/:slug/members` | Yes | member+ | List members |
| POST | `/organizations/:slug/members` | Yes | admin+ | Invite member |
| PUT | `/organizations/:slug/members/:userId` | Yes | admin+ | Update member role |
| DELETE | `/organizations/:slug/members/:userId` | Yes | admin+ | Remove member |

### Permission Matrix (Organization)

| Action | Owner | Admin | Member |
|--------|:-----:|:-----:|:------:|
| View workspace | ✅ | ✅ | ✅ |
| Create board | ✅ | ✅ | ✅ |
| Edit workspace settings | ✅ | ✅ | ❌ |
| Invite members | ✅ | ✅ | ❌ |
| Remove members | ✅ | ✅ | ❌ |
| Change member roles | ✅ | ✅* | ❌ |
| Delete workspace | ✅ | ❌ | ❌ |
| Transfer ownership | ✅ | ❌ | ❌ |

> *Admin can only manage members, not other admins

### Request/Response DTOs

```go
// Request
type CreateOrgRequest struct {
    Name        string  `json:"name" validate:"required,min=1,max=255"`
    Description *string `json:"description" validate:"omitempty,max=1000"`
}

type UpdateOrgRequest struct {
    Name        *string `json:"name" validate:"omitempty,min=1,max=255"`
    Description *string `json:"description" validate:"omitempty,max=1000"`
}

type InviteMemberRequest struct {
    Email string  `json:"email" validate:"required,email"`
    Role  OrgRole `json:"role" validate:"required,oneof=admin member"`
}

// Response
type OrganizationSummary struct {
    ID          string  `json:"id"`
    Name        string  `json:"name"`
    Slug        string  `json:"slug"`
    LogoURL     *string `json:"logo_url"`
    Role        OrgRole `json:"role"`        // User's role in this org
    BoardsCount int     `json:"boards_count"`
}

type OrganizationDetail struct {
    ID           string         `json:"id"`
    Name         string         `json:"name"`
    Slug         string         `json:"slug"`
    Description  *string        `json:"description"`
    LogoURL      *string        `json:"logo_url"`
    Owner        UserSummary    `json:"owner"`
    MembersCount int            `json:"members_count"`
    BoardsCount  int            `json:"boards_count"`
    MyRole       OrgRole        `json:"my_role"`
    CreatedAt    time.Time      `json:"created_at"`
}
```

### Acceptance Criteria

- [ ] Create organization with auto-generated slug
- [ ] Slug is unique and URL-friendly (kebab-case)
- [ ] Creator becomes owner automatically
- [ ] List returns only user's organizations
- [ ] Permission checks enforced on all endpoints
- [ ] Owner cannot be removed
- [ ] Owner cannot leave (must transfer first)
- [ ] Soft delete works

---

## Task 2.2: Board Module

**Duration:** 3-4 days

### Files to Create

```
internal/
├── domain/
│   ├── board.go
│   └── board_member.go
├── repository/
│   └── board_repository.go
├── service/
│   └── board_service.go
├── handler/
│   └── board_handler.go
└── dto/
    ├── request/
    │   └── board_request.go
    └── response/
        └── board_response.go

migrations/
├── 00006_create_boards.sql
└── 00007_create_board_members.sql
```

### Domain Models

```go
// internal/domain/board.go
type BoardVisibility string

const (
    VisibilityPrivate   BoardVisibility = "private"
    VisibilityWorkspace BoardVisibility = "workspace"
    VisibilityPublic    BoardVisibility = "public"
)

type Board struct {
    ID              string
    OrganizationID  string
    Title           string
    Description     *string
    BackgroundColor string          // Hex color #XXXXXX
    BackgroundImage *string
    Visibility      BoardVisibility
    IsClosed        bool
    OwnerID         string
    CreatedAt       time.Time
    UpdatedAt       time.Time
    ClosedAt        *time.Time
    DeletedAt       *time.Time
}

// internal/domain/board_member.go
type BoardRole string

const (
    BoardRoleOwner  BoardRole = "owner"
    BoardRoleAdmin  BoardRole = "admin"
    BoardRoleMember BoardRole = "member"
    BoardRoleViewer BoardRole = "viewer"
)

type BoardMember struct {
    ID       string
    BoardID  string
    UserID   string
    Role     BoardRole
    JoinedAt time.Time
    
    // Joined data
    User     *User
}
```

### Repository Interface

```go
type BoardRepository interface {
    // Boards
    Create(ctx, board *Board) error
    FindByID(ctx, id string) (*Board, error)
    FindByOrgID(ctx, orgID string, includesClosed bool) ([]*Board, error)
    Update(ctx, board *Board) error
    Close(ctx, id string) error
    Reopen(ctx, id string) error
    SoftDelete(ctx, id string) error
    
    // Members
    AddMember(ctx, member *BoardMember) error
    FindMember(ctx, boardID, userID string) (*BoardMember, error)
    FindMembers(ctx, boardID string) ([]*BoardMember, error)
    UpdateMemberRole(ctx, boardID, userID string, role BoardRole) error
    RemoveMember(ctx, boardID, userID string) error
    
    // Access check
    CanUserAccess(ctx, boardID, userID string) (bool, BoardRole, error)
}
```

### Service Methods

```go
type BoardService interface {
    // Boards
    Create(ctx, userID, orgSlug string, req CreateBoardRequest) (*Board, error)
    GetByID(ctx, userID, boardID string) (*BoardDetail, error)
    ListByOrg(ctx, userID, orgSlug string) ([]*BoardSummary, error)
    Update(ctx, userID, boardID string, req UpdateBoardRequest) (*Board, error)
    Close(ctx, userID, boardID string) error
    Reopen(ctx, userID, boardID string) error
    Delete(ctx, userID, boardID string) error
    
    // Members
    ListMembers(ctx, userID, boardID string) ([]*MemberResponse, error)
    UpdateMemberRole(ctx, userID, boardID, targetUserID string, role BoardRole) error
    RemoveMember(ctx, userID, boardID, targetUserID string) error
    LeaveBoard(ctx, userID, boardID string) error
}
```

### API Endpoints (11 endpoints)

| Method | Endpoint | Auth | Permission | Description |
|--------|----------|:----:|------------|-------------|
| GET | `/organizations/:slug/boards` | Yes | org member+ | List boards in org |
| POST | `/organizations/:slug/boards` | Yes | org member+ | Create board |
| GET | `/boards/:id` | Yes | board viewer+ | Get board with lists & cards |
| PUT | `/boards/:id` | Yes | board admin+ | Update board |
| DELETE | `/boards/:id` | Yes | board owner | Delete board |
| POST | `/boards/:id/close` | Yes | board admin+ | Close board |
| POST | `/boards/:id/reopen` | Yes | board admin+ | Reopen board |
| GET | `/boards/:id/members` | Yes | board viewer+ | List board members |
| POST | `/boards/:id/invite` | Yes | board admin+ | Invite to board |
| PUT | `/boards/:id/members/:userId` | Yes | board admin+ | Update member role |
| DELETE | `/boards/:id/members/:userId` | Yes | board admin+ | Remove member |

### Permission Matrix (Board)

| Action | Owner | Admin | Member | Viewer |
|--------|:-----:|:-----:|:------:|:------:|
| View board | ✅ | ✅ | ✅ | ✅ |
| Create/edit cards | ✅ | ✅ | ✅ | ❌ |
| Create/edit lists | ✅ | ✅ | ✅ | ❌ |
| Manage labels | ✅ | ✅ | ❌ | ❌ |
| Invite members | ✅ | ✅ | ❌ | ❌ |
| Edit board settings | ✅ | ✅ | ❌ | ❌ |
| Close/reopen board | ✅ | ✅ | ❌ | ❌ |
| Delete board | ✅ | ❌ | ❌ | ❌ |

### Visibility Rules

| Visibility | Who Can Access |
|------------|----------------|
| `private` | Only board members |
| `workspace` | All organization members |
| `public` | Anyone with the link (read-only for non-members) |

### Request/Response DTOs

```go
// Request
type CreateBoardRequest struct {
    Title           string          `json:"title" validate:"required,min=1,max=255"`
    Description     *string         `json:"description" validate:"omitempty,max=1000"`
    BackgroundColor string          `json:"background_color" validate:"omitempty,hexcolor"`
    Visibility      BoardVisibility `json:"visibility" validate:"omitempty,oneof=private workspace public"`
}

type UpdateBoardRequest struct {
    Title           *string          `json:"title" validate:"omitempty,min=1,max=255"`
    Description     *string          `json:"description" validate:"omitempty,max=1000"`
    BackgroundColor *string          `json:"background_color" validate:"omitempty,hexcolor"`
    Visibility      *BoardVisibility `json:"visibility" validate:"omitempty,oneof=private workspace public"`
}

// Response
type BoardSummary struct {
    ID              string          `json:"id"`
    Title           string          `json:"title"`
    BackgroundColor string          `json:"background_color"`
    Visibility      BoardVisibility `json:"visibility"`
    IsClosed        bool            `json:"is_closed"`
    ListsCount      int             `json:"lists_count"`
    CardsCount      int             `json:"cards_count"`
    MyRole          BoardRole       `json:"my_role"`
}

type BoardDetail struct {
    ID              string          `json:"id"`
    Title           string          `json:"title"`
    Description     *string         `json:"description"`
    BackgroundColor string          `json:"background_color"`
    Visibility      BoardVisibility `json:"visibility"`
    IsClosed        bool            `json:"is_closed"`
    MyRole          BoardRole       `json:"my_role"`
    Organization    OrgSummary      `json:"organization"`
    Lists           []ListWithCards `json:"lists"`   // Empty for now (Phase 3)
    Labels          []Label         `json:"labels"`  // Empty for now (Phase 3)
    Members         []MemberSummary `json:"members"`
    CreatedAt       time.Time       `json:"created_at"`
}
```

### Acceptance Criteria

- [ ] Create board within organization
- [ ] Board visibility controls access correctly
- [ ] Default background color: `#0079bf`
- [ ] Close/reopen board works
- [ ] Closed boards still viewable but not editable
- [ ] 4-level role permissions enforced
- [ ] Board owner cannot be removed

---

## Task 2.3: Invitation Module

**Duration:** 2 days

### Files to Create

```
internal/
├── domain/
│   └── board_invitation.go
├── repository/
│   └── invitation_repository.go
├── service/
│   └── invitation_service.go
└── handler/
    └── invitation_handler.go

migrations/
└── 00008_create_board_invitations.sql
```

### Domain Model

```go
type InvitationStatus string

const (
    InvitationPending  InvitationStatus = "pending"
    InvitationAccepted InvitationStatus = "accepted"
    InvitationDeclined InvitationStatus = "declined"
    InvitationExpired  InvitationStatus = "expired"
)

type BoardInvitation struct {
    ID           string
    BoardID      string
    InviterID    string
    InviteeID    *string          // NULL if user doesn't exist yet
    InviteeEmail string
    Role         BoardRole
    Token        string           // 64-char hex token
    Message      *string
    Status       InvitationStatus
    ExpiresAt    time.Time
    CreatedAt    time.Time
    RespondedAt  *time.Time
    
    // Joined data
    Board        *Board
    Inviter      *User
}
```

### Repository Interface

```go
type InvitationRepository interface {
    Create(ctx, inv *BoardInvitation) error
    FindByToken(ctx, token string) (*BoardInvitation, error)
    FindByBoardAndEmail(ctx, boardID, email string) (*BoardInvitation, error)
    FindPendingByBoardID(ctx, boardID string) ([]*BoardInvitation, error)
    UpdateStatus(ctx, id string, status InvitationStatus) error
    ExpireOldInvitations(ctx) error  // Batch job
}
```

### Service Methods

```go
type InvitationService interface {
    // Create invitation (called from BoardService.Invite)
    CreateInvitation(ctx, boardID, inviterID, email string, role BoardRole, message *string) (*BoardInvitation, error)
    
    // Public endpoints
    GetByToken(ctx, token string) (*InvitationDetail, error)
    Accept(ctx, userID, token string) error
    Decline(ctx, userID, token string) error
}
```

### API Endpoints (3 endpoints)

| Method | Endpoint | Auth | Description |
|--------|----------|:----:|-------------|
| GET | `/invitations/:token` | Optional | Get invitation details |
| POST | `/invitations/:token/accept` | Yes | Accept invitation |
| POST | `/invitations/:token/decline` | Yes | Decline invitation |

### Invitation Flow

```
1. Admin invites email@example.com to Board
   └── Create invitation with 64-char token, expires in 7 days
   └── Send email with link: {FRONTEND_URL}/invite/{token}

2. User clicks link
   └── If not logged in → redirect to login with ?redirect=/invite/{token}
   └── If logged in → show invitation details

3. User accepts
   └── If email matches user's email → add to board
   └── If email doesn't match → error

4. User declines
   └── Mark invitation as declined
```

### Acceptance Criteria

- [ ] Generate secure 64-char hex token
- [ ] Invitation expires after 7 days
- [ ] Email is sent to invitee
- [ ] Existing user can accept immediately
- [ ] Non-existing user sees "register first" message
- [ ] Can't accept invitation for different email
- [ ] Declined invitations can't be accepted
- [ ] Expired invitations return error

---

## Permission Middleware

### Implementation

```go
// internal/middleware/permission.go

// Organization permission check
func RequireOrgRole(minRole OrgRole) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("userID")
        slug := c.Param("slug")
        
        member, err := orgRepo.FindMemberBySlug(ctx, slug, userID)
        if err != nil {
            c.AbortWithStatusJSON(403, ErrForbidden)
            return
        }
        
        if !hasMinRole(member.Role, minRole) {
            c.AbortWithStatusJSON(403, ErrInsufficientRole)
            return
        }
        
        c.Set("orgMember", member)
        c.Set("organization", member.Organization)
        c.Next()
    }
}

// Board permission check
func RequireBoardRole(minRole BoardRole) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("userID")
        boardID := c.Param("id")
        
        canAccess, role, err := boardRepo.CanUserAccess(ctx, boardID, userID)
        if err != nil || !canAccess {
            c.AbortWithStatusJSON(403, ErrNotBoardMember)
            return
        }
        
        if !hasMinBoardRole(role, minRole) {
            c.AbortWithStatusJSON(403, ErrInsufficientRole)
            return
        }
        
        c.Set("boardRole", role)
        c.Next()
    }
}

// Role hierarchy helpers
func hasMinRole(actual, required OrgRole) bool {
    hierarchy := map[OrgRole]int{"owner": 3, "admin": 2, "member": 1}
    return hierarchy[actual] >= hierarchy[required]
}

func hasMinBoardRole(actual, required BoardRole) bool {
    hierarchy := map[BoardRole]int{"owner": 4, "admin": 3, "member": 2, "viewer": 1}
    return hierarchy[actual] >= hierarchy[required]
}
```

---

## Dependency Graph

```
Phase 1 (Auth) ─────────────────────────────────────┐
                                                    │
Task 2.1 (Organizations)                            │
    │                                               │
    ├── migrations (00004, 00005)                   │
    ├── domain/organization.go                      │
    ├── repository/organization_repository.go       │
    ├── service/organization_service.go             │
    └── handler/organization_handler.go             │
            │                                       │
            ▼                                       │
Task 2.2 (Boards) ◀─────────────────────────────────┘
    │
    ├── migrations (00006, 00007)
    ├── domain/board.go
    ├── repository/board_repository.go
    ├── service/board_service.go
    └── handler/board_handler.go
            │
            ▼
Task 2.3 (Invitations)
    │
    ├── migrations (00008)
    ├── domain/board_invitation.go
    ├── service/invitation_service.go
    └── handler/invitation_handler.go
```

---

## Recommended Execution Order

| Day | Tasks | Focus |
|:---:|-------|-------|
| 1 | Migrations | 00004, 00005, 00006, 00007, 00008 |
| 2 | Domain models | organization, board, invitation |
| 3 | Org repository | CRUD + members |
| 4 | Org service + handler | 9 endpoints |
| 5 | Board repository | CRUD + members + access check |
| 6 | Board service + handler | 11 endpoints |
| 7 | Invitation flow | Token generation, accept/decline |
| 8 | Permission middleware | Role checks, integration |
| 9 | Testing | Unit + integration tests |

---

## Phase 2 Deliverables Checklist

### Organizations
- [ ] Create organization with auto-generated slug
- [ ] List user's organizations with role
- [ ] Get organization details
- [ ] Update organization (admin+)
- [ ] Delete organization (owner only)
- [ ] List organization members
- [ ] Invite member (admin+)
- [ ] Update member role (admin+)
- [ ] Remove member (admin+)

### Boards
- [ ] Create board in organization
- [ ] List boards in organization
- [ ] Get board details (empty lists/cards for now)
- [ ] Update board settings (admin+)
- [ ] Close/reopen board (admin+)
- [ ] Delete board (owner only)
- [ ] Manage board members

### Invitations
- [ ] Generate invitation with token
- [ ] Send invitation email
- [ ] View invitation by token
- [ ] Accept invitation
- [ ] Decline invitation
- [ ] Expired invitations rejected

### Security
- [ ] Organization role checks work
- [ ] Board role checks work
- [ ] Visibility rules enforced
- [ ] Owner protection (can't remove/leave)

### Tests
- [ ] Unit tests for services (80% coverage)
- [ ] Integration tests for permission checks
