-- +goose Up
-- +goose StatementBegin
ALTER TABLE posts ADD COLUMN hidden BOOLEAN NOT NULL DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE posts DROP COLUMN hidden;
-- +goose StatementEnd
