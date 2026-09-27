-- +goose Up
-- +goose StatementBegin
ALTER TABLE user DROP COLUMN password;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE user ADD COLUMN password varchar(50) NOT NULL DEFAULT '';
-- +goose StatementEnd
