package repository

import (
	"context"
	"time"

	"github.com/codewebkhongkho/trello-agent/internal/domain"
	"github.com/codewebkhongkho/trello-agent/pkg/apperror"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CommentRepository struct {
	db *pgxpool.Pool
}

func NewCommentRepository(db *pgxpool.Pool) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) Create(ctx context.Context, comment *domain.Comment) error {
	query := `
		INSERT INTO comments (id, card_id, author_id, content, parent_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query,
		comment.ID,
		comment.CardID,
		comment.AuthorID,
		comment.Content,
		comment.ParentID,
		comment.CreatedAt,
		comment.UpdatedAt,
	)
	return err
}

func (r *CommentRepository) FindByID(ctx context.Context, id string) (*domain.Comment, error) {
	query := `
		SELECT c.id, c.card_id, c.author_id, c.content, c.parent_id, c.is_edited,
		       c.created_at, c.updated_at, c.deleted_at,
		       u.id, u.full_name, u.email, u.avatar_url
		FROM comments c
		JOIN users u ON c.author_id = u.id
		WHERE c.id = $1 AND c.deleted_at IS NULL
	`
	row := r.db.QueryRow(ctx, query, id)

	var comment domain.Comment
	var author domain.User

	err := row.Scan(
		&comment.ID, &comment.CardID, &comment.AuthorID, &comment.Content,
		&comment.ParentID, &comment.IsEdited, &comment.CreatedAt, &comment.UpdatedAt, &comment.DeletedAt,
		&author.ID, &author.FullName, &author.Email, &author.AvatarURL,
	)
	if err == pgx.ErrNoRows {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	comment.Author = &author
	return &comment, nil
}

func (r *CommentRepository) FindByCardID(ctx context.Context, cardID string) ([]*domain.Comment, error) {
	query := `
		SELECT c.id, c.card_id, c.author_id, c.content, c.parent_id, c.is_edited,
		       c.created_at, c.updated_at,
		       u.id, u.full_name, u.email, u.avatar_url
		FROM comments c
		JOIN users u ON c.author_id = u.id
		WHERE c.card_id = $1 AND c.deleted_at IS NULL
		ORDER BY c.created_at ASC
	`
	rows, err := r.db.Query(ctx, query, cardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*domain.Comment
	for rows.Next() {
		var comment domain.Comment
		var author domain.User

		err := rows.Scan(
			&comment.ID, &comment.CardID, &comment.AuthorID, &comment.Content,
			&comment.ParentID, &comment.IsEdited, &comment.CreatedAt, &comment.UpdatedAt,
			&author.ID, &author.FullName, &author.Email, &author.AvatarURL,
		)
		if err != nil {
			return nil, err
		}

		comment.Author = &author
		comments = append(comments, &comment)
	}

	return comments, nil
}

func (r *CommentRepository) Update(ctx context.Context, comment *domain.Comment) error {
	query := `
		UPDATE comments
		SET content = $1, is_edited = TRUE, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`
	result, err := r.db.Exec(ctx, query, comment.Content, time.Now(), comment.ID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}

	return nil
}

func (r *CommentRepository) SoftDelete(ctx context.Context, id string) error {
	query := `UPDATE comments SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	result, err := r.db.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}

	return nil
}

func (r *CommentRepository) IsAuthor(ctx context.Context, commentID, userID string) (bool, error) {
	query := `SELECT 1 FROM comments WHERE id = $1 AND author_id = $2 AND deleted_at IS NULL`
	row := r.db.QueryRow(ctx, query, commentID, userID)
	var exists int
	err := row.Scan(&exists)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func (r *CommentRepository) FindCardIDByCommentID(ctx context.Context, commentID string) (string, error) {
	query := `SELECT card_id FROM comments WHERE id = $1`
	row := r.db.QueryRow(ctx, query, commentID)
	var cardID string
	err := row.Scan(&cardID)
	if err == pgx.ErrNoRows {
		return "", apperror.ErrNotFound
	}
	return cardID, err
}
