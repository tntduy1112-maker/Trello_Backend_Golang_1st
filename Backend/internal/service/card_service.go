package service

import (
	"context"
	"time"

	"github.com/codewebkhongkho/trello-agent/internal/domain"
	"github.com/codewebkhongkho/trello-agent/internal/dto/request"
	"github.com/codewebkhongkho/trello-agent/internal/dto/response"
	"github.com/codewebkhongkho/trello-agent/internal/repository"
	"github.com/codewebkhongkho/trello-agent/pkg/apperror"
	"github.com/codewebkhongkho/trello-agent/pkg/cuid"
	"github.com/codewebkhongkho/trello-agent/pkg/position"
)

type CardService struct {
	cardRepo            *repository.CardRepository
	listRepo            *repository.ListRepository
	boardRepo           repository.BoardRepository
	labelRepo           *repository.LabelRepository
	userRepo            repository.UserRepository
	notificationService *NotificationService
	activityService     *ActivityService
}

func NewCardService(
	cardRepo *repository.CardRepository,
	listRepo *repository.ListRepository,
	boardRepo repository.BoardRepository,
	labelRepo *repository.LabelRepository,
	userRepo repository.UserRepository,
	notificationService *NotificationService,
	activityService *ActivityService,
) *CardService {
	return &CardService{
		cardRepo:            cardRepo,
		listRepo:            listRepo,
		boardRepo:           boardRepo,
		labelRepo:           labelRepo,
		userRepo:            userRepo,
		notificationService: notificationService,
		activityService:     activityService,
	}
}

