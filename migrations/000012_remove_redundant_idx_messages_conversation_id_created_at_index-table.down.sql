CREATE INDEX idx_messages_conversation_id_created_at
ON messages (conversation_id, created_at DESC);