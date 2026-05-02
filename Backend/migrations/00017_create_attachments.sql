-- +goose Up
-- +goose StatementBegin
CREATE TABLE attachments (
    id              VARCHAR(30) PRIMARY KEY,
    card_id         VARCHAR(30) NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    uploaded_by     VARCHAR(30) NOT NULL REFERENCES users(id),
    filename        VARCHAR(255) NOT NULL,
    original_name   VARCHAR(255) NOT NULL,
    mime_type       VARCHAR(100) NOT NULL,
    file_size       BIGINT NOT NULL,
    object_key      VARCHAR(500) NOT NULL,
    url             TEXT NOT NULL,
    is_cover        BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_attachments_card ON attachments(card_id);
CREATE INDEX idx_attachments_uploader ON attachments(uploaded_by);
CREATE INDEX idx_attachments_cover ON attachments(card_id, is_cover) WHERE is_cover = TRUE;

-- Add FK to cards for cover_attachment_id
ALTER TABLE cards
    ADD CONSTRAINT fk_cards_cover_attachment
    FOREIGN KEY (cover_attachment_id) REFERENCES attachments(id) ON DELETE SET NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE cards DROP CONSTRAINT IF EXISTS fk_cards_cover_attachment;
DROP TABLE IF EXISTS attachments;
-- +goose StatementEnd
