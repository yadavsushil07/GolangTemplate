-- Add phone and email columns to users for multi-channel OTP
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS phone TEXT,
    ADD COLUMN IF NOT EXISTS email TEXT;

-- Unique indexes (sparse — only enforced when value is non-null)
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_phone ON users(phone) WHERE phone IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email) WHERE email IS NOT NULL;

-- Vendor config: stores vendor contact details for notifications
CREATE TABLE IF NOT EXISTS vendor_config (
    id           SERIAL PRIMARY KEY,
    vendor_name  TEXT NOT NULL DEFAULT 'SBY TWILIGHT',
    phone        TEXT,
    email        TEXT,
    whatsapp     TEXT,
    updated_at   TIMESTAMPTZ DEFAULT NOW()
);
