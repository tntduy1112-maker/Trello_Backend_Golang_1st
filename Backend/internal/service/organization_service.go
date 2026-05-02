package service

import (
	"context"

	"github.com/codewebkhongkho/trello-agent/internal/domain"
	"github.com/codewebkhongkho/trello-agent/internal/dto/request"
	"github.com/codewebkhongkho/trello-agent/internal/dto/response"
	"github.com/codewebkhongkho/trello-agent/internal/repository"
	"github.com/codewebkhongkho/trello-agent/pkg/apperror"
)

type OrganizationService struct {
	orgRepo   repository.OrganizationRepository
	userRepo  repository.UserRepository
	boardRepo repository.BoardRepository
}

func NewOrganizationService(orgRepo repository.OrganizationRepository, userRepo repository.UserRepository, boardRepo repository.BoardRepository) *OrganizationService {
	return &OrganizationService{orgRepo: orgRepo, userRepo: userRepo, boardRepo: boardRepo}
}

func (s *OrganizationService) Create(ctx context.Context, userID string, req *request.CreateOrganizationRequest) (*response.OrganizationDetail, error) {
	slug, err := s.orgRepo.GenerateUniqueSlug(ctx, req.Name)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}

	org := &domain.Organization{
		Name:        req.Name,
		Slug:        slug,
		Description: req.Description,
		OwnerID:     userID,
	}

	if err := s.orgRepo.Create(ctx, org); err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}

	member := &domain.OrganizationMember{
		OrganizationID: org.ID,
		UserID:         userID,
		Role:           domain.OrgRoleOwner,
	}
	if err := s.orgRepo.AddMember(ctx, member); err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}

	user, _ := s.userRepo.FindByID(ctx, userID)

	return &response.OrganizationDetail{
		ID:           org.ID,
		Name:         org.Name,
		Slug:         org.Slug,
		Description:  org.Description,
		LogoURL:      org.LogoURL,
		Owner:        response.ToUserSummary(user),
		MembersCount: 1,
		BoardsCount:  0,
		MyRole:       domain.OrgRoleOwner,
		CreatedAt:    org.CreatedAt,
	}, nil
}

func (s *OrganizationService) List(ctx context.Context, userID string) ([]*response.OrganizationSummary, error) {
	orgs, err := s.orgRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}

	summaries := make([]*response.OrganizationSummary, 0, len(orgs))
	for _, org := range orgs {
		member, _ := s.orgRepo.FindMember(ctx, org.ID, userID)
		boardsCount, _ := s.orgRepo.CountBoards(ctx, org.ID)

		role := domain.OrgRoleMember
		if member != nil {
			role = member.Role
		}

		summaries = append(summaries, &response.OrganizationSummary{
			ID:          org.ID,
			Name:        org.Name,
			Slug:        org.Slug,
			LogoURL:     org.LogoURL,
			Role:        role,
			BoardsCount: boardsCount,
		})
	}

	return summaries, nil
}

func (s *OrganizationService) GetBySlug(ctx context.Context, userID, slug string) (*response.OrganizationDetail, error) {
	member, err := s.orgRepo.FindMemberBySlug(ctx, slug, userID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}
	if member == nil {
		return nil, apperror.ErrForbidden
	}

	org := member.Organization
	owner, _ := s.userRepo.FindByID(ctx, org.OwnerID)
	membersCount, _ := s.orgRepo.CountMembers(ctx, org.ID)
	boardsCount, _ := s.orgRepo.CountBoards(ctx, org.ID)
	isWorkspaceOwner, _ := s.boardRepo.IsWorkspaceOwner(ctx, org.ID, userID)

	return &response.OrganizationDetail{
		ID:               org.ID,
		Name:             org.Name,
		Slug:             org.Slug,
		Description:      org.Description,
		LogoURL:          org.LogoURL,
		Owner:            response.ToUserSummary(owner),
		MembersCount:     membersCount,
		BoardsCount:      boardsCount,
		MyRole:           member.Role,
		IsWorkspaceOwner: isWorkspaceOwner,
		CreatedAt:        org.CreatedAt,
	}, nil
}

