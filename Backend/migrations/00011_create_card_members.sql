-- +goose Up
-- +goose StatementBegin
CREATE TABLE card_members (
    id VARCHAR(30) PRIMARY KEY,
    card_id VARCHAR(30) NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    user_id VARCHAR(30) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(card_id, user_id)
);

CREATE INDEX idx_card_members_card_id ON card_members(card_id);
CREATE INDEX idx_card_members_user_id ON card_members(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS card_members;
-- +goose StatementEnd
