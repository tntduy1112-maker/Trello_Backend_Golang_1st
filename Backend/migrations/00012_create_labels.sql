-- +goose Up
-- +goose StatementBegin
CREATE TABLE labels (
    id VARCHAR(30) PRIMARY KEY,
    board_id VARCHAR(30) NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    name VARCHAR(100),
    color VARCHAR(7) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_labels_board_id ON labels(board_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS labels;
-- +goose StatementEnd
