CREATE TABLE IF NOT EXISTS login_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NULL,
    message TEXT NOT NULL,

    log_type VARCHAR(10) NOT NULL CHECK (
        log_type IN ('warn', 'error', 'info', 'success')
    ),

    ip_address INET NULL,
    user_agent TEXT NULL,

    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_login_logs_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE SET NULL
);

-- Indexes for audit queries
CREATE INDEX IF NOT EXISTS idx_login_logs_user_id
    ON login_logs(user_id);

CREATE INDEX IF NOT EXISTS idx_login_logs_created_at
    ON login_logs(created_at);
