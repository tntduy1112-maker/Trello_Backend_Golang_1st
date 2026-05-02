package repository

import (
	"context"
	"encoding/json"

	"github.com/codewebkhongkho/trello-agent/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ActivityRepository struct {
	db *pgxpool.Pool
}

func NewActivityRepository(db *pgxpool.Pool) *ActivityRepository {
	return &ActivityRepository{db: db}
}

func (r *ActivityRepository) Create(ctx context.Context, activity *domain.ActivityLog) error {
	metadataJSON, err := json.Marshal(activity.Metadata)
	if err != nil {
		metadataJSON = []byte("{}")
	}

	query := `
		INSERT INTO activity_logs (id, board_id, card_id, list_id, user_id, action, metadata, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err = r.db.Exec(ctx, query,
		activity.ID,
		activity.BoardID,
		activity.CardID,
		activity.ListID,
		activity.UserID,
		activity.Action,
		metadataJSON,
		activity.Description,
		activity.CreatedAt,
	)
	return err
}

func (r *ActivityRepository) FindByCard(ctx context.Context, cardID string, limit, offset int) ([]*domain.ActivityLog, error) {
	query := `
		SELECT a.id, a.board_id, a.card_id, a.list_id, a.user_id, a.action, a.metadata, a.description, a.created_at,
		       u.id, u.full_name, u.avatar_url
		FROM activity_logs a
		JOIN users u ON a.user_id = u.id
		WHERE a.card_id = $1
		ORDER BY a.created_at DESC
		LIMIT $2 OFFSET $3
	`
	return r.queryActivities(ctx, query, cardID, limit, offset)
}

func (r *ActivityRepository) FindByBoard(ctx context.Context, boardID string, limit, offset int) ([]*domain.ActivityLog, error) {
	query := `
		SELECT a.id, a.board_id, a.card_id, a.list_id, a.user_id, a.action, a.metadata, a.description, a.created_at,
		       u.id, u.full_name, u.avatar_url
		FROM activity_logs a
		JOIN users u ON a.user_id = u.id
		WHERE a.board_id = $1
		ORDER BY a.created_at DESC
		LIMIT $2 OFFSET $3
	`
	return r.queryActivities(ctx, query, boardID, limit, offset)
}

func (r *ActivityRepository) queryActivities(ctx context.Context, query string, id string, limit, offset int) ([]*domain.ActivityLog, error) {
	rows, err := r.db.Query(ctx, query, id, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []*domain.ActivityLog
	for rows.Next() {
		var activity domain.ActivityLog
		var user domain.User
		var metadataJSON []byte

		err := rows.Scan(
			&activity.ID, &activity.BoardID, &activity.CardID, &activity.ListID,
			&activity.UserID, &activity.Action, &metadataJSON, &activity.Description, &activity.CreatedAt,
			&user.ID, &user.FullName, &user.AvatarURL,
		)
		if err != nil {
			return nil, err
		}

		activity.User = &user

		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &activity.Metadata)
		}

		activities = append(activities, &activity)
	}

	return activities, nil
}

func (r *ActivityRepository) CountByCard(ctx context.Context, cardID string) (int, error) {
	query := `SELECT COUNT(*) FROM activity_logs WHERE card_id = $1`
	var count int
	err := r.db.QueryRow(ctx, query, cardID).Scan(&count)
	return count, err
}

func (r *ActivityRepository) CountByBoard(ctx context.Context, boardID string) (int, error) {
	query := `SELECT COUNT(*) FROM activity_logs WHERE board_id = $1`
	var count int
	err := r.db.QueryRow(ctx, query, boardID).Scan(&count)
	return count, err
}
