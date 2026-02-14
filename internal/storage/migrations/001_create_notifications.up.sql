CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS notifications (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    batch_id        UUID,
    recipient       VARCHAR(255) NOT NULL,
    channel         VARCHAR(50) NOT NULL CHECK (channel IN ('sms', 'email', 'push')),
    content         TEXT NOT NULL,
    priority        VARCHAR(20) NOT NULL DEFAULT 'normal' CHECK (priority IN ('high', 'normal', 'low')),
    status          VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'delivered', 'failed', 'cancelled')),
    idempotency_key VARCHAR(255) UNIQUE,
    provider_msg_id VARCHAR(255),
    retry_count     INT NOT NULL DEFAULT 0,
    max_retries     INT NOT NULL DEFAULT 3,
    scheduled_at    TIMESTAMP WITH TIME ZONE,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    error_message   TEXT
);

CREATE INDEX IF NOT EXISTS idx_notifications_batch_id ON notifications(batch_id);
CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications(status);
CREATE INDEX IF NOT EXISTS idx_notifications_channel ON notifications(channel);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at);
CREATE INDEX IF NOT EXISTS idx_notifications_idempotency ON notifications(idempotency_key) WHERE idempotency_key IS NOT NULL;
