package repository

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/codewebkhongkho/trello-agent/internal/domain"
	"github.com/codewebkhongkho/trello-agent/pkg/cuid"
)

type OrganizationRepository interface {
	Create(ctx context.Context, org *domain.Organization) error
	FindByID(ctx context.Context, id string) (*domain.Organization, error)
	FindBySlug(ctx context.Context, slug string) (*domain.Organization, error)
	FindByUserID(ctx context.Context, userID string) ([]*domain.Organization, error)
	Update(ctx context.Context, org *domain.Organization) error
	SoftDelete(ctx context.Context, id string) error

	SlugExists(ctx context.Context, slug string) (bool, error)
	GenerateUniqueSlug(ctx context.Context, name string) (string, error)

	AddMember(ctx context.Context, member *domain.OrganizationMember) error
	FindMember(ctx context.Context, orgID, userID string) (*domain.OrganizationMember, error)
	FindMemberBySlug(ctx context.Context, slug, userID string) (*domain.OrganizationMember, error)
	FindMembers(ctx context.Context, orgID string) ([]*domain.OrganizationMember, error)
	UpdateMemberRole(ctx context.Context, orgID, userID string, role domain.OrgRole) error
	RemoveMember(ctx context.Context, orgID, userID string) error
	CountMembers(ctx context.Context, orgID string) (int, error)
	CountBoards(ctx context.Context, orgID string) (int, error)
}

type organizationRepository struct {
	db *pgxpool.Pool
}

func NewOrganizationRepository(db *pgxpool.Pool) OrganizationRepository {
	return &organizationRepository{db: db}
}

func (r *organizationRepository) Create(ctx context.Context, org *domain.Organization) error {
	if org.ID == "" {
		org.ID = cuid.New()
	}
	now := time.Now()
	org.CreatedAt = now
	org.UpdatedAt = now

	query := `
		INSERT INTO organizations (id, name, slug, description, logo_url, owner_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(ctx, query,
		org.ID, org.Name, org.Slug, org.Description, org.LogoURL, org.OwnerID, org.CreatedAt, org.UpdatedAt,
	)
	return err
}

func (r *organizationRepository) FindByID(ctx context.Context, id string) (*domain.Organization, error) {
	query := `
		SELECT id, name, slug, description, logo_url, owner_id, created_at, updated_at, deleted_at
		FROM organizations WHERE id = $1 AND deleted_at IS NULL
	`
	var org domain.Organization
	err := r.db.QueryRow(ctx, query, id).Scan(
		&org.ID, &org.Name, &org.Slug, &org.Description, &org.LogoURL,
		&org.OwnerID, &org.CreatedAt, &org.UpdatedAt, &org.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *organizationRepository) FindBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	query := `
		SELECT id, name, slug, description, logo_url, owner_id, created_at, updated_at, deleted_at
		FROM organizations WHERE slug = $1 AND deleted_at IS NULL
	`
	var org domain.Organization
	err := r.db.QueryRow(ctx, query, slug).Scan(
		&org.ID, &org.Name, &org.Slug, &org.Description, &org.LogoURL,
		&org.OwnerID, &org.CreatedAt, &org.UpdatedAt, &org.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *organizationRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.Organization, error) {
	query := `
		SELECT o.id, o.name, o.slug, o.description, o.logo_url, o.owner_id, o.created_at, o.updated_at, o.deleted_at
		FROM organizations o
		INNER JOIN organization_members om ON o.id = om.organization_id
		WHERE om.user_id = $1 AND o.deleted_at IS NULL
		ORDER BY o.name
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []*domain.Organization
	for rows.Next() {
		var org domain.Organization
		if err := rows.Scan(
			&org.ID, &org.Name, &org.Slug, &org.Description, &org.LogoURL,
			&org.OwnerID, &org.CreatedAt, &org.UpdatedAt, &org.DeletedAt,
		); err != nil {
			return nil, err
		}
		orgs = append(orgs, &org)
	}
	return orgs, rows.Err()
}

