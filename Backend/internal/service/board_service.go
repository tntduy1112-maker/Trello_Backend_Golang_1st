package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/codewebkhongkho/trello-agent/internal/domain"
	"github.com/codewebkhongkho/trello-agent/internal/dto/request"
	"github.com/codewebkhongkho/trello-agent/internal/dto/response"
	"github.com/codewebkhongkho/trello-agent/internal/repository"
	"github.com/codewebkhongkho/trello-agent/pkg/apperror"
	"github.com/codewebkhongkho/trello-agent/pkg/email"
)

type BoardService struct {
	boardRepo   repository.BoardRepository
	orgRepo     repository.OrganizationRepository
	invRepo     repository.InvitationRepository
	userRepo    repository.UserRepository
	listRepo    *repository.ListRepository
	labelRepo   *repository.LabelRepository
	emailSvc    *email.Service
	frontendURL string
}

type BoardServiceConfig struct {
	BoardRepo   repository.BoardRepository
	OrgRepo     repository.OrganizationRepository
	InvRepo     repository.InvitationRepository
	UserRepo    repository.UserRepository
	ListRepo    *repository.ListRepository
	LabelRepo   *repository.LabelRepository
	EmailSvc    *email.Service
	FrontendURL string
}

func NewBoardService(cfg BoardServiceConfig) *BoardService {
	return &BoardService{
		boardRepo:   cfg.BoardRepo,
		orgRepo:     cfg.OrgRepo,
		invRepo:     cfg.InvRepo,
		userRepo:    cfg.UserRepo,
		listRepo:    cfg.ListRepo,
		labelRepo:   cfg.LabelRepo,
		emailSvc:    cfg.EmailSvc,
		frontendURL: cfg.FrontendURL,
	}
}

func (s *BoardService) Create(ctx context.Context, userID, orgSlug string, req *request.CreateBoardRequest) (*response.BoardDetail, error) {
	member, err := s.orgRepo.FindMemberBySlug(ctx, orgSlug, userID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}
	if member == nil {
		return nil, apperror.ErrForbidden
	}

	board := &domain.Board{
		OrganizationID:  member.Organization.ID,
		Title:           req.Title,
		Description:     req.Description,
		BackgroundColor: req.BackgroundColor,
		Visibility:      req.Visibility,
		OwnerID:         userID,
	}
	if board.BackgroundColor == "" {
		board.BackgroundColor = domain.DefaultBackgroundColor
	}
	if board.Visibility == "" {
		board.Visibility = domain.VisibilityWorkspace
	}

	if err := s.boardRepo.Create(ctx, board); err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}

	boardMember := &domain.BoardMember{
		BoardID: board.ID,
		UserID:  userID,
		Role:    domain.BoardRoleOwner,
	}
	if err := s.boardRepo.AddMember(ctx, boardMember); err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}

	return s.GetByID(ctx, userID, board.ID)
}

func (s *BoardService) ListByOrg(ctx context.Context, userID, orgSlug string, includeClosed bool) ([]*response.BoardSummary, error) {
	member, err := s.orgRepo.FindMemberBySlug(ctx, orgSlug, userID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}
	if member == nil {
		return nil, apperror.ErrForbidden
	}

	boards, err := s.boardRepo.FindByOrgID(ctx, member.Organization.ID, includeClosed)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}

	summaries := make([]*response.BoardSummary, 0, len(boards))
	for _, board := range boards {
		canAccess, role, _ := s.boardRepo.CanUserAccess(ctx, board.ID, userID)
		if !canAccess {
			continue
		}

		listsCount, _ := s.boardRepo.CountLists(ctx, board.ID)
		cardsCount, _ := s.boardRepo.CountCards(ctx, board.ID)

		summaries = append(summaries, &response.BoardSummary{
			ID:              board.ID,
			Title:           board.Title,
			BackgroundColor: board.BackgroundColor,
			Visibility:      board.Visibility,
			IsClosed:        board.IsClosed,
			ListsCount:      listsCount,
			CardsCount:      cardsCount,
			MyRole:          role,
		})
	}

	return summaries, nil
}

