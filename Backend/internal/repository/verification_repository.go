package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/codewebkhongkho/trello-agent/internal/domain"
	"github.com/codewebkhongkho/trello-agent/pkg/cuid"
)

type verificationRepository struct {
	db *pgxpool.Pool
}

func NewVerificationRepository(db *pgxpool.Pool) VerificationRepository {
	return &verificationRepository{db: db}
}

func (r *verificationRepository) Create(ctx context.Context, v *domain.EmailVerification) error {
	if v.ID == "" {
		v.ID = cuid.New()
	}
	v.CreatedAt = time.Now()

	query := `
		INSERT INTO email_verifications (id, user_id, token, type, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query,
		v.ID,
		v.UserID,
		v.Token,
		v.Type,
		v.ExpiresAt,
		v.CreatedAt,
	)
	return err
}

func (r *verificationRepository) FindByToken(ctx context.Context, token string, tokenType domain.VerificationType) (*domain.EmailVerification, error) {
	query := `
		SELECT id, user_id, token, type, expires_at, used_at, created_at
		FROM email_verifications
		WHERE token = $1 AND type = $2
	`
	var v domain.EmailVerification
	err := r.db.QueryRow(ctx, query, token, tokenType).Scan(
		&v.ID,
		&v.UserID,
		&v.Token,
		&v.Type,
		&v.ExpiresAt,
		&v.UsedAt,
		&v.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *verificationRepository) FindLatestByUserAndType(ctx context.Context, userID string, tokenType domain.VerificationType) (*domain.EmailVerification, error) {
	query := `
		SELECT id, user_id, token, type, expires_at, used_at, created_at
		FROM email_verifications
		WHERE user_id = $1 AND type = $2
		ORDER BY created_at DESC
		LIMIT 1
	`
	var v domain.EmailVerification
	err := r.db.QueryRow(ctx, query, userID, tokenType).Scan(
		&v.ID,
		&v.UserID,
		&v.Token,
		&v.Type,
		&v.ExpiresAt,
		&v.UsedAt,
		&v.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *verificationRepository) MarkUsed(ctx context.Context, id string) error {
	query := `UPDATE email_verifications SET used_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *verificationRepository) DeleteByUserAndType(ctx context.Context, userID string, tokenType domain.VerificationType) error {
	query := `DELETE FROM email_verifications WHERE user_id = $1 AND type = $2`
	_, err := r.db.Exec(ctx, query, userID, tokenType)
	return err
}
