-- verify_write_target
-- verify_read_target (also used to verify adding a comment under the target)
-- verify_delete_target

DROP FUNCTION IF EXISTS verify_delete_target;
DROP FUNCTION IF EXISTS verify_write_target;
DROP FUNCTION IF EXISTS verify_read_target;
DROP FUNCTION IF EXISTS verify_read_target_public;

CREATE FUNCTION verify_write_target (
	write_target_type spec_community_target_type,
	write_target_id INT,
	write_user_id INT
)
RETURNS BOOLEAN
LANGUAGE PLPGSQL
AS $$
DECLARE
	access_allowed BOOLEAN := FALSE;
BEGIN

	CASE write_target_type

	WHEN 'spec' THEN

		SELECT EXISTS(
			SELECT * FROM spec
			WHERE spec.id = write_target_id
				AND spec.owner_type = 'user' AND spec.owner_id = write_user_id
		) INTO access_allowed;

	WHEN 'subspec' THEN

		SELECT EXISTS(
			SELECT * FROM spec_subspec
			INNER JOIN spec ON spec.id = spec_subspec.spec_id
			WHERE spec_subspec.id = write_target_id
				AND spec.owner_type = 'user' AND spec.owner_id = write_user_id
		) INTO access_allowed;

	WHEN 'block' THEN

		SELECT EXISTS(
			SELECT * FROM spec_block
			INNER JOIN spec ON spec.id = spec_block.spec_id
			WHERE spec_block.id = write_target_id
				AND spec.owner_type = 'user' AND spec.owner_id = write_user_id
		) INTO access_allowed;

	WHEN 'comment' THEN

		-- only allow original author to modify comment
		SELECT EXISTS(
			SELECT * FROM spec_community_comment
			WHERE spec_community_comment.id = write_target_id
				AND spec_community_comment.user_id = write_user_id
		) INTO access_allowed;

	END CASE;

	RETURN access_allowed;

END $$;

CREATE FUNCTION verify_read_target (
	read_target_type spec_community_target_type,
	read_target_id INT,
	read_user_id INT
)
RETURNS BOOLEAN
LANGUAGE PLPGSQL
AS $$
DECLARE
	access_allowed BOOLEAN := FALSE;
	comment_record RECORD;
BEGIN

	CASE read_target_type

	WHEN 'spec' THEN

		SELECT EXISTS(
			SELECT * FROM spec
			WHERE spec.id = read_target_id
				AND (
					spec.is_public
					OR (spec.owner_type = 'user' AND spec.owner_id = read_user_id)
				)
		) INTO access_allowed;

	WHEN 'subspec' THEN

		SELECT EXISTS(
			SELECT * FROM spec_subspec
			INNER JOIN spec ON spec.id = spec_subspec.spec_id
			WHERE spec_subspec.id = read_target_id
				AND (
					(spec.is_public AND NOT spec_subspec.is_private)
					OR (spec.owner_type = 'user' AND spec.owner_id = read_user_id)
				)
		) INTO access_allowed;

	WHEN 'block' THEN

		SELECT EXISTS(
			SELECT * FROM spec_block
			INNER JOIN spec ON spec.id = spec_block.spec_id
			LEFT JOIN spec_subspec ON spec_subspec.id = spec_block.subspec_id
			WHERE spec_block.id = read_target_id
				AND (
					(spec.is_public
						AND (spec_subspec.is_private IS NULL OR NOT spec_subspec.is_private)
					)
					OR (spec.owner_type = 'user' AND spec.owner_id = read_user_id)
				)
		) INTO access_allowed;

	WHEN 'comment' THEN

		-- allow reading comments under public spec and subspec

		WITH RECURSIVE comment_stack(id, target_type, target_id) AS (
			-- Anchor
			SELECT id, target_type, target_id
			FROM spec_community_comment
			WHERE id = read_target_id
			-- Recursive Member
			UNION ALL
			SELECT cc.id, cc.target_type, cc.target_id
			FROM spec_community_comment cc, comment_stack cs
			WHERE cs.target_type = 'comment'
				AND cs.target_id = cc.id
		)
		SELECT cc.target_type, cc.target_id
		INTO comment_record
		FROM spec_community_comment cc
		INNER JOIN comment_stack cs ON cs.id = cc.id
		WHERE cc.target_type != 'comment';

		CASE comment_record.target_type
		WHEN 'spec' THEN
			IF verify_read_target('spec', comment_record.target_id, read_user_id) THEN
				RETURN TRUE;
			END IF;
		WHEN 'subspec' THEN
			IF verify_read_target('subspec', comment_record.target_id, read_user_id) THEN
				RETURN TRUE;
			END IF;
		WHEN 'block' THEN
			IF verify_read_target('block', comment_record.target_id, read_user_id) THEN
				RETURN TRUE;
			END IF;
		ELSE
			RETURN FALSE;
		END CASE;

		-- allow reading comments under user's own comment

		SELECT EXISTS(
			WITH RECURSIVE comment_stack(id, target_type, target_id) AS (
				-- Anchor
				SELECT id, target_type, target_id
				FROM spec_community_comment
				WHERE id = read_target_id
				-- Recursive Member
				UNION ALL
				SELECT cc.id, cc.target_type, cc.target_id
				FROM spec_community_comment cc, comment_stack cs
				WHERE cs.target_type = 'comment'
					AND cs.target_id = cc.id
			)
			SELECT *
			FROM spec_community_comment cc
			INNER JOIN comment_stack cs ON cs.id = cc.id
			WHERE cc.user_id = read_user_id
		) INTO access_allowed;

	END CASE;

	RETURN access_allowed;

