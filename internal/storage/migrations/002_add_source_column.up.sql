-- Optional: notification source (api, batch, scheduled)
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS source VARCHAR(50) DEFAULT 'api';
