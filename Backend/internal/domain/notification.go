package domain

import "time"

type NotificationType string

const (
	NotificationCardAssigned          NotificationType = "card_assigned"
	NotificationCardDueSoon           NotificationType = "card_due_soon"
	NotificationCardOverdue           NotificationType = "card_overdue"
	NotificationCommentAdded          NotificationType = "comment_added"
	NotificationCommentReply          NotificationType = "comment_reply"
	NotificationMentioned             NotificationType = "mentioned"
	NotificationBoardInvitation       NotificationType = "board_invitation"
	NotificationChecklistItemAssigned NotificationType = "checklist_item_assigned"
	NotificationCardCompleted         NotificationType = "card_completed"
	NotificationMemberAddedToBoard    NotificationType = "member_added_to_board"
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

	Actor *User `json:"actor,omitempty"`
}
