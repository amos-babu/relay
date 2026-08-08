CREATE TABLE IF NOT EXISTS message_reads (
    
    message_id BIGINT NOT NULL,

    user_id BIGINT NOT NULL,

    read_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (message_id, user_id),
    
    CONSTRAINT fk_message_reads_message
        FOREIGN KEY (message_id)
        REFERENCES messages(id)
        ON DELETE CASCADE,
    
    CONSTRAINT fk_message_reads_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_message_reads_user_id
    ON message_reads(user_id);