func (s *BoardService) GetByID(ctx context.Context, userID, boardID string) (*response.BoardDetail, error) {
	canAccess, role, err := s.boardRepo.CanUserAccess(ctx, boardID, userID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}
	if !canAccess {
		return nil, apperror.ErrForbidden
	}

	board, err := s.boardRepo.FindByID(ctx, boardID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}
	if board == nil {
		return nil, apperror.ErrNotFound
	}

	members, _ := s.boardRepo.FindMembers(ctx, boardID)
	memberResponses := make([]*response.BoardMemberResponse, 0, len(members))
	for _, m := range members {
		memberResponses = append(memberResponses, response.ToBoardMemberResponse(m))
	}

	var orgSummary *response.OrgSummary
	if board.Organization != nil {
		orgSummary = &response.OrgSummary{
			ID:   board.Organization.ID,
			Name: board.Organization.Name,
			Slug: board.Organization.Slug,
		}
	}

	listsWithCards := make([]*response.ListWithCards, 0)
	if s.listRepo != nil {
		lists, err := s.listRepo.FindByBoardIDWithCards(ctx, boardID)
		if err == nil {
			for _, list := range lists {
				cards := make([]response.CardSummary, 0, len(list.Cards))
				for _, card := range list.Cards {
					var assignee *response.UserSummary
					if card.Assignee != nil {
						assignee = &response.UserSummary{
							ID:        card.Assignee.ID,
							FullName:  card.Assignee.FullName,
							AvatarURL: card.Assignee.AvatarURL,
						}
					}
					cards = append(cards, response.CardSummary{
						ID:          card.ID,
						Title:       card.Title,
						Position:    card.Position,
						Priority:    card.Priority,
						DueDate:     card.DueDate,
						IsCompleted: card.IsCompleted,
						Assignee:    assignee,
						Labels:      []response.LabelSummary{},
					})
				}
				listsWithCards = append(listsWithCards, &response.ListWithCards{
					ID:         list.ID,
					Title:      list.Title,
					Position:   list.Position,
					CardsCount: len(list.Cards),
					Cards:      cards,
				})
			}
		}
	}

	labels := make([]*response.LabelSummary, 0)
	if s.labelRepo != nil {
		boardLabels, err := s.labelRepo.FindByBoardID(ctx, boardID)
		if err == nil {
			for _, label := range boardLabels {
				labels = append(labels, &response.LabelSummary{
					ID:    label.ID,
					Name:  label.Name,
					Color: label.Color,
				})
			}
		}
	}

	return &response.BoardDetail{
		ID:              board.ID,
		Title:           board.Title,
		Description:     board.Description,
		BackgroundColor: board.BackgroundColor,
		Visibility:      board.Visibility,
		IsClosed:        board.IsClosed,
		MyRole:          role,
		Organization:    orgSummary,
		Members:         memberResponses,
		Lists:           listsWithCards,
		Labels:          labels,
		CreatedAt:       board.CreatedAt,
	}, nil
}

func (s *BoardService) Update(ctx context.Context, userID, boardID string, req *request.UpdateBoardRequest) (*response.BoardDetail, error) {
	canAccess, role, err := s.boardRepo.CanUserAccess(ctx, boardID, userID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}
	if !canAccess || !role.CanManage() {
		return nil, apperror.ErrForbidden
	}

	board, err := s.boardRepo.FindByID(ctx, boardID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}
	if board == nil {
		return nil, apperror.ErrNotFound
	}

	if req.Title != nil {
		board.Title = *req.Title
	}
	if req.Description != nil {
		board.Description = req.Description
	}
	if req.BackgroundColor != nil {
		board.BackgroundColor = *req.BackgroundColor
	}
	if req.Visibility != nil {
		board.Visibility = *req.Visibility
	}

	if err := s.boardRepo.Update(ctx, board); err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}

	return s.GetByID(ctx, userID, boardID)
}

func (s *BoardService) Close(ctx context.Context, userID, boardID string) error {
	canAccess, role, err := s.boardRepo.CanUserAccess(ctx, boardID, userID)
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}
	if !canAccess || !role.CanManage() {
		return apperror.ErrForbidden
	}

	return s.boardRepo.Close(ctx, boardID)
}

func (s *BoardService) Reopen(ctx context.Context, userID, boardID string) error {
	canAccess, role, err := s.boardRepo.CanUserAccess(ctx, boardID, userID)
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}
	if !canAccess || !role.CanManage() {
		return apperror.ErrForbidden
	}

	return s.boardRepo.Reopen(ctx, boardID)
}

func (s *BoardService) Delete(ctx context.Context, userID, boardID string) error {
	canAccess, role, err := s.boardRepo.CanUserAccess(ctx, boardID, userID)
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}
	if !canAccess || role != domain.BoardRoleOwner {
		return apperror.ErrForbidden
	}

	return s.boardRepo.SoftDelete(ctx, boardID)
}

func (s *BoardService) ListMembers(ctx context.Context, userID, boardID string) ([]*response.BoardMemberResponse, error) {
	canAccess, _, err := s.boardRepo.CanUserAccess(ctx, boardID, userID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}
	if !canAccess {
		return nil, apperror.ErrForbidden
	}

	members, err := s.boardRepo.FindMembers(ctx, boardID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}

	responses := make([]*response.BoardMemberResponse, 0, len(members))
	for _, m := range members {
		responses = append(responses, response.ToBoardMemberResponse(m))
	}

	return responses, nil
}

