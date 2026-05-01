-- +goose Up
-- +goose StatementBegin
DO $$ BEGIN
    CREATE TYPE org_role AS ENUM ('owner', 'admin', 'member');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE TABLE organization_members (
    id              VARCHAR(25) PRIMARY KEY,
    organization_id VARCHAR(25) NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id         VARCHAR(25) NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    role            org_role NOT NULL DEFAULT 'member',

    joined_at       TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT uniq_org_member UNIQUE (organization_id, user_id)
);

CREATE INDEX idx_org_members_org ON organization_members(organization_id);
CREATE INDEX idx_org_members_user ON organization_members(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_org_members_user;
DROP INDEX IF EXISTS idx_org_members_org;
DROP TABLE IF EXISTS organization_members;
DROP TYPE IF EXISTS org_role;
-- +goose StatementEnd
