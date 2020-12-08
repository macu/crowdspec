CREATE TABLE password_reset_request (
	id SERIAL PRIMARY KEY,
	user_id INTEGER NOT NULL REFERENCES user_account (id) ON DELETE CASCADE,
	sent_to_address VARCHAR(50) NOT NULL,
	token VARCHAR(15) UNIQUE NOT NULL,
	created_at TIMESTAMPTZ,
	requested_at_ip_address VARCHAR(39) NOT NULL,
	fulfilled_at_ip_address VARCHAR(39)
);
