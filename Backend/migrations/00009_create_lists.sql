-- +goose Up
-- +goose StatementBegin
CREATE TABLE lists (
    id VARCHAR(30) PRIMARY KEY,
    board_id VARCHAR(30) NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    position DOUBLE PRECISION NOT NULL DEFAULT 0,
    is_archived BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    archived_at TIMESTAMPTZ
);

CREATE INDEX idx_lists_board_id ON lists(board_id);
CREATE INDEX idx_lists_board_position ON lists(board_id, position);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS lists;
-- +goose StatementEnd
