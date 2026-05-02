package service

import (
	"context"
	"time"

	"github.com/codewebkhongkho/trello-agent/internal/domain"
	"github.com/codewebkhongkho/trello-agent/internal/dto/request"
	"github.com/codewebkhongkho/trello-agent/internal/repository"
	"github.com/codewebkhongkho/trello-agent/pkg/apperror"
	"github.com/codewebkhongkho/trello-agent/pkg/cuid"
	"github.com/codewebkhongkho/trello-agent/pkg/position"
)

type ChecklistService struct {
	checklistRepo *repository.ChecklistRepository
	cardRepo      *repository.CardRepository
	boardRepo     repository.BoardRepository
	activityRepo  *repository.ActivityRepository
}

func NewChecklistService(
	checklistRepo *repository.ChecklistRepository,
	cardRepo *repository.CardRepository,
	boardRepo repository.BoardRepository,
	activityRepo *repository.ActivityRepository,
) *ChecklistService {
	return &ChecklistService{
		checklistRepo: checklistRepo,
		cardRepo:      cardRepo,
		boardRepo:     boardRepo,
		activityRepo:  activityRepo,
	}
}

func (s *ChecklistService) ListByCard(ctx context.Context, userID, cardID string) ([]*domain.Checklist, error) {
	boardID, err := s.cardRepo.FindBoardIDByCardID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	canAccess, _, err := s.boardRepo.CanUserAccess(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, apperror.ErrForbidden
	}

	return s.checklistRepo.FindByCardIDWithItems(ctx, cardID)
}

