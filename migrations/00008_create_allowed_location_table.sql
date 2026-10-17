-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS allowed_location (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) NOT NULL UNIQUE
);
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO allowed_location (name) VALUES ('Daytona Milton Keynes'), ('Daytona Sandown Park');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE allowed_location;
-- +goose StatementEnd
