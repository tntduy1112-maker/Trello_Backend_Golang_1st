-- +goose Up
-- +goose StatementBegin
CREATE TYPE card_priority AS ENUM ('none', 'low', 'medium', 'high');

CREATE TABLE cards (
    id VARCHAR(30) PRIMARY KEY,
    list_id VARCHAR(30) NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    position DOUBLE PRECISION NOT NULL DEFAULT 0,
    assignee_id VARCHAR(30) REFERENCES users(id) ON DELETE SET NULL,
    priority card_priority NOT NULL DEFAULT 'none',
    due_date TIMESTAMPTZ,
    is_completed BOOLEAN NOT NULL DEFAULT FALSE,
    completed_at TIMESTAMPTZ,
    cover_attachment_id VARCHAR(30),
    is_archived BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    archived_at TIMESTAMPTZ,
    created_by VARCHAR(30) NOT NULL REFERENCES users(id)
);

CREATE INDEX idx_cards_list_id ON cards(list_id);
CREATE INDEX idx_cards_list_position ON cards(list_id, position);
CREATE INDEX idx_cards_assignee_id ON cards(assignee_id);
CREATE INDEX idx_cards_due_date ON cards(due_date) WHERE due_date IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS cards;
DROP TYPE IF EXISTS card_priority;
-- +goose StatementEnd
