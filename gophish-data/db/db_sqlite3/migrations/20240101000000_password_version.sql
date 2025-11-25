-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE users ADD COLUMN password_version INTEGER DEFAULT 0;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back




