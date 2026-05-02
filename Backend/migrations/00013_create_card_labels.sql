-- +goose Up
-- +goose StatementBegin
CREATE TABLE card_labels (
    id VARCHAR(30) PRIMARY KEY,
    card_id VARCHAR(30) NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    label_id VARCHAR(30) NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(card_id, label_id)
);

CREATE INDEX idx_card_labels_card_id ON card_labels(card_id);
CREATE INDEX idx_card_labels_label_id ON card_labels(label_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS card_labels;
-- +goose StatementEnd
