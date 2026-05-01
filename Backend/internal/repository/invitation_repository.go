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

type InvitationRepository interface {
	Create(ctx context.Context, inv *domain.BoardInvitation) error
	FindByToken(ctx context.Context, token string) (*domain.BoardInvitation, error)
	FindByBoardAndEmail(ctx context.Context, boardID, email string) (*domain.BoardInvitation, error)
	FindPendingByBoardID(ctx context.Context, boardID string) ([]*domain.BoardInvitation, error)
	UpdateStatus(ctx context.Context, id string, status domain.InvitationStatus) error
	Delete(ctx context.Context, id string) error
}

type invitationRepository struct {
	db *pgxpool.Pool
}

func NewInvitationRepository(db *pgxpool.Pool) InvitationRepository {
	return &invitationRepository{db: db}
}

func (r *invitationRepository) Create(ctx context.Context, inv *domain.BoardInvitation) error {
	if inv.ID == "" {
		inv.ID = cuid.New()
	}
	inv.CreatedAt = time.Now()
	if inv.Status == "" {
		inv.Status = domain.InvitationPending
	}

	query := `
		INSERT INTO board_invitations (id, board_id, inviter_id, invitee_id, invitee_email, role, token, message, status, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Exec(ctx, query,
		inv.ID, inv.BoardID, inv.InviterID, inv.InviteeID, inv.InviteeEmail,
		inv.Role, inv.Token, inv.Message, inv.Status, inv.ExpiresAt, inv.CreatedAt,
	)
	return err
}

func (r *invitationRepository) FindByToken(ctx context.Context, token string) (*domain.BoardInvitation, error) {
	query := `
		SELECT bi.id, bi.board_id, bi.inviter_id, bi.invitee_id, bi.invitee_email, bi.role,
		       bi.token, bi.message, bi.status, bi.expires_at, bi.created_at, bi.responded_at,
		       b.id, b.title, b.organization_id,
		       u.id, u.email, u.full_name, u.avatar_url
		FROM board_invitations bi
		INNER JOIN boards b ON bi.board_id = b.id
		INNER JOIN users u ON bi.inviter_id = u.id
		WHERE bi.token = $1
	`
	var inv domain.BoardInvitation
	var board domain.Board
	var inviter domain.User
	err := r.db.QueryRow(ctx, query, token).Scan(
		&inv.ID, &inv.BoardID, &inv.InviterID, &inv.InviteeID, &inv.InviteeEmail, &inv.Role,
		&inv.Token, &inv.Message, &inv.Status, &inv.ExpiresAt, &inv.CreatedAt, &inv.RespondedAt,
		&board.ID, &board.Title, &board.OrganizationID,
		&inviter.ID, &inviter.Email, &inviter.FullName, &inviter.AvatarURL,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	inv.Board = &board
	inv.Inviter = &inviter
	return &inv, nil
}

func (r *invitationRepository) FindByBoardAndEmail(ctx context.Context, boardID, email string) (*domain.BoardInvitation, error) {
	query := `
		SELECT id, board_id, inviter_id, invitee_id, invitee_email, role, token, message, status, expires_at, created_at, responded_at
		FROM board_invitations
		WHERE board_id = $1 AND invitee_email = $2 AND status = 'pending'
		ORDER BY created_at DESC LIMIT 1
	`
	var inv domain.BoardInvitation
	err := r.db.QueryRow(ctx, query, boardID, email).Scan(
		&inv.ID, &inv.BoardID, &inv.InviterID, &inv.InviteeID, &inv.InviteeEmail,
		&inv.Role, &inv.Token, &inv.Message, &inv.Status, &inv.ExpiresAt, &inv.CreatedAt, &inv.RespondedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *invitationRepository) FindPendingByBoardID(ctx context.Context, boardID string) ([]*domain.BoardInvitation, error) {
	query := `
		SELECT bi.id, bi.board_id, bi.inviter_id, bi.invitee_id, bi.invitee_email, bi.role,
		       bi.token, bi.message, bi.status, bi.expires_at, bi.created_at, bi.responded_at,
		       u.id, u.email, u.full_name, u.avatar_url
		FROM board_invitations bi
		INNER JOIN users u ON bi.inviter_id = u.id
		WHERE bi.board_id = $1 AND bi.status = 'pending' AND bi.expires_at > NOW()
		ORDER BY bi.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invitations []*domain.BoardInvitation
	for rows.Next() {
		var inv domain.BoardInvitation
		var inviter domain.User
		if err := rows.Scan(
			&inv.ID, &inv.BoardID, &inv.InviterID, &inv.InviteeID, &inv.InviteeEmail, &inv.Role,
			&inv.Token, &inv.Message, &inv.Status, &inv.ExpiresAt, &inv.CreatedAt, &inv.RespondedAt,
			&inviter.ID, &inviter.Email, &inviter.FullName, &inviter.AvatarURL,
		); err != nil {
			return nil, err
		}
		inv.Inviter = &inviter
		invitations = append(invitations, &inv)
	}
	return invitations, rows.Err()
}

func (r *invitationRepository) UpdateStatus(ctx context.Context, id string, status domain.InvitationStatus) error {
	query := `UPDATE board_invitations SET status = $2, responded_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, status)
	return err
}

func (r *invitationRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM board_invitations WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
