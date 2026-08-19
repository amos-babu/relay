CREATE TABLE IF NOT EXISTS messages (
    
    id BIGSERIAL PRIMARY KEY,

    conversation_id BIGINT NOT NULL,

    sender_id BIGINT NOT NULL,
    
    content TEXT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    FOREIGN KEY (conversation_id)
        REFERENCES conversations(id)
        ON DELETE CASCADE,
    
    FOREIGN KEY (sender_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

-- Index for sender lookups & fast foreign key CASCADE deletes
CREATE INDEX IF NOT EXISTS idx_messages_sender_id 
ON messages (sender_id);

