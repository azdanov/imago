-- +goose Up
-- +goose StatementBegin
CREATE TABLE sessions (
	id SERIAL PRIMARY KEY,
	user_id INTEGER UNIQUE REFERENCES users (id) ON DELETE CASCADE,
	token_hash TEXT UNIQUE NOT NULL,
	created_at TIMESTAMPTZ DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE sessions;
-- +goose StatementEnd
