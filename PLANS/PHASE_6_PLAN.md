# Phase 6: Shareable Board Invitation Links

> Estimated Time: 2-3 days

---

## Overview

A shareable link-based invitation system where board owners/admins can create invitation links with QR codes to share manually.

---

## Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                    INVITATION LINK FLOW                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Board Owner/Admin                                               │
│       │                                                          │
│       │ 1. Create Invitation                                     │
│       │    - Invitee email                                       │
│       │    - Role (member/admin)                                 │
│       │    - Expires in 3 days                                   │
│       ▼                                                          │
│  ┌─────────────────────────────────┐                            │
│  │  System generates:              │                            │
│  │  • Unique link                  │                            │
│  │  • QR Code                      │                            │
│  │                                 │                            │
│  │  https://app.com/invite/abc123  │                            │
│  │  [QR CODE IMAGE]                │                            │
│  │                                 │                            │
│  │  [Copy Link] [Download QR]      │                            │
│  └─────────────────────────────────┘                            │
│       │                                                          │
│       │ 2. Share link manually (chat, email, etc.)              │
│       ▼                                                          │
│  ┌─────────────────────────────────┐                            │
│  │  Invitee clicks link            │                            │
│  │                                 │                            │
│  │  ┌─────────┐    ┌─────────┐    │                            │
│  │  │ New User│    │Existing │    │                            │
│  │  │         │    │  User   │    │                            │
│  │  └────┬────┘    └────┬────┘    │                            │
│  │       │              │         │                            │
│  │  Register       Login/Auto     │                            │
│  │  (email         accept if      │                            │
│  │  pre-filled)    email matches  │                            │
│  │       │              │         │                            │
│  │       └──────┬───────┘         │                            │
│  │              │                 │                            │
│  │         Join Board             │                            │
│  │         with assigned role     │                            │
│  └─────────────────────────────────┘                            │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Database Schema

```sql
CREATE TABLE board_invitations (
    id VARCHAR(30) PRIMARY KEY,
    board_id VARCHAR(30) NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    token VARCHAR(64) UNIQUE NOT NULL,      -- secure random token
    invitee_email VARCHAR(255) NOT NULL,    -- who this invite is for
    role VARCHAR(20) NOT NULL DEFAULT 'member',
    invited_by VARCHAR(30) NOT NULL REFERENCES users(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, accepted, expired, revoked
    expires_at TIMESTAMP NOT NULL,          -- 3 days from creation
    accepted_at TIMESTAMP,
    accepted_by VARCHAR(30) REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    UNIQUE(board_id, invitee_email, status) -- prevent duplicate pending invites
);

CREATE INDEX idx_board_invitations_token ON board_invitations(token);
CREATE INDEX idx_board_invitations_board_id ON board_invitations(board_id);
CREATE INDEX idx_board_invitations_invitee_email ON board_invitations(invitee_email);
```

---

## API Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `POST /boards/:id/invitations` | Required | Create invitation, returns link + QR |
| `GET /boards/:id/invitations` | Required | List all invitations for board |
| `DELETE /invitations/:id` | Required | Revoke invitation |
| `GET /invitations/:token` | None | Get invitation details (public) |
| `POST /invitations/:token/accept` | Required | Accept invitation |

---

## Request/Response Examples

### Create Invitation

```http
POST /api/v1/boards/:id/invitations
Authorization: Bearer <token>
Content-Type: application/json

{
  "email": "alice@test.com",
  "role": "member"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "inv_abc123",
    "board_id": "brd_xyz",
    "token": "a1b2c3d4e5f6...",
    "invitee_email": "alice@test.com",
    "role": "member",
    "status": "pending",
    "expires_at": "2026-05-05T00:00:00Z",
    "invite_url": "http://localhost:5173/invite/a1b2c3d4e5f6...",
    "qr_code_url": "http://localhost:8080/api/v1/invitations/a1b2c3d4e5f6.../qr",
    "created_at": "2026-05-02T00:00:00Z",
    "invited_by": {
      "id": "user_123",
      "full_name": "Bob Smith"
    }
  }
}
```

### Get Invitation (Public)

