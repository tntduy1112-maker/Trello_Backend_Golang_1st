package service

import (
	"context"
	"fmt"
	"time"

	"github.com/codewebkhongkho/trello-agent/internal/domain"
	"github.com/codewebkhongkho/trello-agent/internal/repository"
	"github.com/codewebkhongkho/trello-agent/pkg/apperror"
	"github.com/codewebkhongkho/trello-agent/pkg/cuid"
)

type ActivityService struct {
	activityRepo *repository.ActivityRepository
	cardRepo     *repository.CardRepository
	boardRepo    repository.BoardRepository
}

func NewActivityService(
	activityRepo *repository.ActivityRepository,
	cardRepo *repository.CardRepository,
	boardRepo repository.BoardRepository,
) *ActivityService {
	return &ActivityService{
		activityRepo: activityRepo,
		cardRepo:     cardRepo,
		boardRepo:    boardRepo,
	}
}

type LogActivityRequest struct {
	BoardID  string
	CardID   *string
	ListID   *string
	UserID   string
	Action   domain.ActivityAction
	Metadata map[string]interface{}
}

func (s *ActivityService) Log(ctx context.Context, req LogActivityRequest) error {
	activity := &domain.ActivityLog{
		ID:          cuid.New(),
		BoardID:     req.BoardID,
		CardID:      req.CardID,
		ListID:      req.ListID,
		UserID:      req.UserID,
		Action:      req.Action,
		Metadata:    req.Metadata,
		Description: s.GenerateDescription(req.Action, req.Metadata),
		CreatedAt:   time.Now(),
	}

	return s.activityRepo.Create(ctx, activity)
}

func (s *ActivityService) ListByCard(ctx context.Context, userID, cardID string, page, limit int) ([]*domain.ActivityLog, int, error) {
	boardID, err := s.cardRepo.FindBoardIDByCardID(ctx, cardID)
	if err != nil {
		return nil, 0, err
	}

	canAccess, _, err := s.boardRepo.CanUserAccess(ctx, boardID, userID)
	if err != nil {
		return nil, 0, err
	}
	if !canAccess {
		return nil, 0, apperror.ErrForbidden
	}

	offset := (page - 1) * limit
	activities, err := s.activityRepo.FindByCard(ctx, cardID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.activityRepo.CountByCard(ctx, cardID)
	if err != nil {
		return nil, 0, err
	}

	return activities, total, nil
}

func (s *ActivityService) ListByBoard(ctx context.Context, userID, boardID string, page, limit int) ([]*domain.ActivityLog, int, error) {
	canAccess, _, err := s.boardRepo.CanUserAccess(ctx, boardID, userID)
	if err != nil {
		return nil, 0, err
	}
	if !canAccess {
		return nil, 0, apperror.ErrForbidden
	}

	offset := (page - 1) * limit
	activities, err := s.activityRepo.FindByBoard(ctx, boardID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.activityRepo.CountByBoard(ctx, boardID)
	if err != nil {
		return nil, 0, err
	}

	return activities, total, nil
}

func (s *ActivityService) GenerateDescription(action domain.ActivityAction, metadata map[string]interface{}) string {
	switch action {
	case domain.ActivityCardCreated:
		return "created this card"
	case domain.ActivityCardUpdated:
		return "updated this card"
	case domain.ActivityCardMoved:
		from := getMetaString(metadata, "from_list")
		to := getMetaString(metadata, "to_list")
		return fmt.Sprintf("moved this card from %s to %s", from, to)
	case domain.ActivityCardArchived:
		return "archived this card"
	case domain.ActivityCardAssigned:
		assignee := getMetaString(metadata, "assignee_name")
		return fmt.Sprintf("assigned %s to this card", assignee)
	case domain.ActivityCardUnassigned:
		return "removed the assignee from this card"
	case domain.ActivityCardCompleted:
		return "marked this card as complete"
	case domain.ActivityCardReopened:
		return "marked this card as incomplete"
	case domain.ActivityCardDueDateSet:
		return "set the due date"
	case domain.ActivityCardDueDateRemoved:
		return "removed the due date"
	case domain.ActivityLabelAdded:
		label := getMetaString(metadata, "label_name")
		return fmt.Sprintf("added label %s", label)
	case domain.ActivityLabelRemoved:
		label := getMetaString(metadata, "label_name")
		return fmt.Sprintf("removed label %s", label)
	case domain.ActivityChecklistCreated:
		title := getMetaString(metadata, "checklist_title")
		return fmt.Sprintf("added checklist %s", title)
	case domain.ActivityChecklistDeleted:
		title := getMetaString(metadata, "checklist_title")
		return fmt.Sprintf("removed checklist %s", title)
	case domain.ActivityChecklistItemCompleted:
		item := getMetaString(metadata, "item_title")
		return fmt.Sprintf("completed %s", item)
	case domain.ActivityChecklistItemUncompleted:
		item := getMetaString(metadata, "item_title")
		return fmt.Sprintf("marked %s as incomplete", item)
	case domain.ActivityCommentAdded:
		return "commented on this card"
	case domain.ActivityCommentEdited:
		return "edited a comment"
	case domain.ActivityCommentDeleted:
		return "deleted a comment"
	case domain.ActivityAttachmentAdded:
		name := getMetaString(metadata, "attachment_name")
		return fmt.Sprintf("attached %s", name)
	case domain.ActivityAttachmentDeleted:
		name := getMetaString(metadata, "attachment_name")
		return fmt.Sprintf("removed attachment %s", name)
	case domain.ActivityCoverSet:
		return "set the cover image"
	case domain.ActivityCoverRemoved:
		return "removed the cover image"
	case domain.ActivityMemberAdded:
		return "joined this card"
	case domain.ActivityMemberRemoved:
		return "left this card"
	default:
		return string(action)
	}
}

func getMetaString(metadata map[string]interface{}, key string) string {
	if metadata == nil {
		return ""
	}
	if val, ok := metadata[key].(string); ok {
		return val
	}
	return ""
}