func (s *CardService) Create(ctx context.Context, userID, listID string, req request.CreateCardRequest) (*domain.Card, error) {
	boardID, err := s.listRepo.FindBoardIDByListID(ctx, listID)
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

	maxPos, err := s.cardRepo.GetMaxPosition(ctx, listID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	card := &domain.Card{
		ID:        cuid.New(),
		ListID:    listID,
		Title:     req.Title,
		Position:  position.NextPosition(maxPos),
		Priority:  domain.PriorityNone,
		CreatedBy: userID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.cardRepo.Create(ctx, card); err != nil {
		return nil, err
	}

	if s.activityService != nil {
		_ = s.activityService.Log(ctx, LogActivityRequest{
			BoardID: boardID,
			CardID:  &card.ID,
			ListID:  &listID,
			UserID:  userID,
			Action:  domain.ActivityCardCreated,
			Metadata: map[string]interface{}{
				"card_title": card.Title,
			},
		})
	}

	return card, nil
}

func (s *CardService) GetByID(ctx context.Context, userID, cardID string) (*response.CardDetail, error) {
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

	card, err := s.cardRepo.FindByIDWithDetails(ctx, cardID)
	if err != nil {
		return nil, err
	}

	list, err := s.listRepo.FindByID(ctx, card.ListID)
	if err != nil {
		return nil, err
	}

	labels := make([]response.LabelSummary, 0)
	if card.Labels != nil {
		for _, l := range card.Labels {
			labels = append(labels, response.LabelSummary{
				ID:    l.ID,
				Name:  l.Name,
				Color: l.Color,
			})
		}
	}

	var assignee *response.UserSummary
	if card.Assignee != nil {
		assignee = &response.UserSummary{
			ID:        card.Assignee.ID,
			FullName:  card.Assignee.FullName,
			AvatarURL: card.Assignee.AvatarURL,
		}
	}

	var reporter *response.UserSummary
	if card.CreatedBy != "" {
		reporterUser, err := s.userRepo.FindByID(ctx, card.CreatedBy)
		if err == nil && reporterUser != nil {
			reporter = &response.UserSummary{
				ID:        reporterUser.ID,
				Email:     reporterUser.Email,
				FullName:  reporterUser.FullName,
				AvatarURL: reporterUser.AvatarURL,
			}
		}
	}

	return &response.CardDetail{
		ID:          card.ID,
		Title:       card.Title,
		Description: card.Description,
		Position:    card.Position,
		Priority:    card.Priority,
		DueDate:     card.DueDate,
		IsCompleted: card.IsCompleted,
		List: response.ListSummary{
			ID:    list.ID,
			Title: list.Title,
		},
		Assignee:    assignee,
		Reporter:    reporter,
		Labels:      labels,
		Checklists:  []interface{}{},
		Comments:    []interface{}{},
		Attachments: []interface{}{},
		Activity:    []interface{}{},
		CreatedAt:   card.CreatedAt,
		UpdatedAt:   card.UpdatedAt,
	}, nil
}

func (s *CardService) Update(ctx context.Context, userID, cardID string, req request.UpdateCardRequest) (*domain.Card, error) {
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

	card, err := s.cardRepo.FindByID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		card.Title = *req.Title
	}
	if req.Description != nil {
		card.Description = req.Description
	}
	if req.Priority != nil {
		card.Priority = *req.Priority
	}
	if req.DueDate != nil {
		card.DueDate = req.DueDate
	}

	if err := s.cardRepo.Update(ctx, card); err != nil {
		return nil, err
	}

	return card, nil
}

func (s *CardService) Archive(ctx context.Context, userID, cardID string) error {
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

	return s.cardRepo.Archive(ctx, cardID)
}

func (s *CardService) Restore(ctx context.Context, userID, cardID string) error {
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

	return s.cardRepo.Restore(ctx, cardID)
}

func (s *CardService) Move(ctx context.Context, userID, cardID string, req request.MoveCardRequest) (*domain.Card, error) {
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

	targetBoardID, err := s.listRepo.FindBoardIDByListID(ctx, req.ListID)
	if err != nil {
		return nil, err
	}
	if targetBoardID != boardID {
		return nil, apperror.New("BAD_REQUEST", "Cannot move card to a different board", 400)
	}

	if err := s.cardRepo.UpdatePosition(ctx, cardID, req.ListID, req.Position); err != nil {
		return nil, err
	}

	card, err := s.cardRepo.FindByID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	return card, nil
}

func (s *CardService) Assign(ctx context.Context, userID, cardID, assigneeID string) error {
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

	canAssign, _, err := s.boardRepo.CanUserAccess(ctx, boardID, assigneeID)
	if err != nil {
		return err
	}
	if !canAssign {
		return apperror.New("BAD_REQUEST", "Assignee is not a board member", 400)
	}

	if err := s.cardRepo.Assign(ctx, cardID, assigneeID); err != nil {
		return err
	}

	card, _ := s.cardRepo.FindByID(ctx, cardID)
	assignee, _ := s.userRepo.FindByID(ctx, assigneeID)

	// Log activity
	if s.activityService != nil && assignee != nil {
		_ = s.activityService.Log(ctx, LogActivityRequest{
			BoardID: boardID,
			CardID:  &cardID,
			UserID:  userID,
			Action:  domain.ActivityCardAssigned,
			Metadata: map[string]interface{}{
				"assignee_name": assignee.FullName,
				"assignee_id":   assigneeID,
			},
		})
	}

	// Send notification to assignee
	if s.notificationService != nil && userID != assigneeID {
		actor, _ := s.userRepo.FindByID(ctx, userID)
		if card != nil && actor != nil && assignee != nil {
			s.notificationService.NotifyCardAssigned(ctx, cardID, card.Title, boardID, assignee, actor)
		}
	}

	return nil
}

func (s *CardService) Unassign(ctx context.Context, userID, cardID string) error {
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

	if err := s.cardRepo.Unassign(ctx, cardID); err != nil {
		return err
	}

	// Log activity
	if s.activityService != nil {
		_ = s.activityService.Log(ctx, LogActivityRequest{
			BoardID: boardID,
			CardID:  &cardID,
			UserID:  userID,
			Action:  domain.ActivityCardUnassigned,
		})
	}

	return nil
}

func (s *CardService) MarkComplete(ctx context.Context, userID, cardID string) error {
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

	return s.cardRepo.MarkComplete(ctx, cardID)
}

func (s *CardService) MarkIncomplete(ctx context.Context, userID, cardID string) error {
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

	return s.cardRepo.MarkIncomplete(ctx, cardID)
}
