-- +goose Up
-- +goose StatementBegin
DO $$ BEGIN
    CREATE TYPE board_role AS ENUM ('owner', 'admin', 'member', 'viewer');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE TABLE board_members (
    id          VARCHAR(25) PRIMARY KEY,
    board_id    VARCHAR(25) NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    user_id     VARCHAR(25) NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    role        board_role NOT NULL DEFAULT 'member',

    joined_at   TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT uniq_board_member UNIQUE (board_id, user_id)
);

CREATE INDEX idx_board_members_board ON board_members(board_id);
CREATE INDEX idx_board_members_user ON board_members(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_board_members_user;
DROP INDEX IF EXISTS idx_board_members_board;
DROP TABLE IF EXISTS board_members;
DROP TYPE IF EXISTS board_role;
-- +goose StatementEnd