func (s *BoardService) Invite(ctx context.Context, userID, boardID string, req *request.InviteBoardMemberRequest) (*response.InvitationResponse, error) {
	canAccess, role, err := s.boardRepo.CanUserAccess(ctx, boardID, userID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}
	if !canAccess || !role.CanManage() {
		return nil, apperror.ErrForbidden
	}

	// Check if invitee is already a member
	invitee, _ := s.userRepo.FindByEmail(ctx, req.Email)
	if invitee != nil {
		existingMember, _ := s.boardRepo.FindMember(ctx, boardID, invitee.ID)
		if existingMember != nil {
			return nil, apperror.New("ALREADY_MEMBER", "User is already a board member", 409)
		}
	}

	existingInv, _ := s.invRepo.FindByBoardAndEmail(ctx, boardID, req.Email)
	if existingInv != nil && existingInv.IsPending() {
		return nil, apperror.New("INVITATION_EXISTS", "Pending invitation already exists", 409)
	}

	token := generateInviteToken()
	var inviteeID *string
	if invitee != nil {
		inviteeID = &invitee.ID
	}

	inv := &domain.BoardInvitation{
		BoardID:      boardID,
		InviterID:    userID,
		InviteeID:    inviteeID,
		InviteeEmail: req.Email,
		Role:         req.Role,
		Token:        token,
		Message:      req.Message,
		ExpiresAt:    time.Now().Add(domain.InvitationExpiresIn),
	}

	if err := s.invRepo.Create(ctx, inv); err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}

	inv, _ = s.invRepo.FindByToken(ctx, token)

	return response.ToInvitationResponseWithURL(inv, s.frontendURL), nil
}

func (s *BoardService) ListInvitations(ctx context.Context, userID, boardID string) ([]*response.InvitationResponse, error) {
	canAccess, role, err := s.boardRepo.CanUserAccess(ctx, boardID, userID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}
	if !canAccess || !role.CanManage() {
		return nil, apperror.ErrForbidden
	}

	invitations, err := s.invRepo.FindPendingByBoardID(ctx, boardID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}

	responses := make([]*response.InvitationResponse, len(invitations))
	for i, inv := range invitations {
		responses[i] = response.ToInvitationResponseWithURL(inv, s.frontendURL)
	}
	return responses, nil
}

func (s *BoardService) RevokeInvitation(ctx context.Context, userID, invitationID string) error {
	// Get invitation to check board access
	invitations, err := s.invRepo.FindPendingByBoardID(ctx, "")
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}

	var targetInv *domain.BoardInvitation
	for _, inv := range invitations {
		if inv.ID == invitationID {
			targetInv = inv
			break
		}
	}

	if targetInv == nil {
		return apperror.ErrNotFound
	}

	canAccess, role, err := s.boardRepo.CanUserAccess(ctx, targetInv.BoardID, userID)
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}
	if !canAccess || !role.CanManage() {
		return apperror.ErrForbidden
	}

	return s.invRepo.Delete(ctx, invitationID)
}

func (s *BoardService) UpdateMemberRole(ctx context.Context, userID, boardID, targetUserID string, role domain.BoardRole) error {
	canAccess, myRole, err := s.boardRepo.CanUserAccess(ctx, boardID, userID)
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}
	if !canAccess || !myRole.CanManage() {
		return apperror.ErrForbidden
	}

	targetMember, err := s.boardRepo.FindMember(ctx, boardID, targetUserID)
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}
	if targetMember == nil {
		return apperror.ErrNotFound
	}

	if targetMember.Role == domain.BoardRoleOwner {
		return apperror.New("CANNOT_MODIFY_OWNER", "Cannot modify owner role", 403)
	}

	if myRole == domain.BoardRoleAdmin && targetMember.Role == domain.BoardRoleAdmin {
		return apperror.New("CANNOT_MODIFY_ADMIN", "Admins cannot modify other admins", 403)
	}

	if role == domain.BoardRoleOwner {
		return apperror.New("INVALID_ROLE", "Cannot assign owner role", 400)
	}

	return s.boardRepo.UpdateMemberRole(ctx, boardID, targetUserID, role)
}

func (s *BoardService) RemoveMember(ctx context.Context, userID, boardID, targetUserID string) error {
	canAccess, myRole, err := s.boardRepo.CanUserAccess(ctx, boardID, userID)
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}
	if !canAccess || !myRole.CanManage() {
		return apperror.ErrForbidden
	}

	targetMember, err := s.boardRepo.FindMember(ctx, boardID, targetUserID)
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}
	if targetMember == nil {
		return apperror.ErrNotFound
	}

	if targetMember.Role == domain.BoardRoleOwner {
		return apperror.New("CANNOT_REMOVE_OWNER", "Cannot remove board owner", 403)
	}

	if myRole == domain.BoardRoleAdmin && targetMember.Role == domain.BoardRoleAdmin {
		return apperror.New("CANNOT_REMOVE_ADMIN", "Admins cannot remove other admins", 403)
	}

	return s.boardRepo.RemoveMember(ctx, boardID, targetUserID)
}

func (s *BoardService) Leave(ctx context.Context, userID, boardID string) error {
	member, err := s.boardRepo.FindMember(ctx, boardID, userID)
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}
	if member == nil {
		return apperror.ErrForbidden
	}

	if member.Role == domain.BoardRoleOwner {
		return apperror.New("OWNER_CANNOT_LEAVE", "Owner must transfer ownership before leaving", 403)
	}

	return s.boardRepo.RemoveMember(ctx, boardID, userID)
}

func generateInviteToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
