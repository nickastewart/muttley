-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS location (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) NOT NULL
);  
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE location;
-- +goose StatementEnd
