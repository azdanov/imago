-- +goose Up
-- +goose StatementBegin
CREATE TABLE galleries (
	id SERIAL PRIMARY KEY,
	user_id INTEGER UNIQUE REFERENCES users (id) ON DELETE CASCADE,
	title TEXT,
	created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE galleries;
-- +goose StatementEnd
