-- +goose Up
-- +goose StatementBegin
CREATE TABLE organizations (
    id              VARCHAR(25) PRIMARY KEY,
    name            VARCHAR(255) NOT NULL,
    slug            VARCHAR(100) NOT NULL,
    description     TEXT,
    logo_url        TEXT,

    owner_id        VARCHAR(25) NOT NULL REFERENCES users(id),

    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uniq_organizations_slug UNIQUE (slug)
);

CREATE INDEX idx_organizations_slug ON organizations(slug);
CREATE INDEX idx_organizations_owner ON organizations(owner_id);
CREATE INDEX idx_organizations_deleted ON organizations(deleted_at) WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_organizations_deleted;
DROP INDEX IF EXISTS idx_organizations_owner;
DROP INDEX IF EXISTS idx_organizations_slug;
DROP TABLE IF EXISTS organizations;
-- +goose StatementEnd
