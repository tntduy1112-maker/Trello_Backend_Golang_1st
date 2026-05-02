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

type ListService struct {
	listRepo  *repository.ListRepository
	boardRepo repository.BoardRepository
	cardRepo  *repository.CardRepository
}

func NewListService(
	listRepo *repository.ListRepository,
	boardRepo repository.BoardRepository,
	cardRepo *repository.CardRepository,
) *ListService {
	return &ListService{
		listRepo:  listRepo,
		boardRepo: boardRepo,
		cardRepo:  cardRepo,
	}
}

func (s *ListService) Create(ctx context.Context, userID, boardID string, req request.CreateListRequest) (*domain.List, error) {
	canAccess, role, err := s.boardRepo.CanUserAccess(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !canAccess || !role.CanEditLists() {
		return nil, apperror.ErrForbidden
	}

	maxPos, err := s.listRepo.GetMaxPosition(ctx, boardID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	list := &domain.List{
		ID:        cuid.New(),
		BoardID:   boardID,
		Title:     req.Title,
		Position:  position.NextPosition(maxPos),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.listRepo.Create(ctx, list); err != nil {
		return nil, err
	}

	return list, nil
}

func (s *ListService) Update(ctx context.Context, userID, listID string, req request.UpdateListRequest) (*domain.List, error) {
	list, err := s.listRepo.FindByID(ctx, listID)
	if err != nil {
		return nil, err
	}

	canAccess, role, err := s.boardRepo.CanUserAccess(ctx, list.BoardID, userID)
	if err != nil {
		return nil, err
	}
	if !canAccess || !role.CanEditLists() {
		return nil, apperror.ErrForbidden
	}

	list.Title = req.Title
	if err := s.listRepo.Update(ctx, list); err != nil {
		return nil, err
	}

	return list, nil
}

func (s *ListService) Archive(ctx context.Context, userID, listID string) error {
	list, err := s.listRepo.FindByID(ctx, listID)
	if err != nil {
		return err
	}

	canAccess, role, err := s.boardRepo.CanUserAccess(ctx, list.BoardID, userID)
	if err != nil {
		return err
	}
	if !canAccess || !role.CanEditLists() {
		return apperror.ErrForbidden
	}

	return s.listRepo.Archive(ctx, listID)
}

func (s *ListService) Restore(ctx context.Context, userID, listID string) error {
	list, err := s.listRepo.FindByID(ctx, listID)
	if err != nil {
		return err
	}

	canAccess, role, err := s.boardRepo.CanUserAccess(ctx, list.BoardID, userID)
	if err != nil {
		return err
	}
	if !canAccess || !role.CanEditLists() {
		return apperror.ErrForbidden
	}

	return s.listRepo.Restore(ctx, listID)
}

func (s *ListService) Move(ctx context.Context, userID, listID string, req request.MoveListRequest) (*domain.List, error) {
	list, err := s.listRepo.FindByID(ctx, listID)
	if err != nil {
		return nil, err
	}

	canAccess, role, err := s.boardRepo.CanUserAccess(ctx, list.BoardID, userID)
	if err != nil {
		return nil, err
	}
	if !canAccess || !role.CanEditLists() {
		return nil, apperror.ErrForbidden
	}

	if err := s.listRepo.UpdatePosition(ctx, listID, req.Position); err != nil {
		return nil, err
	}

	list.Position = req.Position
	return list, nil
}

func (s *ListService) Copy(ctx context.Context, userID, listID string, req request.CopyListRequest) (*domain.List, error) {
	sourceList, err := s.listRepo.FindByID(ctx, listID)
	if err != nil {
		return nil, err
	}

	canAccess, role, err := s.boardRepo.CanUserAccess(ctx, sourceList.BoardID, userID)
	if err != nil {
		return nil, err
	}
	if !canAccess || !role.CanEditLists() {
		return nil, apperror.ErrForbidden
	}

	maxPos, err := s.listRepo.GetMaxPosition(ctx, sourceList.BoardID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	newList := &domain.List{
		ID:        cuid.New(),
		BoardID:   sourceList.BoardID,
		Title:     req.Title,
		Position:  position.NextPosition(maxPos),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.listRepo.Create(ctx, newList); err != nil {
		return nil, err
	}

	cards, err := s.cardRepo.FindByListID(ctx, listID, false)
	if err != nil {
		return nil, err
	}

	for _, card := range cards {
		newCard := &domain.Card{
			ID:          cuid.New(),
			ListID:      newList.ID,
			Title:       card.Title,
			Description: card.Description,
			Position:    card.Position,
			Priority:    card.Priority,
			DueDate:     card.DueDate,
			CreatedBy:   userID,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := s.cardRepo.Create(ctx, newCard); err != nil {
			return nil, err
		}
	}

	return newList, nil
}
