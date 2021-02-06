CREATE TABLE user_signup_request (
	id SERIAL PRIMARY KEY,
	username VARCHAR(25) UNIQUE NOT NULL COLLATE case_insensitive,
	email VARCHAR(50) UNIQUE NOT NULL,
	created_at TIMESTAMPTZ NOT NULL,
	reviewed BOOLEAN NOT NULL DEFAULT false,
	approved BOOLEAN NOT NULL DEFAULT false,
	token VARCHAR(15) UNIQUE,
	user_id INTEGER
);
