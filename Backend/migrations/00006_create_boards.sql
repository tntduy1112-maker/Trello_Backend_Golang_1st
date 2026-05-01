-- +goose Up
-- +goose StatementBegin
DO $$ BEGIN
    CREATE TYPE board_visibility AS ENUM ('private', 'workspace', 'public');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE TABLE boards (
    id                  VARCHAR(25) PRIMARY KEY,
    organization_id     VARCHAR(25) NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    title               VARCHAR(255) NOT NULL,
    description         TEXT,
    background_color    VARCHAR(7) DEFAULT '#0079bf',
    background_image    TEXT,

    visibility          board_visibility DEFAULT 'workspace',
    is_closed           BOOLEAN DEFAULT FALSE,

    owner_id            VARCHAR(25) NOT NULL REFERENCES users(id),

    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW(),
    closed_at           TIMESTAMPTZ,
    deleted_at          TIMESTAMPTZ
);

CREATE INDEX idx_boards_org ON boards(organization_id);
CREATE INDEX idx_boards_owner ON boards(owner_id);
CREATE INDEX idx_boards_visibility ON boards(visibility);
CREATE INDEX idx_boards_deleted ON boards(deleted_at) WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_boards_deleted;
DROP INDEX IF EXISTS idx_boards_visibility;
DROP INDEX IF EXISTS idx_boards_owner;
DROP INDEX IF EXISTS idx_boards_org;
DROP TABLE IF EXISTS boards;
DROP TYPE IF EXISTS board_visibility;
-- +goose StatementEnd
