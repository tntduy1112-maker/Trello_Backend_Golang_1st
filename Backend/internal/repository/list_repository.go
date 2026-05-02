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

type ListRepository struct {
	db *pgxpool.Pool
}

func NewListRepository(db *pgxpool.Pool) *ListRepository {
	return &ListRepository{db: db}
}

func (r *ListRepository) Create(ctx context.Context, list *domain.List) error {
	query := `
		INSERT INTO lists (id, board_id, title, position, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query,
		list.ID, list.BoardID, list.Title, list.Position,
		list.CreatedAt, list.UpdatedAt,
	)
	return err
}

func (r *ListRepository) FindByID(ctx context.Context, id string) (*domain.List, error) {
	query := `
		SELECT id, board_id, title, position, is_archived, created_at, updated_at, archived_at
		FROM lists WHERE id = $1
	`
	row := r.db.QueryRow(ctx, query, id)

	var list domain.List
	err := row.Scan(
		&list.ID, &list.BoardID, &list.Title, &list.Position,
		&list.IsArchived, &list.CreatedAt, &list.UpdatedAt, &list.ArchivedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &list, nil
}

func (r *ListRepository) FindByBoardID(ctx context.Context, boardID string, includeArchived bool) ([]*domain.List, error) {
	query := `
		SELECT l.id, l.board_id, l.title, l.position, l.is_archived,
			   l.created_at, l.updated_at, l.archived_at,
			   (SELECT COUNT(*) FROM cards c WHERE c.list_id = l.id AND c.is_archived = false) as cards_count
		FROM lists l
		WHERE l.board_id = $1
	`
	if !includeArchived {
		query += " AND l.is_archived = false"
	}
	query += " ORDER BY l.position ASC"

	rows, err := r.db.Query(ctx, query, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lists []*domain.List
	for rows.Next() {
		var list domain.List
		err := rows.Scan(
			&list.ID, &list.BoardID, &list.Title, &list.Position,
			&list.IsArchived, &list.CreatedAt, &list.UpdatedAt, &list.ArchivedAt,
			&list.CardsCount,
		)
		if err != nil {
			return nil, err
		}
		lists = append(lists, &list)
	}
	return lists, nil
}

func (r *ListRepository) Update(ctx context.Context, list *domain.List) error {
	query := `
		UPDATE lists SET title = $1, updated_at = $2
		WHERE id = $3
	`
	_, err := r.db.Exec(ctx, query, list.Title, time.Now(), list.ID)
	return err
}

func (r *ListRepository) Archive(ctx context.Context, id string) error {
	now := time.Now()
	query := `
		UPDATE lists SET is_archived = true, archived_at = $1, updated_at = $1
		WHERE id = $2
	`
	_, err := r.db.Exec(ctx, query, now, id)
	return err
}

func (r *ListRepository) Restore(ctx context.Context, id string) error {
	query := `
		UPDATE lists SET is_archived = false, archived_at = NULL, updated_at = $1
		WHERE id = $2
	`
	_, err := r.db.Exec(ctx, query, time.Now(), id)
	return err
}

func (r *ListRepository) GetMaxPosition(ctx context.Context, boardID string) (float64, error) {
	query := `SELECT COALESCE(MAX(position), 0) FROM lists WHERE board_id = $1`
	var maxPos float64
	err := r.db.QueryRow(ctx, query, boardID).Scan(&maxPos)
	return maxPos, err
}

func (r *ListRepository) UpdatePosition(ctx context.Context, id string, pos float64) error {
	query := `UPDATE lists SET position = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.Exec(ctx, query, pos, time.Now(), id)
	return err
}

func (r *ListRepository) RebalancePositions(ctx context.Context, boardID string) error {
	query := `
		SELECT id FROM lists
		WHERE board_id = $1 AND is_archived = false
		ORDER BY position ASC
	`
	rows, err := r.db.Query(ctx, query, boardID)
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
		if err := r.UpdatePosition(ctx, id, positions[i]); err != nil {
			return err
		}
	}
	return nil
}

func (r *ListRepository) FindByBoardIDWithCards(ctx context.Context, boardID string) ([]*domain.List, error) {
	lists, err := r.FindByBoardID(ctx, boardID, false)
	if err != nil {
		return nil, err
	}

	cardQuery := `
		SELECT c.id, c.list_id, c.title, c.description, c.position, c.assignee_id,
			   c.priority, c.due_date, c.is_completed, c.completed_at,
			   c.created_at, c.updated_at,
			   u.id, u.full_name, u.avatar_url
		FROM cards c
		LEFT JOIN users u ON c.assignee_id = u.id
		WHERE c.list_id = ANY($1) AND c.is_archived = false
		ORDER BY c.position ASC
	`

	listIDs := make([]string, len(lists))
	listMap := make(map[string]*domain.List)
	for i, l := range lists {
		listIDs[i] = l.ID
		listMap[l.ID] = l
		l.Cards = []*domain.Card{}
	}

	if len(listIDs) == 0 {
		return lists, nil
	}

	rows, err := r.db.Query(ctx, cardQuery, listIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var card domain.Card
		var assigneeID, assigneeName, assigneeAvatar *string

		err := rows.Scan(
			&card.ID, &card.ListID, &card.Title, &card.Description, &card.Position,
			&card.AssigneeID, &card.Priority, &card.DueDate, &card.IsCompleted,
			&card.CompletedAt, &card.CreatedAt, &card.UpdatedAt,
			&assigneeID, &assigneeName, &assigneeAvatar,
		)
		if err != nil {
			return nil, err
		}

		if assigneeID != nil {
			card.Assignee = &domain.User{
				ID:        *assigneeID,
				FullName:  *assigneeName,
				AvatarURL: assigneeAvatar,
			}
		}

		if list, ok := listMap[card.ListID]; ok {
			list.Cards = append(list.Cards, &card)
		}
	}

	return lists, nil
}

func (r *ListRepository) FindBoardIDByListID(ctx context.Context, listID string) (string, error) {
	query := `SELECT board_id FROM lists WHERE id = $1`
	var boardID string
	err := r.db.QueryRow(ctx, query, listID).Scan(&boardID)
	if err == pgx.ErrNoRows {
		return "", apperror.ErrNotFound
	}
	return boardID, err
}
