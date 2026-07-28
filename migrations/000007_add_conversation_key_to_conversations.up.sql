ALTER TABLE conversations
ADD COLUMN IF NOT EXISTS conversation_key TEXT UNIQUE;

