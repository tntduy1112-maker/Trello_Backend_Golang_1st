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

type tokenRepository struct {
	db *pgxpool.Pool
}

func NewTokenRepository(db *pgxpool.Pool) TokenRepository {
	return &tokenRepository{db: db}
}

func (r *tokenRepository) CreateRefreshToken(ctx context.Context, token *domain.RefreshToken) error {
	if token.ID == "" {
		token.ID = cuid.New()
	}
	token.CreatedAt = time.Now()

	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, device_info, ip_address, is_revoked, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(ctx, query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.DeviceInfo,
		token.IPAddress,
		token.IsRevoked,
		token.ExpiresAt,
		token.CreatedAt,
	)
	return err
}

func (r *tokenRepository) FindByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, device_info, ip_address, is_revoked, expires_at, created_at, revoked_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`
	var token domain.RefreshToken
	err := r.db.QueryRow(ctx, query, hash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.DeviceInfo,
		&token.IPAddress,
		&token.IsRevoked,
		&token.ExpiresAt,
		&token.CreatedAt,
		&token.RevokedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &token, nil
}

func (r *tokenRepository) RevokeToken(ctx context.Context, hash string) error {
	query := `UPDATE refresh_tokens SET is_revoked = true, revoked_at = NOW() WHERE token_hash = $1`
	_, err := r.db.Exec(ctx, query, hash)
	return err
}

func (r *tokenRepository) RevokeAllUserTokens(ctx context.Context, userID string) error {
	query := `UPDATE refresh_tokens SET is_revoked = true, revoked_at = NOW() WHERE user_id = $1 AND is_revoked = false`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

func (r *tokenRepository) DeleteExpiredTokens(ctx context.Context) error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < NOW()`
	_, err := r.db.Exec(ctx, query)
	return err
}
