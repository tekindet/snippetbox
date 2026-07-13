DROP INDEX IF EXISTS idx_snippets_user_id;
ALTER TABLE snippets DROP COLUMN IF EXISTS user_id;
