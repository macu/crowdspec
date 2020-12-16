DROP TABLE IF EXISTS spec_community_comment_tag_vote;
DROP TABLE IF EXISTS spec_community_comment_tag;
DROP TABLE IF EXISTS spec_community_comment;
DROP TABLE IF EXISTS tag_tree_tag;
DROP TABLE IF EXISTS tag_tree_category;
DROP TABLE IF EXISTS tag_tree;
DROP TABLE IF EXISTS spec_community_config;
DROP TABLE IF EXISTS tag;

CREATE TABLE tag (
	id SERIAL PRIMARY KEY,
	tag_name VARCHAR(50) NOT NULL
);

CREATE TABLE spec_community_config (
	id SERIAL PRIMARY KEY,
	spec_id INTEGER NOT NULL REFERENCES spec (id) ON DELETE CASCADE,
	subspec_id INTEGER REFERENCES spec_subspec (id) ON DELETE CASCADE, -- defined if targeting subspec layer
	block_id INTEGER REFERENCES spec_block (id) ON DELETE CASCADE, -- defined if targeting block layer
	open_tagging BOOLEAN, -- whether to enable arbitrary tagging by the public (inherited)
	self_tag_tree INTEGER REFERENCES tag_tree (id), -- defined if extending or defining tags for self (inherited)
	sub_tag_tree INTEGER REFERENCES tag_tree (id), -- defined if extending or defining tags for immediate sub items (inherited, overrides self_tag_tree)
	admin_note TEXT,
);

CREATE TABLE tag_tree (
	id SERIAL PRIMARY KEY,
	owner_type spec_owner_type NOT NULL,
	owner_id INTEGER NOT NULL,
	tag_tree_name VARCHAR(255) NOT NULL,
	parent_tree_id
);

CREATE TABLE tag_tree_category (
	id SERIAL PRIMARY KEY,
	tag_tree_id INTEGER NOT NULL REFERENCES tag_tree (id) ON DELETE CASCADE,
	parent_category_id INTEGER REFERENCES tag_tree_category (id) ON DELETE CASCADE,
	category_name VARCHAR(50) NOT NULL,
	order_number INTEGER -- order by order_number or alphabetical
);

CREATE TABLE tag_tree_tag (
	tag_tree_id INTEGER NOT NULL REFERENCES tag_tree (id) ON DELETE CASCADE,
	category_id INTEGER REFERENCES tag_tree_category (id) ON DELETE CASCADE,
	tag_id INTEGER NOT NULL REFERENCES tag (id) ON DELETE CASCADE,
	order_number INTEGER, -- order by order_number or alphabetical
	PRIMARY KEY (category_id, tag_id)
);

CREATE TABLE spec_community_comment (
	id SERIAL PRIMARY KEY,
	spec_id INTEGER NOT NULL REFERENCES spec (id) ON DELETE CASCADE,
	subspec_id INTEGER REFERENCES spec_subspec (id) ON DELETE CASCADE, -- defined if targeting subspec layer
	block_id INTEGER REFERENCES spec_block (id) ON DELETE CASCADE, -- defined if targeting block layer
	user_id INTEGER NOT NULL REFERENCES user_account (id),
	created_at TIMESTAMPTZ NOT NULL,
	comment_body TEXT NOT NULL
);

CREATE TABLE spec_community_comment_tag (
	comment_id INTEGER NOT NULL REFERENCES spec_community_comment (id) ON DELETE CASCADE,
	tag_id INTEGER NOT NULL REFERENCES tag (id) ON DELETE CASCADE,
	added_by_user_id INTEGER NOT NULL REFERENCES user_account (id),
	added_at TIMESTAMPTZ NOT NULL,
	vote_sum INTEGER NOT NULL DEFAULT 0,
	admin_pin BOOLEAN NOT NULL DEFAULT false,
	PRIMARY KEY (comment_id, tag_id)
);

CREATE TABLE spec_community_comment_tag_vote (
	comment_id INTEGER NOT NULL REFERENCES spec_community_comment (id) ON DELETE CASCADE,
	tag_id INTEGER NOT NULL REFERENCES tag (id) ON DELETE CASCADE,
	user_id INTEGER NOT NULL REFERENCES user_account (id),
	updated_at TIMESTAMPTZ NOT NULL,
	vote ENUM('assent', 'dissent') NOT NULL,
	PRIMARY KEY (comment_id, tag_id)
);
