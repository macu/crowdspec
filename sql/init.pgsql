-- Clean up previous instance

DROP TRIGGER IF EXISTS on_spec_delete ON spec;
DROP TRIGGER IF EXISTS on_subspec_delete ON spec_subspec;
DROP TRIGGER IF EXISTS on_block_delete ON spec_block;
DROP TRIGGER IF EXISTS on_comment_delete ON spec_community_comment;
DROP FUNCTION IF EXISTS delete_community_space;

DROP TABLE IF EXISTS spec_community_read;

DROP INDEX IF EXISTS comment_updated_by_target;
DROP INDEX IF EXISTS comment_by_spec;
DROP INDEX IF EXISTS comment_updated_by_user;
DROP TABLE IF EXISTS spec_community_comment;

DROP TYPE IF EXISTS spec_community_target_type;

DROP INDEX IF EXISTS spec_url_by_url;
DROP INDEX IF EXISTS spec_url_by_title;
DROP TABLE IF EXISTS spec_url;

DROP TABLE IF EXISTS spec_block;
DROP TYPE IF EXISTS text_content_type;
DROP TYPE IF EXISTS list_style_type;
DROP TYPE IF EXISTS spec_block_ref_type;

DROP INDEX IF EXISTS spec_subspec_by_name;
DROP TABLE IF EXISTS spec_subspec;

DROP TABLE IF EXISTS spec;
DROP TYPE IF EXISTS spec_owner_type;

DROP TABLE IF EXISTS user_signup_request;
DROP TABLE IF EXISTS password_reset_request;
DROP TABLE IF EXISTS user_session;
DROP TABLE IF EXISTS user_account;

DROP COLLATION IF EXISTS case_insensitive;

-- Create collation for sorting

CREATE COLLATION case_insensitive (
	provider = icu, -- "International Components for Unicode"
	-- und stands for undefined (ICU root collation - language agnostic)
	-- colStrength=primary ignores case and accents
	-- colNumeric=yes sorts strings with numeric parts by numeric value
	-- colAlternate=shifted would recognize equality of equivalent punctuation sequences
	locale = 'und@colStrength=primary;colNumeric=yes',
	deterministic = false
);

-- Create minimal tables for user authentication and session management