func (s *OrganizationService) Update(ctx context.Context, userID, slug string, req *request.UpdateOrganizationRequest) (*response.OrganizationDetail, error) {
	member, err := s.orgRepo.FindMemberBySlug(ctx, slug, userID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}
	if member == nil || !member.Role.HasPermission(domain.OrgRoleAdmin) {
		return nil, apperror.ErrForbidden
	}

	org := member.Organization
	if req.Name != nil {
		org.Name = *req.Name
	}
	if req.Description != nil {
		org.Description = req.Description
	}

	if err := s.orgRepo.Update(ctx, org); err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}

	return s.GetBySlug(ctx, userID, slug)
}

func (s *OrganizationService) Delete(ctx context.Context, userID, slug string) error {
	member, err := s.orgRepo.FindMemberBySlug(ctx, slug, userID)
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}
	if member == nil || member.Role != domain.OrgRoleOwner {
		return apperror.ErrForbidden
	}

	if err := s.orgRepo.SoftDelete(ctx, member.Organization.ID); err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}

	return nil
}

func (s *OrganizationService) ListMembers(ctx context.Context, userID, slug string) ([]*response.OrgMemberResponse, error) {
	member, err := s.orgRepo.FindMemberBySlug(ctx, slug, userID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}
	if member == nil {
		return nil, apperror.ErrForbidden
	}

	members, err := s.orgRepo.FindMembers(ctx, member.Organization.ID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}

	responses := make([]*response.OrgMemberResponse, 0, len(members))
	for _, m := range members {
		responses = append(responses, response.ToOrgMemberResponse(m))
	}

	return responses, nil
}

func (s *OrganizationService) InviteMember(ctx context.Context, userID, slug string, req *request.InviteOrgMemberRequest) (*response.OrgMemberResponse, error) {
	member, err := s.orgRepo.FindMemberBySlug(ctx, slug, userID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}
	if member == nil || !member.Role.HasPermission(domain.OrgRoleAdmin) {
		return nil, apperror.ErrForbidden
	}

	invitee, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}
	if invitee == nil {
		return nil, apperror.New("USER_NOT_FOUND", "User with this email not found", 404)
	}

	org := member.Organization
	existingMember, err := s.orgRepo.FindMember(ctx, org.ID, invitee.ID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}
	if existingMember != nil {
		return nil, apperror.New("ALREADY_MEMBER", "User is already a member of this organization", 409)
	}

	newMember := &domain.OrganizationMember{
		OrganizationID: org.ID,
		UserID:         invitee.ID,
		Role:           req.Role,
	}
	if err := s.orgRepo.AddMember(ctx, newMember); err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}

	newMember.User = invitee
	return response.ToOrgMemberResponse(newMember), nil
}

func (s *OrganizationService) UpdateMemberRole(ctx context.Context, userID, slug, targetUserID string, role domain.OrgRole) error {
	member, err := s.orgRepo.FindMemberBySlug(ctx, slug, userID)
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}
	if member == nil || !member.Role.HasPermission(domain.OrgRoleAdmin) {
		return apperror.ErrForbidden
	}

	org := member.Organization
	targetMember, err := s.orgRepo.FindMember(ctx, org.ID, targetUserID)
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}
	if targetMember == nil {
		return apperror.ErrNotFound
	}

	if targetMember.Role == domain.OrgRoleOwner {
		return apperror.New("CANNOT_MODIFY_OWNER", "Cannot modify owner role", 403)
	}

	if member.Role == domain.OrgRoleAdmin && targetMember.Role == domain.OrgRoleAdmin {
		return apperror.New("CANNOT_MODIFY_ADMIN", "Admins cannot modify other admins", 403)
	}

	if role == domain.OrgRoleOwner {
		return apperror.New("INVALID_ROLE", "Cannot assign owner role", 400)
	}

	if err := s.orgRepo.UpdateMemberRole(ctx, org.ID, targetUserID, role); err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}

	return nil
}

