-- +goose Up
-- +goose StatementBegin
ALTER TABLE magic_link ADD COLUMN token TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE magic_link DROP COLUMN token;
-- +goose StatementEnd
