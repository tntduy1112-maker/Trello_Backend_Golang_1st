package domain

import "time"

type ActivityAction string

const (
	ActivityCardCreated             ActivityAction = "card_created"
	ActivityCardUpdated             ActivityAction = "card_updated"
	ActivityCardMoved               ActivityAction = "card_moved"
	ActivityCardArchived            ActivityAction = "card_archived"
	ActivityCardDeleted             ActivityAction = "card_deleted"
	ActivityCardAssigned            ActivityAction = "card_assigned"
	ActivityCardUnassigned          ActivityAction = "card_unassigned"
	ActivityCardCompleted           ActivityAction = "card_completed"
	ActivityCardReopened            ActivityAction = "card_reopened"
	ActivityCardDueDateSet          ActivityAction = "card_due_date_set"
	ActivityCardDueDateRemoved      ActivityAction = "card_due_date_removed"
	ActivityListCreated             ActivityAction = "list_created"
	ActivityListRenamed             ActivityAction = "list_renamed"
	ActivityListMoved               ActivityAction = "list_moved"
	ActivityListArchived            ActivityAction = "list_archived"
	ActivityLabelAdded              ActivityAction = "label_added"
	ActivityLabelRemoved            ActivityAction = "label_removed"
	ActivityChecklistCreated        ActivityAction = "checklist_created"
	ActivityChecklistDeleted        ActivityAction = "checklist_deleted"
	ActivityChecklistItemCompleted  ActivityAction = "checklist_item_completed"
	ActivityChecklistItemUncompleted ActivityAction = "checklist_item_uncompleted"
	ActivityCommentAdded            ActivityAction = "comment_added"
	ActivityCommentEdited           ActivityAction = "comment_edited"
	ActivityCommentDeleted          ActivityAction = "comment_deleted"
	ActivityAttachmentAdded         ActivityAction = "attachment_added"
	ActivityAttachmentDeleted       ActivityAction = "attachment_deleted"
	ActivityCoverSet                ActivityAction = "cover_set"
	ActivityCoverRemoved            ActivityAction = "cover_removed"
	ActivityMemberAdded             ActivityAction = "member_added"
	ActivityMemberRemoved           ActivityAction = "member_removed"
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

	User *User `json:"user,omitempty"`
}
