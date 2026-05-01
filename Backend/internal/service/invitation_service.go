package service

import (
	"context"

	"github.com/codewebkhongkho/trello-agent/internal/domain"
	"github.com/codewebkhongkho/trello-agent/internal/dto/response"
	"github.com/codewebkhongkho/trello-agent/internal/repository"
	"github.com/codewebkhongkho/trello-agent/pkg/apperror"
)

type InvitationService struct {
	invRepo   repository.InvitationRepository
	boardRepo repository.BoardRepository
	userRepo  repository.UserRepository
}

func NewInvitationService(invRepo repository.InvitationRepository, boardRepo repository.BoardRepository, userRepo repository.UserRepository) *InvitationService {
	return &InvitationService{
		invRepo:   invRepo,
		boardRepo: boardRepo,
		userRepo:  userRepo,
	}
}

func (s *InvitationService) GetByToken(ctx context.Context, token string) (*response.InvitationResponse, error) {
	inv, err := s.invRepo.FindByToken(ctx, token)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}
	if inv == nil {
		return nil, apperror.ErrNotFound
	}

	if inv.IsExpired() && inv.Status == domain.InvitationPending {
		_ = s.invRepo.UpdateStatus(ctx, inv.ID, domain.InvitationExpired)
		inv.Status = domain.InvitationExpired
	}

	return response.ToInvitationResponse(inv), nil
}

func (s *InvitationService) Accept(ctx context.Context, userID, token string) error {
	inv, err := s.invRepo.FindByToken(ctx, token)
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}
	if inv == nil {
		return apperror.ErrNotFound
	}

	if inv.IsExpired() {
		_ = s.invRepo.UpdateStatus(ctx, inv.ID, domain.InvitationExpired)
		return apperror.New("INVITATION_EXPIRED", "This invitation has expired", 400)
	}

	if inv.Status != domain.InvitationPending {
		return apperror.New("INVITATION_NOT_PENDING", "This invitation is no longer pending", 400)
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}
	if user == nil {
		return apperror.ErrUnauthorized
	}

	if user.Email != inv.InviteeEmail {
		return apperror.New("EMAIL_MISMATCH", "This invitation was sent to a different email address", 403)
	}

	existingMember, _ := s.boardRepo.FindMember(ctx, inv.BoardID, userID)
	if existingMember != nil {
		_ = s.invRepo.UpdateStatus(ctx, inv.ID, domain.InvitationAccepted)
		return nil
	}

	member := &domain.BoardMember{
		BoardID: inv.BoardID,
		UserID:  userID,
		Role:    inv.Role,
	}
	if err := s.boardRepo.AddMember(ctx, member); err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}

	if err := s.invRepo.UpdateStatus(ctx, inv.ID, domain.InvitationAccepted); err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}

	return nil
}

func (s *InvitationService) Decline(ctx context.Context, userID, token string) error {
	inv, err := s.invRepo.FindByToken(ctx, token)
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}
	if inv == nil {
		return apperror.ErrNotFound
	}

	if inv.Status != domain.InvitationPending {
		return apperror.New("INVITATION_NOT_PENDING", "This invitation is no longer pending", 400)
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}
	if user == nil {
		return apperror.ErrUnauthorized
	}

	if user.Email != inv.InviteeEmail {
		return apperror.New("EMAIL_MISMATCH", "This invitation was sent to a different email address", 403)
	}

	return s.invRepo.UpdateStatus(ctx, inv.ID, domain.InvitationDeclined)
}
