-- block community features

/*
tag comments[, blocks[, subspecs, specs]]

## completely open tagging
allows the community to take control of the culture of the community space

## allowed tags
admin-supplied tags could have translations
allows restricting comment classes;
may be more professional

targeted tags allows individuals and organisations to emphasize their own focus
while allowing open community tagging

*/

/* ALTER TABLE spec ADD COLUMN resticted_tagging BOOLEAN NOT NULL DEFAULT 0;
ALTER TABLE spec_subspec ADD COLUMN restricted_tagging BOOLEAN; -- null indicates inherit from spec
ALTER TABLE spec_block ADD COLUMN restricted_tagging BOOLEAN; -- null indicates inherit from spec or subspec */

CREATE TABLE user_tag (
	id SERIAL PRIMARY KEY,
	created_at TIMESTAMPTZ NOT NULL,
	tag_label VARCHAR(50) NOT NULL UNIQUE -- e.g. "correct", "priority"
);

/* CREATE TABLE spec_allowed_tag (
	id SERIAL PRIMARY KEY,
	spec_id INTEGER NOT NULL REFERENCES spec (id) ON DELETE CASCADE,
	subspec_id INTEGER REFERENCES spec_subspec (id) ON DELETE CASCADE,
	tag_id INTEGER NOT NULL REFERENCES user_tag (id) ON DELETE CASCADE,
	CONSTRAINT single_assoc UNIQUE (spec_id, subspec_id, tag_id)
); */

CREATE TYPE spec_comment_ref_type AS ENUM ('spec', 'subspec', 'block');

CREATE TABLE spec_comment (
	id SERIAL PRIMARY KEY,
	spec_id INTEGER NOT NULL REFERENCES spec (id) ON DELETE CASCADE,
	user_id INTEGER NOT NULL,
	created_at TIMESTAMPTZ NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL,

	-- comment associated with
	ref_type spec_comment_ref_type NOT NULL,
	ref_id INTEGER NOT NULL,

	body TEXT
);

-- tags anticipated by the admin
CREATE TABLE spec_block_target_tag (
	id SERIAL PRIMARY KEY,
	spec_id INTEGER NOT NULL REFERENCES spec (id) ON DELETE CASCADE,



	subspec_id INTEGER REFERENCES spec_subspec (id) ON DELETE CASCADE,
	block_id INTEGER REFERENCES spec_block (id) ON DELETE CASCADE,
	tag_id INTEGER NOT NULL REFERENCES user_tag (id) ON DELETE CASCADE,
	CONSTRAINT single_assoc UNIQUE (spec_id, subspec_id, block_id, tag_id)
);

CREATE TYPE user_tag_vote_type AS ENUM ('assent', 'dissent');

CREATE TYPE spec_tag_ref_type AS ENUM ('block', 'comment');

CREATE TABLE spec_tag (
	id SERIAL PRIMARY KEY,

	-- fixed
	spec_id INTEGER NOT NULL REFERENCES spec (id) ON DELETE CASCADE,
	ref_type spec_tag_ref_type NOT NULL,
	ref_id INTEGER NOT NULL,
	tag_id INTEGER NOT NULL REFERENCES user_tag (id) ON DELETE CASCADE,

	-- dynamic
	admin_pin BOOLEAN NOT NULL DEFAULT FALSE,
	author_vote user_tag_vote_type,
	pinned BOOLEAN GENERATED ALWAYS AS (
		admin_pin OR author_vote = 'assent'
	) STORED,
	vote_sum INTEGER NOT NULL DEFAULT 0,

	CONSTRAINT pinned_tags UNIQUE (pinned, spec_id, subspec_id, block_id, comment_id, tag_id), -- index

	CONSTRAINT single_target CHECK (
		array_length(array_remove(ARRAY[subspec_id, block_id, comment_id], NULL), 1) <= 1
	)
);

-- user vote on tag on comment
CREATE TABLE spec_tag_vote (
	id SERIAL PRIMARY KEY,
	spec_tag_id INTEGER NOT NULL REFERENCES spec_tag (id) ON DELETE CASCADE,
	user_id INTEGER NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL,
	vote user_tag_vote_type NOT NULL,
	CONSTRAINT single_vote UNIQUE (spec_tag_id, user_id) -- want as index if determining author vote from join
);


-- create triggers to delete community records associated through ref_type and ref_id

CREATE FUNCTION delete_subspec_community () RETURNS TRIGGER AS $$
	BEGIN
		DELETE FROM spec_comment WHERE spec_id = OLD.spec_id AND ref_type = 'subspec' AND ref_id = OLD.id;
	END;
$$ LANGUAGE PLPGSQL;

CREATE TRIGGER on_subspec_delete AFTER DELETE ON spec_subspec
FOR EACH ROW EXECUTE PROCEDURE delete_subspec_community();


CREATE FUNCTION delete_block_community () RETURNS TRIGGER AS $$
	BEGIN
		DELETE FROM spec_comment WHERE spec_id = OLD.spec_id AND ref_type = 'block' AND ref_id = OLD.id;
		DELETE FROM spec_tag WHERE spec_id = OLD.spec_id AND ref_type = 'block' AND ref_id = OLD.id;
	END;
$$ LANGUAGE PLPGSQL;

CREATE TRIGGER on_block_delete AFTER DELETE ON spec_block
FOR EACH ROW EXECUTE PROCEDURE delete_block_community();


CREATE FUNCTION delete_comment_tags () RETURNS TRIGGER AS $$
	BEGIN
		DELETE FROM spec_tag WHERE spec_id = OLD.spec_id AND ref_type = 'comment' AND ref_id = OLD.id;
	END;
$$ LANGUAGE PLPGSQL;

CREATE TRIGGER on_comment_delete AFTER DELETE ON spec_comment
FOR EACH ROW EXECUTE PROCEDURE delete_comment_tags();
