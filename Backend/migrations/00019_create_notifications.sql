-- +goose Up
CREATE TYPE notification_type AS ENUM (
    'card_assigned',
    'card_due_soon',
    'card_overdue',
    'comment_added',
    'comment_reply',
    'mentioned',
    'board_invitation',
    'checklist_item_assigned',
    'card_completed',
    'member_added_to_board'
);

CREATE TABLE notifications (
    id              VARCHAR(25) PRIMARY KEY,
    user_id         VARCHAR(25) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type            notification_type NOT NULL,
    title           VARCHAR(255) NOT NULL,
    message         TEXT,
    board_id        VARCHAR(25) REFERENCES boards(id) ON DELETE CASCADE,
    card_id         VARCHAR(25) REFERENCES cards(id) ON DELETE CASCADE,
    actor_id        VARCHAR(25) REFERENCES users(id) ON DELETE SET NULL,
    is_read         BOOLEAN DEFAULT FALSE,
    read_at         TIMESTAMPTZ,
    metadata        JSONB DEFAULT '{}',
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_notifications_user ON notifications(user_id);
CREATE INDEX idx_notifications_unread ON notifications(user_id, is_read) WHERE is_read = FALSE;
CREATE INDEX idx_notifications_created ON notifications(created_at DESC);
CREATE INDEX idx_notifications_type ON notifications(type);
CREATE INDEX idx_notifications_board ON notifications(board_id);
CREATE INDEX idx_notifications_card ON notifications(card_id);

-- +goose Down
DROP TABLE IF EXISTS notifications;
DROP TYPE IF EXISTS notification_type;
