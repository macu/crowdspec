-- Clean up previous instance
DROP TABLE IF EXISTS org_permission;
DROP TYPE IF EXISTS org_permission_level;
DROP TABLE IF EXISTS organisation;
DROP TABLE IF EXISTS spec_permission;
DROP TYPE IF EXISTS spec_permission_level;
DROP TABLE IF EXISTS spec_url;
DROP TABLE IF EXISTS spec_block;
DROP TYPE IF EXISTS spec_block_ref_type;
DROP TABLE IF EXISTS spec_subspec;
DROP TABLE IF EXISTS spec;
DROP TYPE IF EXISTS spec_owner_type;
DROP TABLE IF EXISTS user_group_member;
DROP TABLE IF EXISTS user_group;
DROP TYPE IF EXISTS grant_type;
DROP TABLE IF EXISTS text_intern;
DROP TYPE IF EXISTS text_content_type;
DROP TYPE IF EXISTS list_style_type;
DROP TABLE IF EXISTS user_session;
DROP TABLE IF EXISTS user_account;
DROP COLLATION IF EXISTS case_insensitive;

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

-- Create spec tables
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
CREATE TABLE spec_subspec (
	id SERIAL PRIMARY KEY,
	spec_id INTEGER NOT NULL REFERENCES spec (id) ON DELETE CASCADE,
	created_at TIMESTAMPTZ NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL,
	subspec_name VARCHAR(255) NOT NULL COLLATE case_insensitive,
	subspec_desc TEXT,
	blocks_updated_at TIMESTAMPTZ,
	INDEX spec_subspec_name (spec_id, subspec_name)
);
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
	content_type text_content_type,
	ref_type spec_block_ref_type,
	ref_id INTEGER,
	block_title VARCHAR(255),
	block_body TEXT
);
CREATE TABLE spec_url (
	id SERIAL PRIMARY KEY,
	spec_id INTEGER NOT NULL REFERENCES spec (id) ON DELETE CASCADE,
	created_at TIMESTAMPTZ NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL,
	url VARCHAR(1024) NOT NULL,
	url_title VARCHAR(255) COLLATE case_insensitive,
	url_desc VARCHAR(255),
	url_image_data TEXT,
	INDEX spec_url_url (spec_id, url),
	INDEX spec_url_title (spec_id, url_title)
);

CREATE TABLE text_intern (
	id SERIAL PRIMARY KEY,
	string TEXT UNIQUE NOT NULL
);
