package service

import (
	"context"
	"fmt"
	"time"

	"github.com/codewebkhongkho/trello-agent/internal/domain"
	"github.com/codewebkhongkho/trello-agent/internal/repository"
	"github.com/codewebkhongkho/trello-agent/pkg/cuid"
)

type NotificationService struct {
	notificationRepo *repository.NotificationRepository
	sseManager       *SSEManager
}

func NewNotificationService(
	notificationRepo *repository.NotificationRepository,
	sseManager *SSEManager,
) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		sseManager:       sseManager,
	}
}

type ListNotificationsResponse struct {
	Notifications []*domain.Notification `json:"notifications"`
	Total         int                    `json:"total"`
	UnreadCount   int                    `json:"unread_count"`
	Page          int                    `json:"page"`
	Limit         int                    `json:"limit"`
}

func (s *NotificationService) List(ctx context.Context, userID string, page, limit int, unreadOnly bool) (*ListNotificationsResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}

	offset := (page - 1) * limit

	notifications, err := s.notificationRepo.FindByUser(ctx, userID, limit, offset, unreadOnly)
	if err != nil {
		return nil, err
	}

	total, err := s.notificationRepo.CountByUser(ctx, userID, unreadOnly)
	if err != nil {
		return nil, err
	}

	unreadCount, err := s.notificationRepo.GetUnreadCount(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &ListNotificationsResponse{
		Notifications: notifications,
		Total:         total,
		UnreadCount:   unreadCount,
		Page:          page,
		Limit:         limit,
	}, nil
}

func (s *NotificationService) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	return s.notificationRepo.GetUnreadCount(ctx, userID)
}

func (s *NotificationService) MarkAsRead(ctx context.Context, userID, notificationID string) error {
	notification, err := s.notificationRepo.FindByID(ctx, notificationID)
	if err != nil {
		return err
	}

	if notification.UserID != userID {
		return fmt.Errorf("notification does not belong to user")
	}

	return s.notificationRepo.MarkAsRead(ctx, notificationID)
}

func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID string) error {
	return s.notificationRepo.MarkAllAsRead(ctx, userID)
}

func (s *NotificationService) Delete(ctx context.Context, userID, notificationID string) error {
	notification, err := s.notificationRepo.FindByID(ctx, notificationID)
	if err != nil {
		return err
	}

	if notification.UserID != userID {
		return fmt.Errorf("notification does not belong to user")
	}

	return s.notificationRepo.Delete(ctx, notificationID)
}

