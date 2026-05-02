package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/codewebkhongkho/trello-agent/internal/domain"
	"github.com/codewebkhongkho/trello-agent/pkg/apperror"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationRepository struct {
	db *pgxpool.Pool
}

func NewNotificationRepository(db *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, notification *domain.Notification) error {
	metadataJSON, err := json.Marshal(notification.Metadata)
	if err != nil {
		metadataJSON = []byte("{}")
	}

	query := `
		INSERT INTO notifications (id, user_id, type, title, message, board_id, card_id, actor_id, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err = r.db.Exec(ctx, query,
		notification.ID,
		notification.UserID,
		notification.Type,
		notification.Title,
		notification.Message,
		notification.BoardID,
		notification.CardID,
		notification.ActorID,
		metadataJSON,
		notification.CreatedAt,
	)
	return err
}

func (r *NotificationRepository) FindByID(ctx context.Context, id string) (*domain.Notification, error) {
	query := `
		SELECT n.id, n.user_id, n.type, n.title, n.message, n.board_id, n.card_id, n.actor_id,
		       n.is_read, n.read_at, n.metadata, n.created_at,
		       u.id, u.full_name, u.avatar_url
		FROM notifications n
		LEFT JOIN users u ON n.actor_id = u.id
		WHERE n.id = $1
	`
	row := r.db.QueryRow(ctx, query, id)

	var n domain.Notification
	var actor domain.User
	var actorID *string
	var metadataJSON []byte

	err := row.Scan(
		&n.ID, &n.UserID, &n.Type, &n.Title, &n.Message, &n.BoardID, &n.CardID, &n.ActorID,
		&n.IsRead, &n.ReadAt, &metadataJSON, &n.CreatedAt,
		&actorID, &actor.FullName, &actor.AvatarURL,
	)
	if err == pgx.ErrNoRows {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if actorID != nil {
		actor.ID = *actorID
		n.Actor = &actor
	}

	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &n.Metadata)
	}

	return &n, nil
}

func (r *NotificationRepository) FindByUser(ctx context.Context, userID string, limit, offset int, unreadOnly bool) ([]*domain.Notification, error) {
	query := `
		SELECT n.id, n.user_id, n.type, n.title, n.message, n.board_id, n.card_id, n.actor_id,
		       n.is_read, n.read_at, n.metadata, n.created_at,
		       u.id, u.full_name, u.avatar_url
		FROM notifications n
		LEFT JOIN users u ON n.actor_id = u.id
		WHERE n.user_id = $1
	`

	args := []interface{}{userID}

	if unreadOnly {
		query += " AND n.is_read = FALSE"
	}

	query += " ORDER BY n.created_at DESC"
	query += " LIMIT $2 OFFSET $3"
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*domain.Notification
	for rows.Next() {
		var n domain.Notification
		var actor domain.User
		var actorID *string
		var metadataJSON []byte

		err := rows.Scan(
			&n.ID, &n.UserID, &n.Type, &n.Title, &n.Message, &n.BoardID, &n.CardID, &n.ActorID,
			&n.IsRead, &n.ReadAt, &metadataJSON, &n.CreatedAt,
			&actorID, &actor.FullName, &actor.AvatarURL,
		)
		if err != nil {
			return nil, err
		}

		if actorID != nil {
			actor.ID = *actorID
			n.Actor = &actor
		}

		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &n.Metadata)
		}

		notifications = append(notifications, &n)
	}

	return notifications, nil
}

func (r *NotificationRepository) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = FALSE`
	var count int
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}

func (r *NotificationRepository) CountByUser(ctx context.Context, userID string, unreadOnly bool) (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1`
	if unreadOnly {
		query += " AND is_read = FALSE"
	}
	var count int
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}

func (r *NotificationRepository) MarkAsRead(ctx context.Context, id string) error {
	query := `UPDATE notifications SET is_read = TRUE, read_at = $1 WHERE id = $2`
	result, err := r.db.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}
	return nil
}

func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, userID string) error {
	query := `UPDATE notifications SET is_read = TRUE, read_at = $1 WHERE user_id = $2 AND is_read = FALSE`
	_, err := r.db.Exec(ctx, query, time.Now(), userID)
	return err
}

func (r *NotificationRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM notifications WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}
	return nil
}

func (r *NotificationRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	query := `DELETE FROM notifications WHERE created_at < $1`
	result, err := r.db.Exec(ctx, query, before)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
