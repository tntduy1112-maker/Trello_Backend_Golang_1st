-- +goose Up
-- +goose StatementBegin
CREATE TABLE checklists (
    id          VARCHAR(30) PRIMARY KEY,
    card_id     VARCHAR(30) NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    title       VARCHAR(255) NOT NULL,
    position    DOUBLE PRECISION NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_checklists_card ON checklists(card_id);
CREATE INDEX idx_checklists_position ON checklists(card_id, position);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS checklists;
-- +goose StatementEnd
