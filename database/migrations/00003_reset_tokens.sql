-- +goose Up
-- +goose StatementBegin
CREATE TABLE reset_tokens (
	id SERIAL PRIMARY KEY,
	user_id INTEGER UNIQUE REFERENCES users (id) ON DELETE CASCADE,
	token_hash TEXT UNIQUE NOT NULL,
	created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE reset_tokens;
-- +goose StatementEnd
