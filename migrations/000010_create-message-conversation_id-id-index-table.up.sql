CREATE INDEX idx_messages_conversation_id_id
ON messages (conversation_id, id DESC);