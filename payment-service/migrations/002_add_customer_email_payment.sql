ALTER TABLE payments
    ADD COLUMN IF NOT EXISTS customer_email TEXT;