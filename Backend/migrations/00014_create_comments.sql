-- +goose Up
-- +goose StatementBegin
CREATE TABLE comments (
    id          VARCHAR(30) PRIMARY KEY,
    card_id     VARCHAR(30) NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    author_id   VARCHAR(30) NOT NULL REFERENCES users(id),
    content     TEXT NOT NULL,
    parent_id   VARCHAR(30) REFERENCES comments(id) ON DELETE CASCADE,
    is_edited   BOOLEAN DEFAULT FALSE,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_comments_card ON comments(card_id);
CREATE INDEX idx_comments_author ON comments(author_id);
CREATE INDEX idx_comments_parent ON comments(parent_id);
CREATE INDEX idx_comments_deleted ON comments(deleted_at) WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS comments;
-- +goose StatementEnd
