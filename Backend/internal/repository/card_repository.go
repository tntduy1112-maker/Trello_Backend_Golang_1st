package repository

import (
	"context"
	"time"

	"github.com/codewebkhongkho/trello-agent/internal/domain"
	"github.com/codewebkhongkho/trello-agent/pkg/apperror"
	"github.com/codewebkhongkho/trello-agent/pkg/position"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CardRepository struct {
	db *pgxpool.Pool
}

func NewCardRepository(db *pgxpool.Pool) *CardRepository {
	return &CardRepository{db: db}
}

func (r *CardRepository) Create(ctx context.Context, card *domain.Card) error {
	query := `
		INSERT INTO cards (id, list_id, title, description, position, priority, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query,
		card.ID, card.ListID, card.Title, card.Description, card.Position,
		card.Priority, card.CreatedBy, card.CreatedAt, card.UpdatedAt,
	)
	return err
}

func (r *CardRepository) FindByID(ctx context.Context, id string) (*domain.Card, error) {
	query := `
		SELECT id, list_id, title, description, position, assignee_id, priority,
			   due_date, is_completed, completed_at, is_archived, created_at, updated_at, created_by
		FROM cards WHERE id = $1
	`
	row := r.db.QueryRow(ctx, query, id)

	var card domain.Card
	err := row.Scan(
		&card.ID, &card.ListID, &card.Title, &card.Description, &card.Position,
		&card.AssigneeID, &card.Priority, &card.DueDate, &card.IsCompleted,
		&card.CompletedAt, &card.IsArchived, &card.CreatedAt, &card.UpdatedAt, &card.CreatedBy,
	)
	if err == pgx.ErrNoRows {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &card, nil
}

func (r *CardRepository) FindByIDWithDetails(ctx context.Context, id string) (*domain.Card, error) {
	card, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if card.AssigneeID != nil {
		userQuery := `SELECT id, full_name, avatar_url FROM users WHERE id = $1`
		var user domain.User
		err := r.db.QueryRow(ctx, userQuery, *card.AssigneeID).Scan(&user.ID, &user.FullName, &user.AvatarURL)
		if err == nil {
			card.Assignee = &user
		}
	}

	labelQuery := `
		SELECT l.id, l.name, l.color
		FROM labels l
		JOIN card_labels cl ON l.id = cl.label_id
		WHERE cl.card_id = $1
	`
	rows, err := r.db.Query(ctx, labelQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var label domain.Label
		if err := rows.Scan(&label.ID, &label.Name, &label.Color); err != nil {
			return nil, err
		}
		card.Labels = append(card.Labels, &label)
	}

	return card, nil
}

func (r *CardRepository) FindByListID(ctx context.Context, listID string, includeArchived bool) ([]*domain.Card, error) {
	query := `
		SELECT c.id, c.list_id, c.title, c.description, c.position, c.assignee_id,
			   c.priority, c.due_date, c.is_completed, c.completed_at,
			   c.is_archived, c.created_at, c.updated_at
		FROM cards c
		WHERE c.list_id = $1
	`
	if !includeArchived {
		query += " AND c.is_archived = false"
	}
	query += " ORDER BY c.position ASC"

	rows, err := r.db.Query(ctx, query, listID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*domain.Card
	for rows.Next() {
		var card domain.Card
		err := rows.Scan(
			&card.ID, &card.ListID, &card.Title, &card.Description, &card.Position,
			&card.AssigneeID, &card.Priority, &card.DueDate, &card.IsCompleted,
			&card.CompletedAt, &card.IsArchived, &card.CreatedAt, &card.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		cards = append(cards, &card)
	}
	return cards, nil
}

func (r *CardRepository) Update(ctx context.Context, card *domain.Card) error {
	query := `
		UPDATE cards SET title = $1, description = $2, priority = $3, due_date = $4, updated_at = $5
		WHERE id = $6
	`
	_, err := r.db.Exec(ctx, query,
		card.Title, card.Description, card.Priority, card.DueDate, time.Now(), card.ID,
	)
	return err
}

func (r *CardRepository) Archive(ctx context.Context, id string) error {
	now := time.Now()
	query := `UPDATE cards SET is_archived = true, archived_at = $1, updated_at = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, now, id)
	return err
}

func (r *CardRepository) Restore(ctx context.Context, id string) error {
	query := `UPDATE cards SET is_archived = false, archived_at = NULL, updated_at = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, time.Now(), id)
	return err
}

func (r *CardRepository) GetMaxPosition(ctx context.Context, listID string) (float64, error) {
	query := `SELECT COALESCE(MAX(position), 0) FROM cards WHERE list_id = $1`
	var maxPos float64
	err := r.db.QueryRow(ctx, query, listID).Scan(&maxPos)
	return maxPos, err
}

func (r *CardRepository) UpdatePosition(ctx context.Context, id, listID string, pos float64) error {
	query := `UPDATE cards SET list_id = $1, position = $2, updated_at = $3 WHERE id = $4`
	_, err := r.db.Exec(ctx, query, listID, pos, time.Now(), id)
	return err
}

func (r *CardRepository) RebalancePositions(ctx context.Context, listID string) error {
	query := `
		SELECT id FROM cards
		WHERE list_id = $1 AND is_archived = false
		ORDER BY position ASC
	`
	rows, err := r.db.Query(ctx, query, listID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return err
		}
		ids = append(ids, id)
	}

	positions := position.Rebalance(len(ids))
	for i, id := range ids {
		updateQuery := `UPDATE cards SET position = $1, updated_at = $2 WHERE id = $3`
		if _, err := r.db.Exec(ctx, updateQuery, positions[i], time.Now(), id); err != nil {
			return err
		}
	}
	return nil
}

func (r *CardRepository) Assign(ctx context.Context, cardID, userID string) error {
	query := `UPDATE cards SET assignee_id = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.Exec(ctx, query, userID, time.Now(), cardID)
	return err
}

func (r *CardRepository) Unassign(ctx context.Context, cardID string) error {
	query := `UPDATE cards SET assignee_id = NULL, updated_at = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, time.Now(), cardID)
	return err
}

func (r *CardRepository) MarkComplete(ctx context.Context, cardID string) error {
	now := time.Now()
	query := `UPDATE cards SET is_completed = true, completed_at = $1, updated_at = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, now, cardID)
	return err
}

func (r *CardRepository) MarkIncomplete(ctx context.Context, cardID string) error {
	query := `UPDATE cards SET is_completed = false, completed_at = NULL, updated_at = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, time.Now(), cardID)
	return err
}

func (r *CardRepository) FindBoardIDByCardID(ctx context.Context, cardID string) (string, error) {
	query := `
		SELECT l.board_id FROM lists l
		JOIN cards c ON c.list_id = l.id
		WHERE c.id = $1
	`
	var boardID string
	err := r.db.QueryRow(ctx, query, cardID).Scan(&boardID)
	if err == pgx.ErrNoRows {
		return "", apperror.ErrNotFound
	}
	return boardID, err
}
