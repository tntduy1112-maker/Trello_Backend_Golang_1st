package service

import (
	"context"
	"time"

	"github.com/codewebkhongkho/trello-agent/internal/domain"
	"github.com/codewebkhongkho/trello-agent/internal/dto/request"
	"github.com/codewebkhongkho/trello-agent/internal/repository"
	"github.com/codewebkhongkho/trello-agent/pkg/apperror"
	"github.com/codewebkhongkho/trello-agent/pkg/cuid"
)

type LabelService struct {
	labelRepo *repository.LabelRepository
	boardRepo repository.BoardRepository
	cardRepo  *repository.CardRepository
}

func NewLabelService(
	labelRepo *repository.LabelRepository,
	boardRepo repository.BoardRepository,
	cardRepo *repository.CardRepository,
) *LabelService {
	return &LabelService{
		labelRepo: labelRepo,
		boardRepo: boardRepo,
		cardRepo:  cardRepo,
	}
}

func (s *LabelService) Create(ctx context.Context, userID, boardID string, req request.CreateLabelRequest) (*domain.Label, error) {
	canAccess, role, err := s.boardRepo.CanUserAccess(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !canAccess || !role.CanManage() {
		return nil, apperror.ErrForbidden
	}

	now := time.Now()
	label := &domain.Label{
		ID:        cuid.New(),
		BoardID:   boardID,
		Name:      req.Name,
		Color:     req.Color,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.labelRepo.Create(ctx, label); err != nil {
		return nil, err
	}

	return label, nil
}

func (s *LabelService) GetByBoardID(ctx context.Context, userID, boardID string) ([]*domain.Label, error) {
	canAccess, _, err := s.boardRepo.CanUserAccess(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, apperror.ErrForbidden
	}

	return s.labelRepo.FindByBoardID(ctx, boardID)
}

func (s *LabelService) Update(ctx context.Context, userID, labelID string, req request.UpdateLabelRequest) (*domain.Label, error) {
	boardID, err := s.labelRepo.FindBoardIDByLabelID(ctx, labelID)
	if err != nil {
		return nil, err
	}

	canAccess, role, err := s.boardRepo.CanUserAccess(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !canAccess || !role.CanManage() {
		return nil, apperror.ErrForbidden
	}

	label, err := s.labelRepo.FindByID(ctx, labelID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		label.Name = req.Name
	}
	if req.Color != nil {
		label.Color = *req.Color
	}

	if err := s.labelRepo.Update(ctx, label); err != nil {
		return nil, err
	}

	return label, nil
}

func (s *LabelService) Delete(ctx context.Context, userID, labelID string) error {
	boardID, err := s.labelRepo.FindBoardIDByLabelID(ctx, labelID)
	if err != nil {
		return err
	}

	canAccess, role, err := s.boardRepo.CanUserAccess(ctx, boardID, userID)
	if err != nil {
		return err
	}
	if !canAccess || !role.CanManage() {
		return apperror.ErrForbidden
	}

	return s.labelRepo.Delete(ctx, labelID)
}

func (s *LabelService) AssignToCard(ctx context.Context, userID, cardID, labelID string) error {
	cardBoardID, err := s.cardRepo.FindBoardIDByCardID(ctx, cardID)
	if err != nil {
		return err
	}

	canAccess, role, err := s.boardRepo.CanUserAccess(ctx, cardBoardID, userID)
	if err != nil {
		return err
	}
	if !canAccess || !role.CanEdit() {
		return apperror.ErrForbidden
	}

	labelBoardID, err := s.labelRepo.FindBoardIDByLabelID(ctx, labelID)
	if err != nil {
		return err
	}

	if cardBoardID != labelBoardID {
		return apperror.New("BAD_REQUEST", "Label does not belong to the same board as the card", 400)
	}

	return s.labelRepo.AssignToCard(ctx, cardID, labelID)
}

func (s *LabelService) RemoveFromCard(ctx context.Context, userID, cardID, labelID string) error {
	cardBoardID, err := s.cardRepo.FindBoardIDByCardID(ctx, cardID)
	if err != nil {
		return err
	}

	canAccess, role, err := s.boardRepo.CanUserAccess(ctx, cardBoardID, userID)
	if err != nil {
		return err
	}
	if !canAccess || !role.CanEdit() {
		return apperror.ErrForbidden
	}

	return s.labelRepo.RemoveFromCard(ctx, cardID, labelID)
}
