-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    first_name varchar(50) NOT NULL,
    last_name varchar(50) NOT NULL,
    email varchar(255) NOT NULL,
    profile_id varchar(50) NOT NULL,
    password varchar(50) NOT NULL,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP 
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE user;
-- +goose StatementEnd
