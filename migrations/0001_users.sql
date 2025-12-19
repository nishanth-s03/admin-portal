-- Enable UUID generation (PostgreSQL)
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    username VARCHAR(150) NOT NULL UNIQUE,

    role VARCHAR(20) NOT NULL CHECK (
        role IN ('user', 'admin', 'super-admin')
    ),

    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_activated BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

-- Useful index for auth lookups
CREATE INDEX IF NOT EXISTS idx_users_username
    ON users(username);

CREATE INDEX IF NOT EXISTS idx_users_id
    ON users(username);