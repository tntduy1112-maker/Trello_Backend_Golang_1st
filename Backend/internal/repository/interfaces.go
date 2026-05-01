package repository

import (
	"context"

	"github.com/codewebkhongkho/trello-agent/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	SoftDelete(ctx context.Context, id string) error
	UpdateTokensValidAfter(ctx context.Context, id string) error
}

type TokenRepository interface {
	CreateRefreshToken(ctx context.Context, token *domain.RefreshToken) error
	FindByHash(ctx context.Context, hash string) (*domain.RefreshToken, error)
	RevokeToken(ctx context.Context, hash string) error
	RevokeAllUserTokens(ctx context.Context, userID string) error
	DeleteExpiredTokens(ctx context.Context) error
}

type VerificationRepository interface {
	Create(ctx context.Context, verification *domain.EmailVerification) error
	FindByToken(ctx context.Context, token string, tokenType domain.VerificationType) (*domain.EmailVerification, error)
	FindLatestByUserAndType(ctx context.Context, userID string, tokenType domain.VerificationType) (*domain.EmailVerification, error)
	MarkUsed(ctx context.Context, id string) error
	DeleteByUserAndType(ctx context.Context, userID string, tokenType domain.VerificationType) error
}
