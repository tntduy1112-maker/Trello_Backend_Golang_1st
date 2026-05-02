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

type BoardRepository interface {
	Create(ctx context.Context, board *domain.Board) error
	FindByID(ctx context.Context, id string) (*domain.Board, error)
	FindByOrgID(ctx context.Context, orgID string, includeClosed bool) ([]*domain.Board, error)
	Update(ctx context.Context, board *domain.Board) error
	Close(ctx context.Context, id string) error
	Reopen(ctx context.Context, id string) error
	SoftDelete(ctx context.Context, id string) error

	AddMember(ctx context.Context, member *domain.BoardMember) error
	FindMember(ctx context.Context, boardID, userID string) (*domain.BoardMember, error)
	FindMembers(ctx context.Context, boardID string) ([]*domain.BoardMember, error)
	FindMembersByOrgID(ctx context.Context, orgID string) ([]*domain.BoardMember, error)
	UpdateMemberRole(ctx context.Context, boardID, userID string, role domain.BoardRole) error
	RemoveMember(ctx context.Context, boardID, userID string) error

	CanUserAccess(ctx context.Context, boardID, userID string) (bool, domain.BoardRole, error)
	IsWorkspaceOwner(ctx context.Context, orgID, userID string) (bool, error)
	CountLists(ctx context.Context, boardID string) (int, error)
	CountCards(ctx context.Context, boardID string) (int, error)
}

type boardRepository struct {
	db *pgxpool.Pool
}

func NewBoardRepository(db *pgxpool.Pool) BoardRepository {
	return &boardRepository{db: db}
}

func (r *boardRepository) Create(ctx context.Context, board *domain.Board) error {
	if board.ID == "" {
		board.ID = cuid.New()
	}
	now := time.Now()
	board.CreatedAt = now
	board.UpdatedAt = now
	if board.BackgroundColor == "" {
		board.BackgroundColor = domain.DefaultBackgroundColor
	}
	if board.Visibility == "" {
		board.Visibility = domain.VisibilityWorkspace
	}

	query := `
		INSERT INTO boards (id, organization_id, title, description, background_color, background_image, visibility, is_closed, owner_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Exec(ctx, query,
		board.ID, board.OrganizationID, board.Title, board.Description,
		board.BackgroundColor, board.BackgroundImage, board.Visibility,
		board.IsClosed, board.OwnerID, board.CreatedAt, board.UpdatedAt,
	)
	return err
}

func (r *boardRepository) FindByID(ctx context.Context, id string) (*domain.Board, error) {
	query := `
		SELECT b.id, b.organization_id, b.title, b.description, b.background_color, b.background_image,
		       b.visibility, b.is_closed, b.owner_id, b.created_at, b.updated_at, b.closed_at, b.deleted_at,
		       o.id, o.name, o.slug
		FROM boards b
		INNER JOIN organizations o ON b.organization_id = o.id
		WHERE b.id = $1 AND b.deleted_at IS NULL
	`
	var board domain.Board
	var org domain.Organization
	err := r.db.QueryRow(ctx, query, id).Scan(
		&board.ID, &board.OrganizationID, &board.Title, &board.Description,
		&board.BackgroundColor, &board.BackgroundImage, &board.Visibility,
		&board.IsClosed, &board.OwnerID, &board.CreatedAt, &board.UpdatedAt,
		&board.ClosedAt, &board.DeletedAt, &org.ID, &org.Name, &org.Slug,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	board.Organization = &org
	return &board, nil
}

func (r *boardRepository) FindByOrgID(ctx context.Context, orgID string, includeClosed bool) ([]*domain.Board, error) {
	query := `
		SELECT id, organization_id, title, description, background_color, background_image,
		       visibility, is_closed, owner_id, created_at, updated_at, closed_at, deleted_at
		FROM boards
		WHERE organization_id = $1 AND deleted_at IS NULL
	`
	if !includeClosed {
		query += ` AND is_closed = false`
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var boards []*domain.Board
	for rows.Next() {
		var board domain.Board
		if err := rows.Scan(
			&board.ID, &board.OrganizationID, &board.Title, &board.Description,
			&board.BackgroundColor, &board.BackgroundImage, &board.Visibility,
			&board.IsClosed, &board.OwnerID, &board.CreatedAt, &board.UpdatedAt,
			&board.ClosedAt, &board.DeletedAt,
		); err != nil {
			return nil, err
		}
		boards = append(boards, &board)
	}
	return boards, rows.Err()
}

func (r *boardRepository) Update(ctx context.Context, board *domain.Board) error {
	board.UpdatedAt = time.Now()
	query := `
		UPDATE boards SET title = $2, description = $3, background_color = $4,
		       background_image = $5, visibility = $6, updated_at = $7
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query,
		board.ID, board.Title, board.Description, board.BackgroundColor,
		board.BackgroundImage, board.Visibility, board.UpdatedAt,
	)
	return err
}

