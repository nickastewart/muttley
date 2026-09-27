-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS magic_link (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email varchar(255) NOT NULL,
    token_hash varchar(64) NOT NULL UNIQUE,
    purpose varchar(16) NOT NULL,
    expires_at TEXT NOT NULL,
    used_at TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_magic_link_email ON magic_link (email);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE magic_link;
-- +goose StatementEnd
