-- +goose Up
-- +goose StatementBegin
CREATE TABLE email_verifications (
    id          VARCHAR(25) PRIMARY KEY,
    user_id     VARCHAR(25) NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    token       VARCHAR(64) NOT NULL,
    type        VARCHAR(20) NOT NULL,

    expires_at  TIMESTAMPTZ NOT NULL,
    used_at     TIMESTAMPTZ,

    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_email_verifications_user_type ON email_verifications(user_id, type);
CREATE INDEX idx_email_verifications_token ON email_verifications(token, type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_email_verifications_token;
DROP INDEX IF EXISTS idx_email_verifications_user_type;
DROP TABLE IF EXISTS email_verifications;
-- +goose StatementEnd