END $$;

CREATE FUNCTION verify_read_target_public (
	read_target_type spec_community_target_type,
	read_target_id INT
)
RETURNS BOOLEAN
LANGUAGE PLPGSQL
AS $$
DECLARE
	access_allowed BOOLEAN := FALSE;
	comment_record RECORD;
BEGIN

	CASE read_target_type

	WHEN 'spec' THEN

		SELECT EXISTS(
			SELECT * FROM spec
			WHERE spec.id = read_target_id
				AND spec.is_public
		) INTO access_allowed;

	WHEN 'subspec' THEN

		SELECT EXISTS(
			SELECT * FROM spec_subspec
			INNER JOIN spec ON spec.id = spec_subspec.spec_id
			WHERE spec_subspec.id = read_target_id
				AND spec.is_public
				AND NOT spec_subspec.is_private
		) INTO access_allowed;

	WHEN 'block' THEN

		SELECT EXISTS(
			SELECT * FROM spec_block
			INNER JOIN spec ON spec.id = spec_block.spec_id
			LEFT JOIN spec_subspec ON spec_subspec.id = spec_block.subspec_id
			WHERE spec_block.id = read_target_id
				AND spec.is_public
				AND (spec_subspec.is_private IS NULL OR NOT spec_subspec.is_private)
		) INTO access_allowed;

	WHEN 'comment' THEN

		-- allow reading comments under public spec and subspec

		WITH RECURSIVE comment_stack(id, target_type, target_id) AS (
			-- Anchor
			SELECT id, target_type, target_id
			FROM spec_community_comment
			WHERE id = read_target_id
			-- Recursive Member
			UNION ALL
			SELECT cc.id, cc.target_type, cc.target_id
			FROM spec_community_comment cc, comment_stack cs
			WHERE cs.target_type = 'comment'
				AND cs.target_id = cc.id
		)
		SELECT cc.target_type, cc.target_id
		INTO comment_record
		FROM spec_community_comment cc
		INNER JOIN comment_stack cs ON cs.id = cc.id
		WHERE cc.target_type != 'comment';

		CASE comment_record.target_type
		WHEN 'spec' THEN
			RETURN verify_read_target_public('spec', comment_record.target_id);
		WHEN 'subspec' THEN
			RETURN verify_read_target_public('subspec', comment_record.target_id);
		WHEN 'block' THEN
			RETURN verify_read_target_public('block', comment_record.target_id);
		ELSE
			RETURN FALSE;
		END CASE;

	END CASE;

	RETURN access_allowed;

END $$;

CREATE FUNCTION verify_delete_target (
	delete_target_type spec_community_target_type,
	delete_target_id INT,
	delete_user_id INT
)
RETURNS BOOLEAN
LANGUAGE PLPGSQL
AS $$
DECLARE
	access_allowed BOOLEAN := FALSE;
BEGIN

	CASE delete_target_type

	WHEN 'spec' THEN

		RETURN verify_write_target('spec', delete_target_id, delete_user_id);

	WHEN 'subspec' THEN

		RETURN verify_write_target('subspec', delete_target_id, delete_user_id);

	WHEN 'block' THEN

		RETURN verify_write_target('block', delete_target_id, delete_user_id);

	WHEN 'comment' THEN

		-- allow original author or spec owner to delete comment
		SELECT EXISTS(
			SELECT * FROM spec_community_comment
			INNER JOIN spec ON spec.id = spec_community_comment.spec_id
			WHERE spec_community_comment.id = delete_target_id
				AND (
					spec_community_comment.user_id = delete_user_id
					OR (spec.owner_type = 'user' AND spec.owner_id = delete_user_id)
				)
		) INTO access_allowed;

	END CASE;

	RETURN access_allowed;

END $$;
