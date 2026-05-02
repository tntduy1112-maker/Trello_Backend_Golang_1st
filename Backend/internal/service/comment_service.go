package service

import (
	"context"
	"regexp"
	"time"

	"github.com/codewebkhongkho/trello-agent/internal/domain"
	"github.com/codewebkhongkho/trello-agent/internal/dto/request"
	"github.com/codewebkhongkho/trello-agent/internal/repository"
	"github.com/codewebkhongkho/trello-agent/pkg/apperror"
	"github.com/codewebkhongkho/trello-agent/pkg/cuid"
)

var mentionRegex = regexp.MustCompile(`@\[([^\]]+)\]\(([^)]+)\)`)

type CommentService struct {
	commentRepo         *repository.CommentRepository
	cardRepo            *repository.CardRepository
	boardRepo           repository.BoardRepository
	activityRepo        *repository.ActivityRepository
	userRepo            repository.UserRepository
	notificationService *NotificationService
}

func NewCommentService(
	commentRepo *repository.CommentRepository,
	cardRepo *repository.CardRepository,
	boardRepo repository.BoardRepository,
	activityRepo *repository.ActivityRepository,
	userRepo repository.UserRepository,
	notificationService *NotificationService,
) *CommentService {
	return &CommentService{
		commentRepo:         commentRepo,
		cardRepo:            cardRepo,
		boardRepo:           boardRepo,
		activityRepo:        activityRepo,
		userRepo:            userRepo,
		notificationService: notificationService,
	}
}

func (s *CommentService) ListByCard(ctx context.Context, userID, cardID string) ([]*domain.Comment, error) {
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

	allComments, err := s.commentRepo.FindByCardID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	return s.buildCommentTree(allComments), nil
}

func (s *CommentService) buildCommentTree(comments []*domain.Comment) []*domain.Comment {
	commentMap := make(map[string]*domain.Comment)
	var topLevel []*domain.Comment

	for _, c := range comments {
		commentMap[c.ID] = c
		c.Replies = []domain.Comment{}
	}

	for _, c := range comments {
		if c.ParentID != nil {
			if parent, ok := commentMap[*c.ParentID]; ok {
				parent.Replies = append(parent.Replies, *c)
			}
		} else {
			topLevel = append(topLevel, c)
		}
	}

	// Reverse top-level comments so newest is first
	for i, j := 0, len(topLevel)-1; i < j; i, j = i+1, j-1 {
		topLevel[i], topLevel[j] = topLevel[j], topLevel[i]
	}

	return topLevel
}

func (s *CommentService) Create(ctx context.Context, userID string, req request.CreateCommentRequest) (*domain.Comment, error) {
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

	now := time.Now()
	comment := &domain.Comment{
		ID:        cuid.New(),
		CardID:    req.CardID,
		AuthorID:  userID,
		Content:   req.Content,
		ParentID:  req.ParentID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return nil, err
	}

	// Log activity
	s.logActivity(ctx, boardID, &req.CardID, userID, domain.ActivityCommentAdded, map[string]interface{}{
		"comment_id": comment.ID,
	})

	// Send notifications
	if s.notificationService != nil {
		card, _ := s.cardRepo.FindByID(ctx, req.CardID)
		actor, _ := s.userRepo.FindByID(ctx, userID)

		if card != nil && actor != nil {
			// Notify card assignee (if not the commenter)
			if card.AssigneeID != nil && *card.AssigneeID != userID {
				s.notificationService.NotifyCommentAdded(ctx, req.CardID, card.Title, boardID, card.AssigneeID, actor)
			}

			// Process @mentions and notify mentioned users
			s.processMentions(ctx, req.Content, req.CardID, card.Title, boardID, comment.ID, actor)
		}
	}

	return s.commentRepo.FindByID(ctx, comment.ID)
}

func (s *CommentService) processMentions(ctx context.Context, content, cardID, cardTitle, boardID, commentID string, actor *domain.User) {
	matches := mentionRegex.FindAllStringSubmatch(content, -1)
	notifiedUsers := make(map[string]bool)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		mentionedUserID := match[2]

		// Skip if already notified or if it's the actor
		if notifiedUsers[mentionedUserID] || mentionedUserID == actor.ID {
			continue
		}

		// Verify user exists and has access to the board
		mentionedUser, err := s.userRepo.FindByID(ctx, mentionedUserID)
		if err != nil || mentionedUser == nil {
			continue
		}

		canAccess, _, err := s.boardRepo.CanUserAccess(ctx, boardID, mentionedUserID)
		if err != nil || !canAccess {
			continue
		}

		s.notificationService.NotifyMention(ctx, cardID, cardTitle, boardID, commentID, content, mentionedUser, actor)
		notifiedUsers[mentionedUserID] = true
	}
}

func (s *CommentService) Update(ctx context.Context, userID, commentID string, req request.UpdateCommentRequest) (*domain.Comment, error) {
	comment, err := s.commentRepo.FindByID(ctx, commentID)
	if err != nil {
		return nil, err
	}

	if comment.AuthorID != userID {
		return nil, apperror.ErrForbidden
	}

	comment.Content = req.Content
	if err := s.commentRepo.Update(ctx, comment); err != nil {
		return nil, err
	}

	return s.commentRepo.FindByID(ctx, commentID)
}

func (s *CommentService) Delete(ctx context.Context, userID, commentID string) error {
	comment, err := s.commentRepo.FindByID(ctx, commentID)
	if err != nil {
		return err
	}

	boardID, err := s.cardRepo.FindBoardIDByCardID(ctx, comment.CardID)
	if err != nil {
		return err
	}

	// Check if user is author or board admin
	isAuthor := comment.AuthorID == userID
	_, role, err := s.boardRepo.CanUserAccess(ctx, boardID, userID)
	if err != nil {
		return err
	}
	isAdmin := role == domain.BoardRoleAdmin || role == domain.BoardRoleOwner

	if !isAuthor && !isAdmin {
		return apperror.ErrForbidden
	}

	return s.commentRepo.SoftDelete(ctx, commentID)
}

func (s *CommentService) logActivity(ctx context.Context, boardID string, cardID *string, userID string, action domain.ActivityAction, metadata map[string]interface{}) {
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
