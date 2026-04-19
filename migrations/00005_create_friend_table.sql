-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS friend (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    friend_id INTEGER NOT NULL,
    friend_status VARCHAR(25) NOT NULL,
    accepted_date TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT,
    row_version INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (user_id) REFERENCES user (id),
    FOREIGN KEY (friend_id) REFERENCES user (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE friend;
-- +goose StatementEnd