func (s *NotificationService) NotifyCardAssigned(ctx context.Context, cardID, cardTitle, boardID string, assignee, actor *domain.User) error {
	if assignee.ID == actor.ID {
		return nil
	}

	notification := &domain.Notification{
		ID:        cuid.New(),
		UserID:    assignee.ID,
		Type:      domain.NotificationCardAssigned,
		Title:     "You were assigned to a card",
		Message:   fmt.Sprintf("%s assigned you to \"%s\"", actor.FullName, cardTitle),
		BoardID:   &boardID,
		CardID:    &cardID,
		ActorID:   &actor.ID,
		Metadata: map[string]interface{}{
			"card_title": cardTitle,
			"actor_name": actor.FullName,
		},
		CreatedAt: time.Now(),
		Actor:     actor,
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	s.broadcastToUser(assignee.ID, notification)
	return nil
}

func (s *NotificationService) NotifyCommentAdded(ctx context.Context, cardID, cardTitle, boardID string, cardAssigneeID *string, actor *domain.User) error {
	if cardAssigneeID == nil || *cardAssigneeID == actor.ID {
		return nil
	}

	notification := &domain.Notification{
		ID:        cuid.New(),
		UserID:    *cardAssigneeID,
		Type:      domain.NotificationCommentAdded,
		Title:     "New comment on your card",
		Message:   fmt.Sprintf("%s commented on \"%s\"", actor.FullName, cardTitle),
		BoardID:   &boardID,
		CardID:    &cardID,
		ActorID:   &actor.ID,
		Metadata: map[string]interface{}{
			"card_title": cardTitle,
			"actor_name": actor.FullName,
		},
		CreatedAt: time.Now(),
		Actor:     actor,
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	s.broadcastToUser(*cardAssigneeID, notification)
	return nil
}

func (s *NotificationService) NotifyBoardInvitation(ctx context.Context, inviteeEmail, boardTitle, boardID string, inviter *domain.User) error {
	notification := &domain.Notification{
		ID:      cuid.New(),
		UserID:  inviteeEmail,
		Type:    domain.NotificationBoardInvitation,
		Title:   "Board invitation",
		Message: fmt.Sprintf("%s invited you to join \"%s\"", inviter.FullName, boardTitle),
		BoardID: &boardID,
		ActorID: &inviter.ID,
		Metadata: map[string]interface{}{
			"board_title":  boardTitle,
			"inviter_name": inviter.FullName,
		},
		CreatedAt: time.Now(),
		Actor:     inviter,
	}

	return s.notificationRepo.Create(ctx, notification)
}

func (s *NotificationService) NotifyCardDueSoon(ctx context.Context, cardID, cardTitle, boardID string, dueDate *time.Time, assignee *domain.User) error {
	if dueDate == nil {
		return nil
	}

	timeUntil := time.Until(*dueDate)
	dueIn := "24 hours"
	if timeUntil < time.Hour {
		dueIn = "less than an hour"
	} else if timeUntil < 2*time.Hour {
		dueIn = "1 hour"
	} else {
		dueIn = fmt.Sprintf("%d hours", int(timeUntil.Hours()))
	}

	notification := &domain.Notification{
		ID:      cuid.New(),
		UserID:  assignee.ID,
		Type:    domain.NotificationCardDueSoon,
		Title:   "Card due soon",
		Message: fmt.Sprintf("\"%s\" is due in %s", cardTitle, dueIn),
		BoardID: &boardID,
		CardID:  &cardID,
		Metadata: map[string]interface{}{
			"card_title": cardTitle,
			"due_date":   dueDate,
		},
		CreatedAt: time.Now(),
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	s.broadcastToUser(assignee.ID, notification)
	return nil
}

func (s *NotificationService) NotifyCardOverdue(ctx context.Context, cardID, cardTitle, boardID string, dueDate *time.Time, assignee *domain.User) error {
	notification := &domain.Notification{
		ID:      cuid.New(),
		UserID:  assignee.ID,
		Type:    domain.NotificationCardOverdue,
		Title:   "Card overdue",
		Message: fmt.Sprintf("\"%s\" is overdue", cardTitle),
		BoardID: &boardID,
		CardID:  &cardID,
		Metadata: map[string]interface{}{
			"card_title": cardTitle,
			"due_date":   dueDate,
		},
		CreatedAt: time.Now(),
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	s.broadcastToUser(assignee.ID, notification)
	return nil
}

func (s *NotificationService) NotifyMention(ctx context.Context, cardID, cardTitle, boardID, commentID, commentContent string, mentionedUser, actor *domain.User) error {
	if mentionedUser.ID == actor.ID {
		return nil
	}

	notification := &domain.Notification{
		ID:      cuid.New(),
		UserID:  mentionedUser.ID,
		Type:    domain.NotificationMentioned,
		Title:   "You were mentioned in a comment",
		Message: fmt.Sprintf("%s mentioned you in \"%s\"", actor.FullName, cardTitle),
		BoardID: &boardID,
		CardID:  &cardID,
		ActorID: &actor.ID,
		Metadata: map[string]interface{}{
			"card_title":      cardTitle,
			"actor_name":      actor.FullName,
			"comment_id":      commentID,
			"comment_preview": truncateString(commentContent, 100),
		},
		CreatedAt: time.Now(),
		Actor:     actor,
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	s.broadcastToUser(mentionedUser.ID, notification)
	return nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func (s *NotificationService) broadcastToUser(userID string, notification *domain.Notification) {
	if s.sseManager == nil {
		return
	}
	s.sseManager.SendToUser(userID, SSEEvent{
		Type: "notification",
		Data: notification,
	})
}
