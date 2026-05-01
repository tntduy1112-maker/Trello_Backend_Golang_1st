-- +goose Up
-- +goose StatementBegin
DO $$ BEGIN
    CREATE TYPE invitation_status AS ENUM ('pending', 'accepted', 'declined', 'expired');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE TABLE board_invitations (
    id              VARCHAR(25) PRIMARY KEY,
    board_id        VARCHAR(25) NOT NULL REFERENCES boards(id) ON DELETE CASCADE,

    inviter_id      VARCHAR(25) NOT NULL REFERENCES users(id),

    invitee_id      VARCHAR(25) REFERENCES users(id),
    invitee_email   VARCHAR(255) NOT NULL,

    role            board_role NOT NULL DEFAULT 'member',
    token           VARCHAR(64) NOT NULL,
    message         TEXT,

    status          invitation_status DEFAULT 'pending',
    expires_at      TIMESTAMPTZ NOT NULL,

    created_at      TIMESTAMPTZ DEFAULT NOW(),
    responded_at    TIMESTAMPTZ
);

CREATE INDEX idx_board_invitations_board ON board_invitations(board_id);
CREATE INDEX idx_board_invitations_email ON board_invitations(invitee_email);
CREATE INDEX idx_board_invitations_token ON board_invitations(token);
CREATE INDEX idx_board_invitations_status ON board_invitations(status) WHERE status = 'pending';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_board_invitations_status;
DROP INDEX IF EXISTS idx_board_invitations_token;
DROP INDEX IF EXISTS idx_board_invitations_email;
DROP INDEX IF EXISTS idx_board_invitations_board;
DROP TABLE IF EXISTS board_invitations;
DROP TYPE IF EXISTS invitation_status;
-- +goose StatementEnd
