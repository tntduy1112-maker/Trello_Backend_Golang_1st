package repository

import (
	"context"
	"time"

	"github.com/codewebkhongkho/trello-agent/internal/domain"
	"github.com/codewebkhongkho/trello-agent/pkg/apperror"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AttachmentRepository struct {
	db *pgxpool.Pool
}

func NewAttachmentRepository(db *pgxpool.Pool) *AttachmentRepository {
	return &AttachmentRepository{db: db}
}

func (r *AttachmentRepository) Create(ctx context.Context, attachment *domain.Attachment) error {
	query := `
		INSERT INTO attachments (id, card_id, uploaded_by, filename, original_name, mime_type, file_size, object_key, url, is_cover, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Exec(ctx, query,
		attachment.ID,
		attachment.CardID,
		attachment.UploadedBy,
		attachment.Filename,
		attachment.OriginalName,
		attachment.MimeType,
		attachment.FileSize,
		attachment.ObjectKey,
		attachment.URL,
		attachment.IsCover,
		attachment.CreatedAt,
	)
	return err
}

func (r *AttachmentRepository) FindByID(ctx context.Context, id string) (*domain.Attachment, error) {
	query := `
		SELECT a.id, a.card_id, a.uploaded_by, a.filename, a.original_name, a.mime_type,
		       a.file_size, a.object_key, a.url, a.is_cover, a.created_at,
		       u.id, u.full_name, u.avatar_url
		FROM attachments a
		JOIN users u ON a.uploaded_by = u.id
		WHERE a.id = $1 AND a.deleted_at IS NULL
	`
	row := r.db.QueryRow(ctx, query, id)

	var att domain.Attachment
	var uploader domain.User

	err := row.Scan(
		&att.ID, &att.CardID, &att.UploadedBy, &att.Filename, &att.OriginalName,
		&att.MimeType, &att.FileSize, &att.ObjectKey, &att.URL, &att.IsCover, &att.CreatedAt,
		&uploader.ID, &uploader.FullName, &uploader.AvatarURL,
	)
	if err == pgx.ErrNoRows {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	att.Uploader = &uploader
	return &att, nil
}

func (r *AttachmentRepository) FindByCardID(ctx context.Context, cardID string) ([]*domain.Attachment, error) {
	query := `
		SELECT a.id, a.card_id, a.uploaded_by, a.filename, a.original_name, a.mime_type,
		       a.file_size, a.object_key, a.url, a.is_cover, a.created_at,
		       u.id, u.full_name, u.avatar_url
		FROM attachments a
		JOIN users u ON a.uploaded_by = u.id
		WHERE a.card_id = $1 AND a.deleted_at IS NULL
		ORDER BY a.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, cardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attachments []*domain.Attachment
	for rows.Next() {
		var att domain.Attachment
		var uploader domain.User

		err := rows.Scan(
			&att.ID, &att.CardID, &att.UploadedBy, &att.Filename, &att.OriginalName,
			&att.MimeType, &att.FileSize, &att.ObjectKey, &att.URL, &att.IsCover, &att.CreatedAt,
			&uploader.ID, &uploader.FullName, &uploader.AvatarURL,
		)
		if err != nil {
			return nil, err
		}

		att.Uploader = &uploader
		attachments = append(attachments, &att)
	}

	return attachments, nil
}

func (r *AttachmentRepository) SoftDelete(ctx context.Context, id string) error {
	query := `UPDATE attachments SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	result, err := r.db.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}

	return nil
}

func (r *AttachmentRepository) SetCover(ctx context.Context, cardID, attachmentID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Remove current cover
	_, err = tx.Exec(ctx, `UPDATE attachments SET is_cover = FALSE WHERE card_id = $1`, cardID)
	if err != nil {
		return err
	}

	// Set new cover
	_, err = tx.Exec(ctx, `UPDATE attachments SET is_cover = TRUE WHERE id = $1`, attachmentID)
	if err != nil {
		return err
	}

	// Update card's cover_attachment_id
	_, err = tx.Exec(ctx, `UPDATE cards SET cover_attachment_id = $1 WHERE id = $2`, attachmentID, cardID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *AttachmentRepository) RemoveCover(ctx context.Context, cardID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `UPDATE attachments SET is_cover = FALSE WHERE card_id = $1`, cardID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `UPDATE cards SET cover_attachment_id = NULL WHERE id = $1`, cardID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *AttachmentRepository) GetCover(ctx context.Context, cardID string) (*domain.Attachment, error) {
	query := `
		SELECT id, card_id, uploaded_by, filename, original_name, mime_type, file_size, object_key, url, is_cover, created_at
		FROM attachments
		WHERE card_id = $1 AND is_cover = TRUE AND deleted_at IS NULL
	`
	row := r.db.QueryRow(ctx, query, cardID)

	var att domain.Attachment
	err := row.Scan(&att.ID, &att.CardID, &att.UploadedBy, &att.Filename, &att.OriginalName,
		&att.MimeType, &att.FileSize, &att.ObjectKey, &att.URL, &att.IsCover, &att.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &att, err
}

func (r *AttachmentRepository) FindCardIDByAttachmentID(ctx context.Context, attachmentID string) (string, error) {
	query := `SELECT card_id FROM attachments WHERE id = $1`
	var cardID string
	err := r.db.QueryRow(ctx, query, attachmentID).Scan(&cardID)
	if err == pgx.ErrNoRows {
		return "", apperror.ErrNotFound
	}
	return cardID, err
}

func (r *AttachmentRepository) IsUploader(ctx context.Context, attachmentID, userID string) (bool, error) {
	query := `SELECT 1 FROM attachments WHERE id = $1 AND uploaded_by = $2 AND deleted_at IS NULL`
	row := r.db.QueryRow(ctx, query, attachmentID, userID)
	var exists int
	err := row.Scan(&exists)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}
