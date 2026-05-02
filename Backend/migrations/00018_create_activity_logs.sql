-- +goose Up
-- +goose StatementBegin
CREATE TYPE activity_action AS ENUM (
    'card_created', 'card_updated', 'card_moved', 'card_archived', 'card_deleted',
    'card_assigned', 'card_unassigned', 'card_completed', 'card_reopened',
    'card_due_date_set', 'card_due_date_removed',
    'list_created', 'list_renamed', 'list_moved', 'list_archived',
    'label_added', 'label_removed',
    'checklist_created', 'checklist_deleted',
    'checklist_item_completed', 'checklist_item_uncompleted',
    'comment_added', 'comment_edited', 'comment_deleted',
    'attachment_added', 'attachment_deleted', 'cover_set', 'cover_removed',
    'member_added', 'member_removed'
);

CREATE TABLE activity_logs (
    id          VARCHAR(30) PRIMARY KEY,
    board_id    VARCHAR(30) NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    card_id     VARCHAR(30) REFERENCES cards(id) ON DELETE SET NULL,
    list_id     VARCHAR(30) REFERENCES lists(id) ON DELETE SET NULL,
    user_id     VARCHAR(30) NOT NULL REFERENCES users(id),
    action      activity_action NOT NULL,
    metadata    JSONB DEFAULT '{}',
    description TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_activity_logs_board ON activity_logs(board_id);
CREATE INDEX idx_activity_logs_card ON activity_logs(card_id);
CREATE INDEX idx_activity_logs_user ON activity_logs(user_id);
CREATE INDEX idx_activity_logs_created ON activity_logs(created_at DESC);
CREATE INDEX idx_activity_logs_metadata ON activity_logs USING GIN (metadata);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS activity_logs;
DROP TYPE IF EXISTS activity_action;
-- +goose StatementEnd