func (r *organizationRepository) Update(ctx context.Context, org *domain.Organization) error {
	org.UpdatedAt = time.Now()
	query := `
		UPDATE organizations SET name = $2, description = $3, logo_url = $4, updated_at = $5
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, org.ID, org.Name, org.Description, org.LogoURL, org.UpdatedAt)
	return err
}

func (r *organizationRepository) SoftDelete(ctx context.Context, id string) error {
	query := `UPDATE organizations SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *organizationRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM organizations WHERE slug = $1 AND deleted_at IS NULL)`
	var exists bool
	err := r.db.QueryRow(ctx, query, slug).Scan(&exists)
	return exists, err
}

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

func (r *organizationRepository) GenerateUniqueSlug(ctx context.Context, name string) (string, error) {
	base := strings.ToLower(name)
	base = nonAlphanumeric.ReplaceAllString(base, "-")
	base = strings.Trim(base, "-")
	if base == "" {
		base = "workspace"
	}
	if len(base) > 50 {
		base = base[:50]
	}

	slug := base
	for i := 1; ; i++ {
		exists, err := r.SlugExists(ctx, slug)
		if err != nil {
			return "", err
		}
		if !exists {
			return slug, nil
		}
		slug = base + "-" + cuid.New()[:8]
		if i > 10 {
			slug = cuid.New()
			break
		}
	}
	return slug, nil
}

func (r *organizationRepository) AddMember(ctx context.Context, member *domain.OrganizationMember) error {
	if member.ID == "" {
		member.ID = cuid.New()
	}
	member.JoinedAt = time.Now()

	query := `
		INSERT INTO organization_members (id, organization_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(ctx, query, member.ID, member.OrganizationID, member.UserID, member.Role, member.JoinedAt)
	return err
}

func (r *organizationRepository) FindMember(ctx context.Context, orgID, userID string) (*domain.OrganizationMember, error) {
	query := `
		SELECT om.id, om.organization_id, om.user_id, om.role, om.joined_at,
		       u.id, u.email, u.full_name, u.avatar_url
		FROM organization_members om
		INNER JOIN users u ON om.user_id = u.id
		WHERE om.organization_id = $1 AND om.user_id = $2
	`
	var member domain.OrganizationMember
	var user domain.User
	err := r.db.QueryRow(ctx, query, orgID, userID).Scan(
		&member.ID, &member.OrganizationID, &member.UserID, &member.Role, &member.JoinedAt,
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

func (r *organizationRepository) FindMemberBySlug(ctx context.Context, slug, userID string) (*domain.OrganizationMember, error) {
	query := `
		SELECT om.id, om.organization_id, om.user_id, om.role, om.joined_at,
		       o.id, o.name, o.slug, o.description, o.logo_url, o.owner_id, o.created_at, o.updated_at
		FROM organization_members om
		INNER JOIN organizations o ON om.organization_id = o.id
		WHERE o.slug = $1 AND om.user_id = $2 AND o.deleted_at IS NULL
	`
	var member domain.OrganizationMember
	var org domain.Organization
	err := r.db.QueryRow(ctx, query, slug, userID).Scan(
		&member.ID, &member.OrganizationID, &member.UserID, &member.Role, &member.JoinedAt,
		&org.ID, &org.Name, &org.Slug, &org.Description, &org.LogoURL, &org.OwnerID, &org.CreatedAt, &org.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	member.Organization = &org
	return &member, nil
}

func (r *organizationRepository) FindMembers(ctx context.Context, orgID string) ([]*domain.OrganizationMember, error) {
	query := `
		SELECT om.id, om.organization_id, om.user_id, om.role, om.joined_at,
		       u.id, u.email, u.full_name, u.avatar_url
		FROM organization_members om
		INNER JOIN users u ON om.user_id = u.id
		WHERE om.organization_id = $1
		ORDER BY om.role, om.joined_at
	`
	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*domain.OrganizationMember
	for rows.Next() {
		var member domain.OrganizationMember
		var user domain.User
		if err := rows.Scan(
			&member.ID, &member.OrganizationID, &member.UserID, &member.Role, &member.JoinedAt,
			&user.ID, &user.Email, &user.FullName, &user.AvatarURL,
		); err != nil {
			return nil, err
		}
		member.User = &user
		members = append(members, &member)
	}
	return members, rows.Err()
}

func (r *organizationRepository) UpdateMemberRole(ctx context.Context, orgID, userID string, role domain.OrgRole) error {
	query := `UPDATE organization_members SET role = $3 WHERE organization_id = $1 AND user_id = $2`
	_, err := r.db.Exec(ctx, query, orgID, userID, role)
	return err
}

func (r *organizationRepository) RemoveMember(ctx context.Context, orgID, userID string) error {
	query := `DELETE FROM organization_members WHERE organization_id = $1 AND user_id = $2`
	_, err := r.db.Exec(ctx, query, orgID, userID)
	return err
}

func (r *organizationRepository) CountMembers(ctx context.Context, orgID string) (int, error) {
	query := `SELECT COUNT(*) FROM organization_members WHERE organization_id = $1`
	var count int
	err := r.db.QueryRow(ctx, query, orgID).Scan(&count)
	return count, err
}

func (r *organizationRepository) CountBoards(ctx context.Context, orgID string) (int, error) {
	query := `SELECT COUNT(*) FROM boards WHERE organization_id = $1 AND deleted_at IS NULL`
	var count int
	err := r.db.QueryRow(ctx, query, orgID).Scan(&count)
	return count, err
}