func (s *ChecklistService) Create(ctx context.Context, userID string, req request.CreateChecklistRequest) (*domain.Checklist, error) {
	boardID, err := s.cardRepo.FindBoardIDByCardID(ctx, req.CardID)
	if err != nil {
		return nil, err
	}

	canAccess, role, err := s.boardRepo.CanUserAccess(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !canAccess || !role.CanEdit() {
		return nil, apperror.ErrForbidden
	}

	maxPos, err := s.checklistRepo.GetMaxPosition(ctx, req.CardID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	checklist := &domain.Checklist{
		ID:        cuid.New(),
		CardID:    req.CardID,
		Title:     req.Title,
		Position:  position.NextPosition(maxPos),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.checklistRepo.Create(ctx, checklist); err != nil {
		return nil, err
	}

	s.logActivity(ctx, boardID, &req.CardID, userID, domain.ActivityChecklistCreated, map[string]interface{}{
		"checklist_id":    checklist.ID,
		"checklist_title": checklist.Title,
	})

	checklist.Items = []domain.ChecklistItem{}
	checklist.Progress = &domain.ChecklistProgress{Completed: 0, Total: 0}
	return checklist, nil
}

func (s *ChecklistService) Update(ctx context.Context, userID, checklistID string, req request.UpdateChecklistRequest) (*domain.Checklist, error) {
	checklist, err := s.checklistRepo.FindByID(ctx, checklistID)
	if err != nil {
		return nil, err
	}

	boardID, err := s.cardRepo.FindBoardIDByCardID(ctx, checklist.CardID)
	if err != nil {
		return nil, err
	}

	canAccess, role, err := s.boardRepo.CanUserAccess(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !canAccess || !role.CanEdit() {
		return nil, apperror.ErrForbidden
	}

	if req.Title != nil {
		checklist.Title = *req.Title
	}

	if err := s.checklistRepo.Update(ctx, checklist); err != nil {
		return nil, err
	}

	return s.checklistRepo.FindByID(ctx, checklistID)
}

func (s *ChecklistService) Delete(ctx context.Context, userID, checklistID string) error {
	checklist, err := s.checklistRepo.FindByID(ctx, checklistID)
	if err != nil {
		return err
	}

	boardID, err := s.cardRepo.FindBoardIDByCardID(ctx, checklist.CardID)
	if err != nil {
		return err
	}

	canAccess, role, err := s.boardRepo.CanUserAccess(ctx, boardID, userID)
	if err != nil {
		return err
	}
	if !canAccess || !role.CanEdit() {
		return apperror.ErrForbidden
	}

	if err := s.checklistRepo.Delete(ctx, checklistID); err != nil {
		return err
	}

	s.logActivity(ctx, boardID, &checklist.CardID, userID, domain.ActivityChecklistDeleted, map[string]interface{}{
		"checklist_title": checklist.Title,
	})

	return nil
}

func (s *ChecklistService) CreateItem(ctx context.Context, userID string, req request.CreateChecklistItemRequest) (*domain.ChecklistItem, error) {
	checklist, err := s.checklistRepo.FindByID(ctx, req.ChecklistID)
	if err != nil {
		return nil, err
	}

	boardID, err := s.cardRepo.FindBoardIDByCardID(ctx, checklist.CardID)
	if err != nil {
		return nil, err
	}

	canAccess, role, err := s.boardRepo.CanUserAccess(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !canAccess || !role.CanEdit() {
		return nil, apperror.ErrForbidden
	}

	maxPos, err := s.checklistRepo.GetItemMaxPosition(ctx, req.ChecklistID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	item := &domain.ChecklistItem{
		ID:          cuid.New(),
		ChecklistID: req.ChecklistID,
		Title:       req.Title,
		Position:    position.NextPosition(maxPos),
		AssigneeID:  req.AssigneeID,
		DueDate:     req.DueDate,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.checklistRepo.CreateItem(ctx, item); err != nil {
		return nil, err
	}

	return s.checklistRepo.FindItemByID(ctx, item.ID)
}

func (s *ChecklistService) UpdateItem(ctx context.Context, userID, itemID string, req request.UpdateChecklistItemRequest) (*domain.ChecklistItem, error) {
	item, err := s.checklistRepo.FindItemByID(ctx, itemID)
	if err != nil {
		return nil, err
	}

	cardID, err := s.checklistRepo.FindCardIDByChecklistID(ctx, item.ChecklistID)
	if err != nil {
		return nil, err
	}

	boardID, err := s.cardRepo.FindBoardIDByCardID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	canAccess, role, err := s.boardRepo.CanUserAccess(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !canAccess || !role.CanEdit() {
		return nil, apperror.ErrForbidden
	}

	if req.Title != nil {
		item.Title = *req.Title
	}
	if req.AssigneeID != nil {
		item.AssigneeID = req.AssigneeID
	}
	if req.DueDate != nil {
		item.DueDate = req.DueDate
	}

	if err := s.checklistRepo.UpdateItem(ctx, item); err != nil {
		return nil, err
	}

	return s.checklistRepo.FindItemByID(ctx, itemID)
}

func (s *ChecklistService) DeleteItem(ctx context.Context, userID, itemID string) error {
	item, err := s.checklistRepo.FindItemByID(ctx, itemID)
	if err != nil {
		return err
	}

	cardID, err := s.checklistRepo.FindCardIDByChecklistID(ctx, item.ChecklistID)
	if err != nil {
		return err
	}

	boardID, err := s.cardRepo.FindBoardIDByCardID(ctx, cardID)
	if err != nil {
		return err
	}

	canAccess, role, err := s.boardRepo.CanUserAccess(ctx, boardID, userID)
	if err != nil {
		return err
	}
	if !canAccess || !role.CanEdit() {
		return apperror.ErrForbidden
	}

	return s.checklistRepo.DeleteItem(ctx, itemID)
}

func (s *ChecklistService) ToggleItem(ctx context.Context, userID, itemID string) (*domain.ChecklistItem, error) {
	item, err := s.checklistRepo.FindItemByID(ctx, itemID)
	if err != nil {
		return nil, err
	}

	cardID, err := s.checklistRepo.FindCardIDByChecklistID(ctx, item.ChecklistID)
	if err != nil {
		return nil, err
	}

	boardID, err := s.cardRepo.FindBoardIDByCardID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	canAccess, role, err := s.boardRepo.CanUserAccess(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !canAccess || !role.CanEdit() {
		return nil, apperror.ErrForbidden
	}

	updatedItem, err := s.checklistRepo.ToggleItemComplete(ctx, itemID, userID)
	if err != nil {
		return nil, err
	}

	action := domain.ActivityChecklistItemCompleted
	if !updatedItem.IsCompleted {
		action = domain.ActivityChecklistItemUncompleted
	}

	s.logActivity(ctx, boardID, &cardID, userID, action, map[string]interface{}{
		"item_title": updatedItem.Title,
	})

	return updatedItem, nil
}

func (s *ChecklistService) logActivity(ctx context.Context, boardID string, cardID *string, userID string, action domain.ActivityAction, metadata map[string]interface{}) {
	activity := &domain.ActivityLog{
		ID:        cuid.New(),
		BoardID:   boardID,
		CardID:    cardID,
		UserID:    userID,
		Action:    action,
		Metadata:  metadata,
		CreatedAt: time.Now(),
	}
	s.activityRepo.Create(ctx, activity)
}
