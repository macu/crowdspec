ALTER TABLE password_reset_request DROP COLUMN requested_at_ip_address, DROP COLUMN fulfilled_at_ip_address;
ALTER TABLE password_reset_request ALTER COLUMN token DROP NOT NULL;