CREATE TABLE user_account (
	id SERIAL PRIMARY KEY,
	username VARCHAR(25) UNIQUE NOT NULL COLLATE case_insensitive,
	email VARCHAR(50) UNIQUE NOT NULL,
	auth_hash VARCHAR(60) NOT NULL,
	user_settings JSON,
	created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE user_session (
	token VARCHAR(30) PRIMARY KEY,
	user_id INTEGER NOT NULL REFERENCES user_account (id) ON DELETE CASCADE,
	expires TIMESTAMPTZ NOT NULL
);

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

CREATE TABLE password_reset_request (
	id SERIAL PRIMARY KEY,
	user_id INTEGER NOT NULL REFERENCES user_account (id) ON DELETE CASCADE,
	sent_to_address VARCHAR(50) NOT NULL,
	token VARCHAR(15) UNIQUE,
	created_at TIMESTAMPTZ
);

-- Create spec content tables

CREATE TYPE spec_owner_type AS ENUM (
	'user',
	'org'
);

CREATE TABLE spec (
	id SERIAL PRIMARY KEY,
	owner_type spec_owner_type NOT NULL,
	owner_id INTEGER NOT NULL,
	created_at TIMESTAMPTZ NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL,
	spec_name VARCHAR(255) NOT NULL,
	spec_desc TEXT,
	is_public BOOLEAN NOT NULL DEFAULT false,
	blocks_updated_at TIMESTAMPTZ
);

CREATE TABLE spec_subspec (
	id SERIAL PRIMARY KEY,
	spec_id INTEGER NOT NULL REFERENCES spec (id) ON DELETE CASCADE,
	created_at TIMESTAMPTZ NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL,
	subspec_name VARCHAR(255) NOT NULL COLLATE case_insensitive,
	subspec_desc TEXT,
	blocks_updated_at TIMESTAMPTZ,
	is_private BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE INDEX spec_subspec_by_name ON spec_subspec (spec_id, is_private, subspec_name);

CREATE TYPE list_style_type AS ENUM (
	'bullet',
	'numbered',
	'none'
);

CREATE TYPE text_content_type AS ENUM (
	'plaintext',
	'markdown',
	'html'
);

CREATE TYPE spec_block_ref_type AS ENUM (
	'org',
	'spec',
	'subspec',
	'block',
	'image',
	'video',
	'url',
	'file'
);

CREATE TABLE spec_block (
	id SERIAL PRIMARY KEY,
	spec_id INTEGER NOT NULL REFERENCES spec (id) ON DELETE CASCADE,
	created_at TIMESTAMPTZ NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL,
	subspec_id INTEGER REFERENCES spec_subspec (id) ON DELETE CASCADE,
	parent_id INTEGER REFERENCES spec_block (id) ON DELETE CASCADE,
	order_number INTEGER NOT NULL,
	style_type list_style_type NOT NULL DEFAULT 'none',
	content_type text_content_type NOT NULL,
	ref_type spec_block_ref_type,
	ref_id INTEGER,
	block_title VARCHAR(255),
	block_body TEXT,
	rendered_html TEXT
);

CREATE TABLE spec_url (
	id SERIAL PRIMARY KEY,
	spec_id INTEGER NOT NULL REFERENCES spec (id) ON DELETE CASCADE,
	created_at TIMESTAMPTZ NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL,
	url VARCHAR(1024) NOT NULL,
	url_title VARCHAR(255) COLLATE case_insensitive,
	url_desc VARCHAR(255),
	url_image_data TEXT
);
CREATE INDEX spec_url_by_url ON spec_url (spec_id, url);
CREATE INDEX spec_url_by_title ON spec_url (spec_id, url_title);

-- Create community tables

CREATE TYPE spec_community_target_type AS ENUM (
	'spec',
	'subspec',
	'block',
	'comment'
);

CREATE TABLE spec_community_comment (
	id SERIAL PRIMARY KEY,
	spec_id INTEGER NOT NULL REFERENCES spec (id) ON DELETE CASCADE,
	target_type spec_community_target_type NOT NULL,
	target_id INTEGER NOT NULL,
	user_id INTEGER NOT NULL REFERENCES user_account (id),
	created_at TIMESTAMPTZ NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL,
	comment_body TEXT NOT NULL
);
CREATE INDEX comment_updated_by_target ON spec_community_comment (target_type, target_id, updated_at);
CREATE INDEX comment_by_spec ON spec_community_comment (spec_id);
CREATE INDEX comment_updated_by_user ON spec_community_comment (user_id, updated_at);

CREATE TABLE spec_community_read (
	user_id INTEGER NOT NULL REFERENCES user_account (id),
	target_type spec_community_target_type NOT NULL,
	target_id INTEGER NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL,
	/* content_hidden BOOLEAN NOT NULL DEFAULT false, */
	PRIMARY KEY (user_id, target_type, target_id)
);

-- Create triggers to delete community records associated through target_type and target_id

CREATE FUNCTION delete_community_space ()
RETURNS TRIGGER
LANGUAGE PLPGSQL
AS $$
DECLARE
	targetType spec_community_target_type;
BEGIN
	targetType := TG_ARGV[0];
	DELETE FROM spec_community_comment
		WHERE target_type = targetType AND target_id = OLD.id;
	DELETE FROM spec_community_read
		WHERE target_type = targetType AND target_id = OLD.id;
	RETURN NULL;
END $$;

CREATE TRIGGER on_spec_delete AFTER DELETE ON spec
	FOR EACH ROW EXECUTE PROCEDURE delete_community_space('spec');

CREATE TRIGGER on_subspec_delete AFTER DELETE ON spec_subspec
	FOR EACH ROW EXECUTE PROCEDURE delete_community_space('subspec');

CREATE TRIGGER on_block_delete AFTER DELETE ON spec_block
	FOR EACH ROW EXECUTE PROCEDURE delete_community_space('block');

CREATE TRIGGER on_comment_delete AFTER DELETE ON spec_community_comment
	FOR EACH ROW EXECUTE PROCEDURE delete_community_space('comment');

/*
-- Create organisation and user group tables

CREATE TABLE user_group (
	id SERIAL PRIMARY KEY,
	group_name VARCHAR(50),
	created_at TIMESTAMPTZ NOT NULL
);
CREATE TABLE user_group_member (
	group_id INTEGER NOT NULL REFERENCES user_group (id) ON DELETE CASCADE,
	user_id INTEGER NOT NULL REFERENCES user_account (id) ON DELETE CASCADE,
	PRIMARY KEY (group_id, user_id)
);

CREATE TYPE grant_type AS ENUM (
	'user', -- grant applies to user_account
	'group' -- grant applies to user_group
);

CREATE TABLE organisation (
	id SERIAL PRIMARY KEY,
	org_handle VARCHAR(50) UNIQUE, -- allows nulls
	org_name VARCHAR(120) NOT NULL,
	org_intro TEXT
);
CREATE TYPE org_permission_level AS ENUM (
	'owner', -- user or group owns org
	'admin', -- user or group can administrate org
	'internal', -- user or group are internal (can manage specs)
	'external' -- user or group are external (can view and contribute to listed org specs)
);
CREATE TABLE org_permission (
	org_id INTEGER NOT NULL REFERENCES organisation (id) ON DELETE CASCADE,
	grant_type grant_type NOT NULL,
	grant_id INTEGER NOT NULL,
	permission_level org_permission_level NOT NULL,
	CONSTRAINT user_owns CHECK (grant_type = 'user' OR permission_level != 'owner')
);
CREATE TYPE spec_permission_level AS ENUM (
	'admin', -- user or group can administrate spec
	'editor', -- user or group can manage spec
	'contributor' -- user or group can view and contribute to spec
);
CREATE TABLE spec_permission (
	spec_id INTEGER NOT NULL REFERENCES spec (id) ON DELETE CASCADE,
	grant_type grant_type NOT NULL,
	grant_id INTEGER NOT NULL, -- referrent of type grant_type
	permission_level spec_permission_level NOT NULL
	-- TODO add index
);


CREATE TABLE text_intern (
	id SERIAL PRIMARY KEY,
	string TEXT UNIQUE NOT NULL
);
*/
