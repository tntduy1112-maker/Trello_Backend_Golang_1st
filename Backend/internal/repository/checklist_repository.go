package repository

import (
	"context"
	"time"

	"github.com/codewebkhongkho/trello-agent/internal/domain"
	"github.com/codewebkhongkho/trello-agent/pkg/apperror"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChecklistRepository struct {
	db *pgxpool.Pool
}

func NewChecklistRepository(db *pgxpool.Pool) *ChecklistRepository {
	return &ChecklistRepository{db: db}
}

func (r *ChecklistRepository) Create(ctx context.Context, checklist *domain.Checklist) error {
	query := `
		INSERT INTO checklists (id, card_id, title, position, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query,
		checklist.ID,
		checklist.CardID,
		checklist.Title,
		checklist.Position,
		checklist.CreatedAt,
		checklist.UpdatedAt,
	)
	return err
}

func (r *ChecklistRepository) FindByID(ctx context.Context, id string) (*domain.Checklist, error) {
	query := `SELECT id, card_id, title, position, created_at, updated_at FROM checklists WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)

	var checklist domain.Checklist
	err := row.Scan(&checklist.ID, &checklist.CardID, &checklist.Title, &checklist.Position, &checklist.CreatedAt, &checklist.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, apperror.ErrNotFound
	}
	return &checklist, err
}

func (r *ChecklistRepository) FindByCardID(ctx context.Context, cardID string) ([]*domain.Checklist, error) {
	query := `SELECT id, card_id, title, position, created_at, updated_at FROM checklists WHERE card_id = $1 ORDER BY position`
	rows, err := r.db.Query(ctx, query, cardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checklists []*domain.Checklist
	for rows.Next() {
		var cl domain.Checklist
		if err := rows.Scan(&cl.ID, &cl.CardID, &cl.Title, &cl.Position, &cl.CreatedAt, &cl.UpdatedAt); err != nil {
			return nil, err
		}
		checklists = append(checklists, &cl)
	}
	return checklists, nil
}

func (r *ChecklistRepository) FindByCardIDWithItems(ctx context.Context, cardID string) ([]*domain.Checklist, error) {
	checklists, err := r.FindByCardID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	for _, cl := range checklists {
		items, err := r.FindItemsByChecklistID(ctx, cl.ID)
		if err != nil {
			return nil, err
		}
		cl.Items = items

		completed := 0
		for _, item := range items {
			if item.IsCompleted {
				completed++
			}
		}
		cl.Progress = &domain.ChecklistProgress{
			Completed: completed,
			Total:     len(items),
		}
	}

	return checklists, nil
}

func (r *ChecklistRepository) Update(ctx context.Context, checklist *domain.Checklist) error {
	query := `UPDATE checklists SET title = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.Exec(ctx, query, checklist.Title, time.Now(), checklist.ID)
	return err
}

func (r *ChecklistRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM checklists WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *ChecklistRepository) GetMaxPosition(ctx context.Context, cardID string) (float64, error) {
	query := `SELECT COALESCE(MAX(position), 0) FROM checklists WHERE card_id = $1`
	var maxPos float64
	err := r.db.QueryRow(ctx, query, cardID).Scan(&maxPos)
	return maxPos, err
}

func (r *ChecklistRepository) FindCardIDByChecklistID(ctx context.Context, checklistID string) (string, error) {
	query := `SELECT card_id FROM checklists WHERE id = $1`
	var cardID string
	err := r.db.QueryRow(ctx, query, checklistID).Scan(&cardID)
	if err == pgx.ErrNoRows {
		return "", apperror.ErrNotFound
	}
	return cardID, err
}

// Checklist Items

func (r *ChecklistRepository) CreateItem(ctx context.Context, item *domain.ChecklistItem) error {
	query := `
		INSERT INTO checklist_items (id, checklist_id, title, position, assignee_id, due_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(ctx, query,
		item.ID, item.ChecklistID, item.Title, item.Position,
		item.AssigneeID, item.DueDate, item.CreatedAt, item.UpdatedAt,
	)
	return err
}

func (r *ChecklistRepository) FindItemByID(ctx context.Context, id string) (*domain.ChecklistItem, error) {
	query := `
		SELECT id, checklist_id, title, position, is_completed, completed_at, completed_by, assignee_id, due_date, created_at, updated_at
		FROM checklist_items WHERE id = $1
	`
	row := r.db.QueryRow(ctx, query, id)

	var item domain.ChecklistItem
	err := row.Scan(&item.ID, &item.ChecklistID, &item.Title, &item.Position,
		&item.IsCompleted, &item.CompletedAt, &item.CompletedBy, &item.AssigneeID,
		&item.DueDate, &item.CreatedAt, &item.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, apperror.ErrNotFound
	}
	return &item, err
}

func (r *ChecklistRepository) FindItemsByChecklistID(ctx context.Context, checklistID string) ([]domain.ChecklistItem, error) {
	query := `
		SELECT id, checklist_id, title, position, is_completed, completed_at, completed_by, assignee_id, due_date, created_at, updated_at
		FROM checklist_items WHERE checklist_id = $1 ORDER BY position
	`
	rows, err := r.db.Query(ctx, query, checklistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.ChecklistItem
	for rows.Next() {
		var item domain.ChecklistItem
		if err := rows.Scan(&item.ID, &item.ChecklistID, &item.Title, &item.Position,
			&item.IsCompleted, &item.CompletedAt, &item.CompletedBy, &item.AssigneeID,
			&item.DueDate, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *ChecklistRepository) UpdateItem(ctx context.Context, item *domain.ChecklistItem) error {
	query := `
		UPDATE checklist_items
		SET title = $1, assignee_id = $2, due_date = $3, updated_at = $4
		WHERE id = $5
	`
	_, err := r.db.Exec(ctx, query, item.Title, item.AssigneeID, item.DueDate, time.Now(), item.ID)
	return err
}

func (r *ChecklistRepository) DeleteItem(ctx context.Context, id string) error {
	query := `DELETE FROM checklist_items WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *ChecklistRepository) ToggleItemComplete(ctx context.Context, id, userID string) (*domain.ChecklistItem, error) {
	item, err := r.FindItemByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var query string
	if item.IsCompleted {
		query = `UPDATE checklist_items SET is_completed = FALSE, completed_at = NULL, completed_by = NULL, updated_at = $1 WHERE id = $2`
		_, err = r.db.Exec(ctx, query, time.Now(), id)
	} else {
		now := time.Now()
		query = `UPDATE checklist_items SET is_completed = TRUE, completed_at = $1, completed_by = $2, updated_at = $1 WHERE id = $3`
		_, err = r.db.Exec(ctx, query, now, userID, id)
	}

	if err != nil {
		return nil, err
	}

	return r.FindItemByID(ctx, id)
}

func (r *ChecklistRepository) GetItemMaxPosition(ctx context.Context, checklistID string) (float64, error) {
	query := `SELECT COALESCE(MAX(position), 0) FROM checklist_items WHERE checklist_id = $1`
	var maxPos float64
	err := r.db.QueryRow(ctx, query, checklistID).Scan(&maxPos)
	return maxPos, err
}
