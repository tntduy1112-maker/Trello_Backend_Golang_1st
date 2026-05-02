package service

import (
	"context"
	"time"

	"github.com/codewebkhongkho/trello-agent/internal/domain"
	"github.com/codewebkhongkho/trello-agent/internal/dto/request"
	"github.com/codewebkhongkho/trello-agent/internal/dto/response"
	"github.com/codewebkhongkho/trello-agent/internal/repository"
	"github.com/codewebkhongkho/trello-agent/pkg/apperror"
	"github.com/codewebkhongkho/trello-agent/pkg/hash"
	"github.com/codewebkhongkho/trello-agent/pkg/jwt"
)

type InvitationService struct {
	invRepo    repository.InvitationRepository
	boardRepo  repository.BoardRepository
	userRepo   repository.UserRepository
	orgRepo    repository.OrganizationRepository
	jwtManager *jwt.Manager
}

func NewInvitationService(invRepo repository.InvitationRepository, boardRepo repository.BoardRepository, userRepo repository.UserRepository, orgRepo repository.OrganizationRepository, jwtManager *jwt.Manager) *InvitationService {
	return &InvitationService{
		invRepo:    invRepo,
		boardRepo:  boardRepo,
		userRepo:   userRepo,
		orgRepo:    orgRepo,
		jwtManager: jwtManager,
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

	resp := response.ToInvitationResponse(inv)

	existingUser, _ := s.userRepo.FindByEmail(ctx, inv.InviteeEmail)
	resp.IsNewUser = existingUser == nil

	return resp, nil
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

	board, err := s.boardRepo.FindByID(ctx, inv.BoardID)
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}
	if board != nil {
		existingOrgMember, _ := s.orgRepo.FindMember(ctx, board.OrganizationID, userID)
		if existingOrgMember == nil {
			orgMember := &domain.OrganizationMember{
				OrganizationID: board.OrganizationID,
				UserID:         userID,
				Role:           domain.OrgRoleMember,
			}
			_ = s.orgRepo.AddMember(ctx, orgMember)
		}
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

func (s *InvitationService) AcceptWithPassword(ctx context.Context, token string, req *request.AcceptInvitationWithPasswordRequest) (*response.AcceptInvitationResponse, error) {
	inv, err := s.invRepo.FindByToken(ctx, token)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}
	if inv == nil {
		return nil, apperror.ErrNotFound
	}

	if inv.IsExpired() {
		_ = s.invRepo.UpdateStatus(ctx, inv.ID, domain.InvitationExpired)
		return nil, apperror.New("INVITATION_EXPIRED", "This invitation has expired", 400)
	}

	if inv.Status != domain.InvitationPending {
		return nil, apperror.New("INVITATION_NOT_PENDING", "This invitation is no longer pending", 400)
	}

	existingUser, _ := s.userRepo.FindByEmail(ctx, inv.InviteeEmail)
	if existingUser != nil {
		return nil, apperror.New("USER_EXISTS", "An account with this email already exists. Please log in instead.", 400)
	}

	passwordHash, err := hash.HashPassword(req.Password)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}

	user := &domain.User{
		Email:        inv.InviteeEmail,
		PasswordHash: passwordHash,
		FullName:     req.FullName,
		IsVerified:   true,
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}

	member := &domain.BoardMember{
		BoardID: inv.BoardID,
		UserID:  user.ID,
		Role:    inv.Role,
	}
	if err := s.boardRepo.AddMember(ctx, member); err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}

	board, err := s.boardRepo.FindByID(ctx, inv.BoardID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}
	if board != nil {
		existingOrgMember, _ := s.orgRepo.FindMember(ctx, board.OrganizationID, user.ID)
		if existingOrgMember == nil {
			orgMember := &domain.OrganizationMember{
				OrganizationID: board.OrganizationID,
				UserID:         user.ID,
				Role:           domain.OrgRoleMember,
			}
			_ = s.orgRepo.AddMember(ctx, orgMember)
		}
	}

	if err := s.invRepo.UpdateStatus(ctx, inv.ID, domain.InvitationAccepted); err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}

	tokenPair, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}

	return &response.AcceptInvitationResponse{
		User:         response.ToUserResponse(user),
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    int64(time.Until(tokenPair.ExpiresAt).Seconds()),
		BoardID:      inv.BoardID,
	}, nil
}