func (r *boardRepository) Close(ctx context.Context, id string) error {
	query := `UPDATE boards SET is_closed = true, closed_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *boardRepository) Reopen(ctx context.Context, id string) error {
	query := `UPDATE boards SET is_closed = false, closed_at = NULL, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *boardRepository) SoftDelete(ctx context.Context, id string) error {
	query := `UPDATE boards SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *boardRepository) AddMember(ctx context.Context, member *domain.BoardMember) error {
	if member.ID == "" {
		member.ID = cuid.New()
	}
	member.JoinedAt = time.Now()

	query := `
		INSERT INTO board_members (id, board_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(ctx, query, member.ID, member.BoardID, member.UserID, member.Role, member.JoinedAt)
	return err
}

func (r *boardRepository) FindMember(ctx context.Context, boardID, userID string) (*domain.BoardMember, error) {
	query := `
		SELECT bm.id, bm.board_id, bm.user_id, bm.role, bm.joined_at,
		       u.id, u.email, u.full_name, u.avatar_url
		FROM board_members bm
		INNER JOIN users u ON bm.user_id = u.id
		WHERE bm.board_id = $1 AND bm.user_id = $2
	`
	var member domain.BoardMember
	var user domain.User
	err := r.db.QueryRow(ctx, query, boardID, userID).Scan(
		&member.ID, &member.BoardID, &member.UserID, &member.Role, &member.JoinedAt,
		&user.ID, &user.Email, &user.FullName, &user.AvatarURL,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	member.User = &user
	return &member, nil
}

func (r *boardRepository) FindMembers(ctx context.Context, boardID string) ([]*domain.BoardMember, error) {
	query := `
		SELECT bm.id, bm.board_id, bm.user_id, bm.role, bm.joined_at,
		       u.id, u.email, u.full_name, u.avatar_url
		FROM board_members bm
		INNER JOIN users u ON bm.user_id = u.id
		WHERE bm.board_id = $1
		ORDER BY bm.role, bm.joined_at
	`
	rows, err := r.db.Query(ctx, query, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*domain.BoardMember
	for rows.Next() {
		var member domain.BoardMember
		var user domain.User
		if err := rows.Scan(
			&member.ID, &member.BoardID, &member.UserID, &member.Role, &member.JoinedAt,
			&user.ID, &user.Email, &user.FullName, &user.AvatarURL,
		); err != nil {
			return nil, err
		}
		member.User = &user
		members = append(members, &member)
	}
	return members, rows.Err()
}

func (r *boardRepository) FindMembersByOrgID(ctx context.Context, orgID string) ([]*domain.BoardMember, error) {
	query := `
		SELECT bm.id, bm.board_id, bm.user_id, bm.role, bm.joined_at,
		       u.id, u.email, u.full_name, u.avatar_url,
		       b.id, b.title
		FROM board_members bm
		INNER JOIN users u ON bm.user_id = u.id
		INNER JOIN boards b ON bm.board_id = b.id
		WHERE b.organization_id = $1 AND b.deleted_at IS NULL
		ORDER BY u.full_name, b.title
	`
	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*domain.BoardMember
	for rows.Next() {
		var member domain.BoardMember
		var user domain.User
		var board domain.Board
		if err := rows.Scan(
			&member.ID, &member.BoardID, &member.UserID, &member.Role, &member.JoinedAt,
			&user.ID, &user.Email, &user.FullName, &user.AvatarURL,
			&board.ID, &board.Title,
		); err != nil {
			return nil, err
		}
		member.User = &user
		member.Board = &board
		members = append(members, &member)
	}
	return members, rows.Err()
}

func (r *boardRepository) UpdateMemberRole(ctx context.Context, boardID, userID string, role domain.BoardRole) error {
	query := `UPDATE board_members SET role = $3 WHERE board_id = $1 AND user_id = $2`
	_, err := r.db.Exec(ctx, query, boardID, userID, role)
	return err
}

func (r *boardRepository) RemoveMember(ctx context.Context, boardID, userID string) error {
	query := `DELETE FROM board_members WHERE board_id = $1 AND user_id = $2`
	_, err := r.db.Exec(ctx, query, boardID, userID)
	return err
}

func (r *boardRepository) CanUserAccess(ctx context.Context, boardID, userID string) (bool, domain.BoardRole, error) {
	// Board-centric model: only explicit board members have access
	member, err := r.FindMember(ctx, boardID, userID)
	if err != nil {
		return false, "", err
	}
	if member != nil {
		return true, member.Role, nil
	}

	// No access if not a board member
	return false, "", nil
}

func (r *boardRepository) IsWorkspaceOwner(ctx context.Context, orgID, userID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM board_members bm
			INNER JOIN boards b ON bm.board_id = b.id
			WHERE b.organization_id = $1
			AND bm.user_id = $2
			AND bm.role = 'owner'
			AND b.deleted_at IS NULL
		)
	`
	var exists bool
	if err := r.db.QueryRow(ctx, query, orgID, userID).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *boardRepository) CountLists(ctx context.Context, boardID string) (int, error) {
	query := `SELECT COUNT(*) FROM lists WHERE board_id = $1 AND is_archived = false`
	var count int
	err := r.db.QueryRow(ctx, query, boardID).Scan(&count)
	if err != nil && err.Error() == `ERROR: relation "lists" does not exist (SQLSTATE 42P01)` {
		return 0, nil
	}
	return count, err
}

func (r *boardRepository) CountCards(ctx context.Context, boardID string) (int, error) {
	query := `
		SELECT COUNT(*) FROM cards c
		INNER JOIN lists l ON c.list_id = l.id
		WHERE l.board_id = $1 AND c.is_archived = false
	`
	var count int
	err := r.db.QueryRow(ctx, query, boardID).Scan(&count)
	if err != nil && err.Error() == `ERROR: relation "cards" does not exist (SQLSTATE 42P01)` {
		return 0, nil
	}
	return count, err
}
