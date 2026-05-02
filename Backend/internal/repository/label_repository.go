package repository

import (
	"context"
	"time"

	"github.com/codewebkhongkho/trello-agent/internal/domain"
	"github.com/codewebkhongkho/trello-agent/pkg/apperror"
	"github.com/codewebkhongkho/trello-agent/pkg/cuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LabelRepository struct {
	db *pgxpool.Pool
}

func NewLabelRepository(db *pgxpool.Pool) *LabelRepository {
	return &LabelRepository{db: db}
}

func (r *LabelRepository) Create(ctx context.Context, label *domain.Label) error {
	query := `
		INSERT INTO labels (id, board_id, name, color, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query,
		label.ID, label.BoardID, label.Name, label.Color,
		label.CreatedAt, label.UpdatedAt,
	)
	return err
}

func (r *LabelRepository) FindByID(ctx context.Context, id string) (*domain.Label, error) {
	query := `SELECT id, board_id, name, color, created_at, updated_at FROM labels WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)

	var label domain.Label
	err := row.Scan(&label.ID, &label.BoardID, &label.Name, &label.Color, &label.CreatedAt, &label.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &label, nil
}

func (r *LabelRepository) FindByBoardID(ctx context.Context, boardID string) ([]*domain.Label, error) {
	query := `SELECT id, board_id, name, color, created_at, updated_at FROM labels WHERE board_id = $1 ORDER BY created_at ASC`
	rows, err := r.db.Query(ctx, query, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var labels []*domain.Label
	for rows.Next() {
		var label domain.Label
		err := rows.Scan(&label.ID, &label.BoardID, &label.Name, &label.Color, &label.CreatedAt, &label.UpdatedAt)
		if err != nil {
			return nil, err
		}
		labels = append(labels, &label)
	}
	return labels, nil
}

func (r *LabelRepository) Update(ctx context.Context, label *domain.Label) error {
	query := `UPDATE labels SET name = $1, color = $2, updated_at = $3 WHERE id = $4`
	_, err := r.db.Exec(ctx, query, label.Name, label.Color, time.Now(), label.ID)
	return err
}

func (r *LabelRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM labels WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *LabelRepository) AssignToCard(ctx context.Context, cardID, labelID string) error {
	query := `
		INSERT INTO card_labels (id, card_id, label_id, assigned_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (card_id, label_id) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, cuid.New(), cardID, labelID, time.Now())
	return err
}

func (r *LabelRepository) RemoveFromCard(ctx context.Context, cardID, labelID string) error {
	query := `DELETE FROM card_labels WHERE card_id = $1 AND label_id = $2`
	_, err := r.db.Exec(ctx, query, cardID, labelID)
	return err
}

func (r *LabelRepository) FindByCardID(ctx context.Context, cardID string) ([]*domain.Label, error) {
	query := `
		SELECT l.id, l.board_id, l.name, l.color, l.created_at, l.updated_at
		FROM labels l
		JOIN card_labels cl ON l.id = cl.label_id
		WHERE cl.card_id = $1
	`
	rows, err := r.db.Query(ctx, query, cardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var labels []*domain.Label
	for rows.Next() {
		var label domain.Label
		err := rows.Scan(&label.ID, &label.BoardID, &label.Name, &label.Color, &label.CreatedAt, &label.UpdatedAt)
		if err != nil {
			return nil, err
		}
		labels = append(labels, &label)
	}
	return labels, nil
}

func (r *LabelRepository) IsAssignedToCard(ctx context.Context, cardID, labelID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM card_labels WHERE card_id = $1 AND label_id = $2)`
	var exists bool
	err := r.db.QueryRow(ctx, query, cardID, labelID).Scan(&exists)
	return exists, err
}

func (r *LabelRepository) FindBoardIDByLabelID(ctx context.Context, labelID string) (string, error) {
	query := `SELECT board_id FROM labels WHERE id = $1`
	var boardID string
	err := r.db.QueryRow(ctx, query, labelID).Scan(&boardID)
	if err == pgx.ErrNoRows {
		return "", apperror.ErrNotFound
	}
	return boardID, err
}