func (s *OrganizationService) RemoveMember(ctx context.Context, userID, slug, targetUserID string) error {
	member, err := s.orgRepo.FindMemberBySlug(ctx, slug, userID)
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}
	if member == nil || !member.Role.HasPermission(domain.OrgRoleAdmin) {
		return apperror.ErrForbidden
	}

	org := member.Organization
	targetMember, err := s.orgRepo.FindMember(ctx, org.ID, targetUserID)
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}
	if targetMember == nil {
		return apperror.ErrNotFound
	}

	if targetMember.Role == domain.OrgRoleOwner {
		return apperror.New("CANNOT_REMOVE_OWNER", "Cannot remove organization owner", 403)
	}

	if member.Role == domain.OrgRoleAdmin && targetMember.Role == domain.OrgRoleAdmin {
		return apperror.New("CANNOT_REMOVE_ADMIN", "Admins cannot remove other admins", 403)
	}

	if err := s.orgRepo.RemoveMember(ctx, org.ID, targetUserID); err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}

	return nil
}

func (s *OrganizationService) Leave(ctx context.Context, userID, slug string) error {
	member, err := s.orgRepo.FindMemberBySlug(ctx, slug, userID)
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}
	if member == nil {
		return apperror.ErrForbidden
	}

	if member.Role == domain.OrgRoleOwner {
		return apperror.New("OWNER_CANNOT_LEAVE", "Owner must transfer ownership before leaving", 403)
	}

	if err := s.orgRepo.RemoveMember(ctx, member.Organization.ID, userID); err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}

	return nil
}

func (s *OrganizationService) GetMemberRole(ctx context.Context, slug, userID string) (*domain.OrganizationMember, error) {
	return s.orgRepo.FindMemberBySlug(ctx, slug, userID)
}

func (s *OrganizationService) ListBoardMembers(ctx context.Context, userID, slug string) ([]*response.BoardMemberWithBoards, error) {
	member, err := s.orgRepo.FindMemberBySlug(ctx, slug, userID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}
	if member == nil {
		return nil, apperror.ErrForbidden
	}

	boardMembers, err := s.boardRepo.FindMembersByOrgID(ctx, member.Organization.ID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal)
	}

	userBoardsMap := make(map[string]*response.BoardMemberWithBoards)
	for _, bm := range boardMembers {
		if _, exists := userBoardsMap[bm.UserID]; !exists {
			userBoardsMap[bm.UserID] = &response.BoardMemberWithBoards{
				User: &response.UserSummary{
					ID:        bm.User.ID,
					Email:     bm.User.Email,
					FullName:  bm.User.FullName,
					AvatarURL: bm.User.AvatarURL,
				},
				Boards: []response.BoardAccess{},
			}
		}
		userBoardsMap[bm.UserID].Boards = append(userBoardsMap[bm.UserID].Boards, response.BoardAccess{
			ID:           bm.Board.ID,
			Title:        bm.Board.Title,
			Role:         bm.Role,
			MembershipID: bm.ID,
		})
	}

	result := make([]*response.BoardMemberWithBoards, 0, len(userBoardsMap))
	for _, v := range userBoardsMap {
		result = append(result, v)
	}

	return result, nil
}

func (s *OrganizationService) UpdateBoardMemberRole(ctx context.Context, userID, slug, boardID, targetUserID string, role domain.BoardRole) error {
	member, err := s.orgRepo.FindMemberBySlug(ctx, slug, userID)
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}
	if member == nil {
		return apperror.ErrForbidden
	}

	isWorkspaceOwner, err := s.boardRepo.IsWorkspaceOwner(ctx, member.Organization.ID, userID)
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}
	if !isWorkspaceOwner {
		return apperror.New("NOT_WORKSPACE_OWNER", "Only workspace owners can manage board roles", 403)
	}

	board, err := s.boardRepo.FindByID(ctx, boardID)
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}
	if board == nil || board.OrganizationID != member.Organization.ID {
		return apperror.ErrNotFound
	}

	targetMember, err := s.boardRepo.FindMember(ctx, boardID, targetUserID)
	if err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}
	if targetMember == nil {
		return apperror.ErrNotFound
	}

	// Prevent owners from demoting themselves
	if userID == targetUserID && targetMember.Role == domain.BoardRoleOwner && role != domain.BoardRoleOwner {
		return apperror.New("CANNOT_DEMOTE_SELF", "You cannot demote yourself from owner role", 403)
	}

	if err := s.boardRepo.UpdateMemberRole(ctx, boardID, targetUserID, role); err != nil {
		return apperror.Wrap(err, apperror.ErrInternal)
	}

	return nil
}