```http
GET /api/v1/invitations/:token
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "inv_abc123",
    "board": {
      "id": "brd_xyz",
      "title": "Board to test Notification"
    },
    "invitee_email": "alice@test.com",
    "role": "member",
    "status": "pending",
    "expires_at": "2026-05-05T00:00:00Z",
    "invited_by": {
      "full_name": "Bob Smith"
    }
  }
}
```

### Accept Invitation

```http
POST /api/v1/invitations/:token/accept
Authorization: Bearer <token>
```

**Response:**
```json
{
  "success": true,
  "message": "Successfully joined the board",
  "data": {
    "board_id": "brd_xyz",
    "role": "member"
  }
}
```

---

## Frontend Components

| Component | Path | Description |
|-----------|------|-------------|
| `CreateInvitationModal` | `components/board/CreateInvitationModal.jsx` | Form + link/QR display |
| `InvitationLinkDisplay` | `components/board/InvitationLinkDisplay.jsx` | Copy button + QR code |
| `BoardInvitationsPanel` | `components/board/BoardInvitationsPanel.jsx` | List sent invitations |
| `AcceptInvitationPage` | `pages/invitations/AcceptInvitationPage.jsx` | Public landing page |

---

## Frontend Pages

### Accept Invitation Page (`/invite/:token`)

```
┌─────────────────────────────────────┐
│                                     │
│  You're invited to join             │
│  "Board to test Notification"       │
│                                     │
│  Invited by: Bob Smith              │
│  Role: Member                       │
│  Expires: May 5, 2026               │
│                                     │
│  ┌─────────────────────────────┐   │
│  │ This invitation is for:     │   │
│  │ alice@test.com              │   │
│  └─────────────────────────────┘   │
│                                     │
│  [Login to Accept]  [Register]      │
│                                     │
└─────────────────────────────────────┘
```

### States:
- **Valid invitation**: Show accept button
- **Expired**: Show "This invitation has expired"
- **Already accepted**: Show "Already a member" + link to board
- **Wrong email**: Show "This invitation is for another email"

---

## Validation Rules

1. **Create Invitation:**
   - User must be board owner or admin
   - Cannot invite existing board members
   - Cannot create duplicate pending invitations for same email
   - Email must be valid format

2. **Accept Invitation:**
   - Invitation must not be expired
   - Invitation must be in "pending" status
   - User's email must match invitee_email
   - User must not already be a board member

---

## Tasks Checklist

### Day 1: Backend

- [ ] Migration: Create/update `board_invitations` table
- [ ] Domain: `Invitation` struct
- [ ] Repository: `InvitationRepository`
- [ ] Service: `InvitationService`
  - [ ] `Create(boardID, email, role, invitedBy) -> Invitation`
  - [ ] `GetByToken(token) -> Invitation`
  - [ ] `ListByBoard(boardID) -> []Invitation`
  - [ ] `Accept(token, userID) -> error`
  - [ ] `Revoke(id, userID) -> error`
- [ ] Handler: `InvitationHandler`
- [ ] QR Code generation endpoint
- [ ] Routes registration

### Day 2: Frontend

- [ ] `CreateInvitationModal` component
- [ ] `InvitationLinkDisplay` component (copy + QR)
- [ ] `AcceptInvitationPage` page
- [ ] Update `InviteMemberModal` to use new flow
- [ ] Redux slice for invitations
- [ ] Service: `invitation.service.js`

### Day 3: Testing & Polish

- [ ] Test: Create invitation flow
- [ ] Test: Accept as existing user
- [ ] Test: Register via invitation
- [ ] Test: Expired invitation handling
- [ ] Test: Wrong email handling
- [ ] Test: Already member handling

---

## Security Considerations

1. **Token Security:**
   - Use cryptographically secure random tokens (32+ bytes)
   - Tokens should be URL-safe (base64url encoding)

2. **Rate Limiting:**
   - Limit invitation creation: 10/hour/user
   - Limit token lookups: 30/minute/IP

3. **Email Verification:**
   - Only allow accepting if logged-in user's email matches invitation email
   - This prevents invitation link hijacking

---

## Dependencies

- Phase 1: Authentication (user registration, login)
- Phase 2: Board management (board membership)
- Phase 5: Notifications (optional - notify on accept)
