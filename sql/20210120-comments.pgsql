DROP TRIGGER IF EXISTS on_spec_delete ON spec;
DROP TRIGGER IF EXISTS on_subspec_delete ON spec_subspec;
DROP TRIGGER IF EXISTS on_block_delete ON spec_block;
DROP TRIGGER IF EXISTS on_comment_delete ON spec_community_comment;
DROP FUNCTION IF EXISTS delete_community_space;
DROP TABLE IF EXISTS spec_community_read;
DROP TABLE IF EXISTS spec_community_comment;
DROP TYPE IF EXISTS spec_community_target_type;

-- add community comments

CREATE TYPE spec_community_target_type AS ENUM ('spec', 'subspec', 'block', 'comment');

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

CREATE TABLE spec_community_read (
	user_id INTEGER NOT NULL REFERENCES user_account (id),
	target_type spec_community_target_type NOT NULL,
	target_id INTEGER NOT NULL,
	/* content_hidden BOOLEAN NOT NULL DEFAULT false, */
	PRIMARY KEY (user_id, target_type, target_id)
);

-- create triggers to delete community records associated through target_type and target_id
-- should things be archived or outright deleted?

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
