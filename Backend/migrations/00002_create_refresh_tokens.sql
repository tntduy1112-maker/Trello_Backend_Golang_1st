-- +goose Up
-- +goose StatementBegin
CREATE TABLE refresh_tokens (
    id              VARCHAR(25) PRIMARY KEY,
    user_id         VARCHAR(25) NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    token_hash      VARCHAR(64) NOT NULL,

    device_info     VARCHAR(255),
    ip_address      VARCHAR(45),

    is_revoked      BOOLEAN DEFAULT FALSE,
    expires_at      TIMESTAMPTZ NOT NULL,

    created_at      TIMESTAMPTZ DEFAULT NOW(),
    revoked_at      TIMESTAMPTZ,

    CONSTRAINT uniq_refresh_tokens_hash UNIQUE (token_hash)
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_expires ON refresh_tokens(expires_at) WHERE is_revoked = FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_refresh_tokens_expires;
DROP INDEX IF EXISTS idx_refresh_tokens_hash;
DROP INDEX IF EXISTS idx_refresh_tokens_user_id;
DROP TABLE IF EXISTS refresh_tokens;
-- +goose StatementEnd
