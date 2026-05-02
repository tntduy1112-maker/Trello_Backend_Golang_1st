-- +goose Up
-- +goose StatementBegin
CREATE TABLE checklist_items (
    id              VARCHAR(30) PRIMARY KEY,
    checklist_id    VARCHAR(30) NOT NULL REFERENCES checklists(id) ON DELETE CASCADE,
    title           VARCHAR(500) NOT NULL,
    position        DOUBLE PRECISION NOT NULL,
    is_completed    BOOLEAN DEFAULT FALSE,
    completed_at    TIMESTAMPTZ,
    completed_by    VARCHAR(30) REFERENCES users(id),
    assignee_id     VARCHAR(30) REFERENCES users(id) ON DELETE SET NULL,
    due_date        TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_checklist_items_checklist ON checklist_items(checklist_id);
CREATE INDEX idx_checklist_items_position ON checklist_items(checklist_id, position);
CREATE INDEX idx_checklist_items_assignee ON checklist_items(assignee_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS checklist_items;
-- +goose StatementEnd
