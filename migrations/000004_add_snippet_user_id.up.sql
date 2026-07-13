ALTER TABLE snippets ADD COLUMN user_id INTEGER REFERENCES users(id) ON DELETE SET NULL;
CREATE INDEX idx_snippets_user_id ON snippets(user_id);
