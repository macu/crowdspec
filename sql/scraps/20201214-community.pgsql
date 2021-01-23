DROP TABLE IF EXISTS spec_community_tag_vote;
DROP TYPE IF EXISTS spec_community_tag_vote_type;
DROP TABLE IF EXISTS spec_community_tag;
DROP TYPE IF EXISTS spec_community_tag_pin_type;
DROP TABLE IF EXISTS spec_community_comment;
DROP TABLE IF EXISTS spec_community_config;
DROP TYPE IF EXISTS spec_community_target_type;
DROP TABLE IF EXISTS tag_list_tag;
DROP TABLE IF EXISTS tag_list;
DROP TABLE IF EXISTS tag;

CREATE TABLE tag (
	id SERIAL PRIMARY KEY, -- record sequence added
	created_at TIMESTAMPTZ NOT NULL, -- record time first added
	tag_name VARCHAR(50) NOT NULL UNIQUE -- creates an index for unique
);

CREATE TYPE spec_community_target_type AS ENUM ('spec', 'subspec', 'block', 'comment');

CREATE TABLE spec_community_config (
	spec_id INTEGER,-- NOT NULL REFERENCES spec (id) ON DELETE CASCADE,
	target_type spec_community_target_type NOT NULL,
	target_id INTEGER NOT NULL,

	policy_id INTEGER,-- NOT NULL REFERENCES spec_community_policy (id) ON DELETE CASCADE,

	admin_note TEXT,

	PRIMARY KEY (target_type, target_id)
);

CREATE TABLE spec_community_policy (
	id SERIAL PRIMARY KEY,

	base_policy_id INTEGER REFERENCES spec_community_policy (id), -- initialize this policy from base

	policy_name VARCHAR(255),

	-- these properties are inherited by conditional policies, even if exclusive,
	-- and inherited by subs and comments policies
	community_disabled BOOLEAN, -- inherited by subs and comments until next override
	open_tagging_disabled BOOLEAN, -- inherited by subs and comments until next override

	subs_policy_id INTEGER REFERENCES spec_community_policy (id), -- passed down if defined, else this policy applies

	-- TODO is this needed with spec_community_conditional_policy.target_type?
	comments_policy_id INTEGER REFERENCES spec_community_policy (id), -- passed down if not overridden by subs's comment policy
);

CREATE TYPE spec_community_tag_policy AS ENUM ('restricted', 'admin-only');

CREATE TABLE spec_community_policy_tag (
	policy_id INTEGER NOT NULL REFERENCES spec_community_policy (id),
	conditional_tag_id INTEGER NOT NULL REFERENCES tag (id),
	tag_id INTEGER NOT NULL REFERENCES tag (id),
	tag_policy spec_community_tag_policy,
	order_number INTEGER,

	PRIMARY KEY (policy_id, tag_id)
);

CREATE TABLE spec_community_conditional_policy (
	parent_policy_id INTEGER NOT NULL REFERENCES spec_community_policy (id),, -- default policy if no conditional policies apply
	order_number INTEGER NOT NULL, -- order policies are expanded
	exclusive_on_match BOOLEAN NOT NULL DEFAULT false, -- first matching exclusive policy replaces current policy
	target_type spec_community_tag_target_type, -- match if specified
	match_all_tags INTEGER ARRAY, -- match with `<@` is contained by
	match_any_tags INTEGER ARRAY, -- match with `&&` overlap (have elements in common)
	match_no_tags INTEGER ARRAY, -- match with NOT overlap
	policy_id INTEGER NOT NULL REFERENCES spec_community_policy (id),, -- invoke policy if all conditions met
);

CREATE TABLE spec_community_comment (
	id SERIAL PRIMARY KEY,
	spec_id INTEGER,-- NOT NULL REFERENCES spec (id) ON DELETE CASCADE,
	target_type spec_community_target_type NOT NULL,
	target_id INTEGER,-- NOT NULL,
	user_id INTEGER,-- NOT NULL REFERENCES user_account (id),
	created_at TIMESTAMPTZ,-- NOT NULL,
	updated_at TIMESTAMPTZ,
	comment_body TEXT --NOT NULL
);

CREATE TYPE spec_community_tag_pin_type AS ENUM ('check', 'poll', 'badge', 'enabled');

CREATE TABLE spec_community_tag (
	spec_id INTEGER NOT NULL REFERENCES spec (id) ON DELETE CASCADE,
	target_type spec_community_target_type NOT NULL,
	target_id INTEGER NOT NULL,
	tag_id INTEGER NOT NULL REFERENCES tag (id) ON DELETE CASCADE,
	added_by_user_id INTEGER NOT NULL REFERENCES user_account (id),
	added_at TIMESTAMPTZ NOT NULL,
	assent_votes INTEGER NOT NULL DEFAULT 0,
	dissent_votes INTEGER NOT NULL DEFAULT 0,
	-- {n} / NULL -> NULL; return 0 when no votes
	vote_ratio REAL GENERATED ALWAYS AS
		COALESCE((assent_votes - dissent_votes) / NULLIF(assent_votes + dissent_votes, 0), 0) STORED,
	admin_pin spec_community_tag_pin_type,
	PRIMARY KEY (target_type, target_id, tag_id)
);

CREATE TYPE spec_community_tag_vote_type AS ENUM ('assent', 'dissent');

CREATE TABLE spec_community_tag_vote (
	spec_id INTEGER NOT NULL REFERENCES spec (id) ON DELETE CASCADE,
	target_type spec_community_tag_target_type NOT NULL,
	target_id INTEGER NOT NULL,
	tag_id INTEGER NOT NULL REFERENCES tag (id) ON DELETE CASCADE,
	user_id INTEGER NOT NULL REFERENCES user_account (id),
	updated_at TIMESTAMPTZ NOT NULL,
	vote spec_community_tag_vote_type NOT NULL,
	PRIMARY KEY (target_type, target_id, tag_id, user_id)
);

-- TODO use this
CREATE TYPE community_policy_path_node AS (
	node_id INTEGER,
	node_type spec_community_tag_target_type
);

-- returns the items from a comment up to the nearest policy
-- out policyId will be the ID of the first policy found
-- out subspecId will be defined if policy comes from subspec
-- out blockIds will be ordered from leaf (parent of top comment) .. topmost
-- out commentIds will be ordered leaf (given) .. topmost
CREATE OR REPLACE FUNCTION get_comment_policy_path
(specId INTEGER, commentId INTEGER,
	out policyId INTEGER,
	out pathIds INTEGER[], out pathTypes spec_community_tag_target_type[])
LANGUAGE PLPGSQL
AS $$
DECLARE

	cursorComment RECORD;
	topmostComment RECORD;
	superIds INTEGER[];
	superTypes spec_community_tag_target_type[];

BEGIN

	FOR cursorComment IN
		WITH RECURSIVE select_comments (comment_id, target_type, target_id, policy_id, lvl) AS (
			SELECT c.id, c.target_type, c.target_id, cfg.policy_id, 1
				FROM spec_community_comment AS c
				LEFT JOIN spec_community_config AS cfg
					ON cfg.target_type = 'comment' AND cfg.target_id = c.id
				WHERE c.spec_id = specId AND c.id = commentId
			UNION
			SELECT c2.id, c2.target_type, c2.target_id, cfg2.policy_id, s.lvl + 1
				FROM spec_community_comment AS c2
				INNER JOIN select_comments AS s
					ON s.target_type = 'comment' AND c2.id = s.target_id
				LEFT JOIN spec_community_config AS cfg2
					ON cfg2.target_type = 'comment' AND cfg2.target_id = c2.id
				WHERE c2.spec_id = specId AND s.policy_id IS NULL
		)
		SELECT *
		FROM select_comments
		ORDER BY lvl DESC
	LOOP
		IF topmostComment IS NULL THEN
			topmostComment := cursorComment;
		END IF;
		pathIds := array_prepend(pathIds, cursorComment.comment_id);
		pathTypes := array_prepend(pathTypes, 'comment');
	END LOOP;

	IF topmostComment IS NULL THEN
		-- Nothing found
		RETURN;
	END IF;

	IF topmostComment.policy_id IS NOT NULL THEN
		-- Policy found
		policyId := topmostComment.policy_id;
		RETURN;
	END IF;

	IF topmostComment.target_type = 'block' THEN
		policyId, superIds, superTypes := get_block_policy_path(specId, topmostComment.target_id);
	ELSEIF topmostComment.target_type = 'subspec' THEN
		policyId, superIds, superTypes := get_subspec_policy_path(specId, topmostComment.target_id);
	ELSEIF topmostComment.target_type = 'spec' THEN
		policyId, superIds, superTypes := get_spec_policy_path(specId);
	END IF;

	IF array_length(superIds, 1) > 0 THEN
		pathIds := array_cat(superIds, pathIds);
		pathTypes := array_cat(superTypes, pathTypes);
	END IF;

END $$;

--
CREATE OR REPLACE FUNCTION get_block_policy_path
(specId INTEGER, blockId INTEGER,
	out policyId INTEGER,
	out pathIds INTEGER[], out pathTypes spec_community_tag_target_type[])
LANGUAGE PLPGSQL
AS $$
DECLARE

	cursorBlock RECORD;
	topmostBlock RECORD;
	subspecId INTEGER;
	superIds INTEGER[];
	superTypes spec_community_tag_target_type[];

BEGIN

	FOR cursorBlock IN
		WITH RECURSIVE select_blocks (block_id, parent_id, policy_id, lvl) AS (
			SELECT b.id, b.parent_id, cgf.policy_id, 1
				FROM spec_block AS b
				LEFT JOIN spec_community_config AS cfg
					ON cfg.target_type = 'block' AND cfg.target_id = b.id
				WHERE b.spec_id = specId AND b.id = blockId
			UNION
			SELECT b2.id, b2.parent_id, cfg2.policy_id, s.lvl + 1
				FROM spec_block AS b2
				INNER JOIN select_blocks AS s
					ON b2.id = s.parent_id
				LEFT JOIN spec_community_config AS cfg2
					ON cfg2.target_type = 'block' AND cfg2.target_id = b2.id
				WHERE b2.spec_id = specId AND s.policy_id IS NULL
		)
		SELECT *
		FROM select_blocks
		ORDER BY lvl DESC
	LOOP
		IF topmostBlock IS NULL THEN
			topmostBlock := cursorBlock;
		END IF;
		pathIds := array_prepend(pathIds, cursorBlock.block_id);
		pathTypes := array_prepend(pathTypes, 'block');
	END LOOP;

	IF topmostBlock IS NULL THEN
		-- Nothing found
		RETURN;
	END IF;

	IF topmostBlock.policy_id IS NOT NULL THEN
		-- Policy found
		policyId := topmostBlock.policy_id;
		RETURN;
	END IF;

	-- Check whether block is member of a subspec
	SELECT subspec_id
		FROM spec_block
		INTO subspecId
		WHERE id = topmostBlock.id;

	IF subspecId IS NOT NULL THEN
		policyId, superIds, superTypes := get_subspec_policy_path(specId, subspecId);
	ELSE
		policyId, superIds, superTypes := get_spec_policy_path(specId);
	END IF;

	IF array_length(superIds, 1) > 0 THEN
		pathIds := array_cat(superIds, pathIds);
		pathTypes := array_cat(superTypes, pathTypes);
	END IF;

END $$;

--
CREATE OR REPLACE FUNCTION get_subspec_policy_path
(specId INTEGER, subspecId INTEGER,
	out policyId INTEGER,
	out pathIds INTEGER[], out pathTypes spec_community_tag_target_type[])
LANGUAGE PLPGSQL
AS $$
DECLARE

	cursorSubspec RECORD;
	superIds INTEGER[];
	superTypes spec_community_tag_target_type[];

BEGIN

	SELECT spec_subspec.id AS subspec_id, cfg.policy_id
		FROM spec_subspec
		INTO cursorSubspec
		LEFT JOIN spec_community_config AS cfg
			ON cfg.target_type = 'subspec' AND cfg.target_id = spec_subspec.id
		WHERE spec_subspec.spec_id = specId
			AND spec_subspec.id = subspecId;

	IF cursorSubspec IS NOT NULL THEN
		pathIds := ARRAY[cursorSubspec.subspec_id];
		pathTypes := ARRAY['subspec'];
		policyId := cursorSubspec.policy_id;
	END IF;

	IF policyId IS NULL THEN
		policyId, superIds, superTypes := get_spec_policy_path(specId);
	END IF;

	IF array_length(superIds, 1) > 0 THEN
		pathIds := array_cat(superIds, pathIds);
		pathTypes := array_cat(superTypes, pathTypes);
	END IF;

END $$;

--
CREATE OR REPLACE FUNCTION get_spec_policy_path
(specId INTEGER,
	out policyId INTEGER,
	out pathIds INTEGER[], out pathTypes spec_community_tag_target_type[])
LANGUAGE PLPGSQL
AS $$
DECLARE

	cursorSpec RECORD;

BEGIN

	SELECT spec.id AS spec_id, cft.policy_id
		FROM spec
		INTO cursorSpec
		LEFT JOIN spec_community_config AS cfg
			ON cfg.target_type = 'spec' AND cfg.target_id = spec.id
		WHERE spec.id = specId;

	IF cursorSpec IS NOT NULL THEN
		pathIds := ARRAY[specId];
		pathTypes := ARRAY['spec'];
		policyId := cursorSpec.policy_id;
	END IF;

END $$;


-- policyId is the policy defined at the beginning of the given path;
-- returns the expressed policy the policy for the node at the end of the path
CREATE OR REPLACE FUNCTION express_policy_to_node
(specId INTEGER, policyId INTEGER,
	pathIds INTEGER[], out pathTypes spec_community_tag_target_type[],
	out policyIds INTEGER[])
LANGUAGE PLPGSQL
AS $$
DECLARE

	i INTEGER := 1;
	inheritOnNone BOOLEAN;

BEGIN

	IF array_length(pathIds) = 0 THEN
		RETURN;
	END IF;

	-- start with given policy
	policyIds := express_policy(specId, ARRAY[policyId], pathIds[0], pathTypes[0]);

	-- express policy for each step in path
	WHILE i < array_length(pathIds) LOOP
		-- step
		inheritOnNone := pathTypes[i - 1] = pathTypes[i];
		CASE pathTypes[i]
			WHEN 'comment' THEN
				policyIds := extract_comments_policy(policyIds, inheritOnNone);
			ELSE
				policyIds := extract_subs_policy(policyIds, inheritOnNone);
		END CASE;
		-- express policy for sub or comment
		policyIds := express_policy(specId, policyIds, pathIds[i], pathTypes[i]);
		i := i + 1;
	END LOOP;

END $$;

CREATE TYPE policy_cond AS (
	policy_id INTEGER NOT NULL, -- refers to conditional policy
	all_tags INTEGER[],
	any_tags INTEGER[],
	not_tags INTEGER[],
	exclusive_on_match BOOLEAN
);

CREATE TYPE policy_expression AS (
	policy_id INTEGER NOT NULL, -- refers to owner policy, repeated for all conditional components
	policy_cond policy_cond,
	policy_position INTEGER[] -- array of base IDs from root down to policy_id
);



-- fully expands and expresses the given policy set,
-- including conditional policies matching pinned tags on the specified node,
-- returning the policies themselves, following their base policies,
-- as well as all matching conditional policies, in order of appearance
CREATE OR REPLACE FUNCTION express_policy
(specId INTEGER, expressPolicyIds INTEGER[],
	nodeId INTEGER, nodeType spec_community_tag_target_type,
	out policyIds INTEGER[])
LANGUAGE PLPGSQL
AS $$
DECLARE

	pinnedTagIds INTEGER[];
	policyExpressions policy_expression[];

BEGIN

	-- select node pinned tags
	SELECT tag_id
		FROM spec_community_tag
		INTO pinnedTagIds
		WHERE target_type = nodeType AND target_id = nodeId
			AND admin_pin IS NOT NULL;

	-- express
	policyExpressions := express_policy_part(specId, expandPolicyIds, pinnedTagIds, ARRAY[]);

	-- return only IDs
	SELECT e.policy_id FROM unnest(policyExpressions) AS e INTO policyIds;

END $$;

-- fully expands and expresses the given policy set,
-- including conditional policies matching pinned tags on the specified node,
-- returning the policies themselves, following their base policies,
-- as well as all matching conditional policies, in order of appearance
CREATE OR REPLACE FUNCTION express_policy_part
(specId INTEGER, expressPolicyIds INTEGER[], pinnedTagIds INTEGER[],
	inout renderedPolicies policy_expression[])
LANGUAGE PLPGSQL
AS $$
DECLARE

	expandedPolicies policy_expression[]; -- unconditional and unexpressed conditional policies
	mergePolicies policy_expression[]; -- will only contain unconditional and expressed matched policies

	currentPolicy policy_expression;
	skipExclusiveComponents BOOLEAN := false; -- true after first conditional component found

BEGIN

	-- expand bases and include all unexpanded conditional policies,
	-- which will be expressed individually if they are matched;
	-- expand_policy will not return anything already in renderedPolicies
	expandedPolicies := expand_policy(specId, expressPolicyIds, renderedPolicyIds);

	-- aggregate policies into expandedPolicies matching the pinned tags
	FOREACH currentPolicy IN ARRAY expandedPolicies LOOP

		IF currentPolicy.policy_cond IS NULL THEN
			-- policy is non-conditional

			IF (
				SELECT COUNT(*) FROM UNNEST(mergePolicies) AS p
				WHERE p.policy_id = currentPolicy.policy_id
			) > 0 THEN
				-- already present in mergePolicies
				CONTINUE;
			END IF;

			mergePolicies = array_cat(mergePolicies, currentPolicy);

			CONTINUE;
		END IF;

		-- check conditional component for match
		IF (
				array_length(currentPolicy.policy_cond.all_tags) = 0 -- undefined
				OR currentPolicy.policy_cond.all_tags <@ pinnedTagIds -- contained by
			)
			AND (
				array_length(currentPolicy.policy_cond.any_tags) = 0 -- undefined
				OR currentPolicy.policy_cond.any_tags && pinnedTagIds -- overlap
			)
			AND (
				array_length(currentPolicy.policy_cond.not_tags) = 0 -- undefined
				OR (NOT (currentPolicy.policy_cond.not_tags && pinnedTagIds)) -- no overlap
			)
		) THEN
			-- matching
			IF currentPolicy.policy_cond.exclusive_on_match THEN
				-- if exclusive policy inherits from current policy,
				-- include non-exclusive parts of this policy

				IF NOT skipExclusiveComponents THEN
					IF currentPolicy.policy_cond.
				END IF;

				-- first exclusive match,
				-- return expressed policy
				renderedPolicyIds, renderedPolicyConds, renderedPolicyPositions :=
					express_policy_part(specId, ARRAY[expandedPolicyIds[i]],
						 nodeId, nodeType, pinnedTagIds,
						 renderedPolicyIds, renderedPolicyConds, renderedPolicyPositions);
				RETURN;
			ELSE
				-- express conditional policy
				activatedPolicyIds, activatedPolicyConds, activatedPolicyPositions :=
					express_policy_part(specId, ARRAY[expandPolicyIds],
						nodeId, nodeType, pinnedTagIds,
						renderedPolicyIds, renderedPolicyConds, renderedPolicyPositions);
				expandedPolicyIds, expandedPolicyConds, expandedPolicyPositions :=
					merge_policy_sets(expandedPolicyIds, expandedPolicyConds, expandedPolicyPositions,
						activatedPolicyIds, activatedPolicyConds, activatedPolicyPositions);
			END IF;
		END IF;
	END LOOP;

END $$;

-- fully expands the given policy set,
-- returning the policies themselves, following their base policies,
-- as well as all unevaluated conditional policies, in order of appearance.
-- if run consecutively on its own return value it should produce the same output.
CREATE OR REPLACE FUNCTION expand_policy
(specId INTEGER, expandPolicyIds INTEGER[],
	renderedPolicies policy_expression[], -- don't repeat policies present
	out policyExpressions policy_expression[])
LANGUAGE PLPGSQL
AS $$
DECLARE
BEGIN
END $$;

-- merge two policy sets in order; assumes the first set given is already in order
CREATE OR REPLACE FUNCTION merge_policy_sets
(policyExpressions1 policy_expression[], policyExpressions2 policy_expression[],
	out mergedPolicyExpressions policy_expression[])
LANGUAGE PLPGSQL
AS $$
DECLARE
BEGIN
END $$;

-- returns the aggregate subs policy IDs
CREATE OR REPLACE FUNCTION extract_subs_policy
(specId INTEGER, parentPolicyIds INTEGER[],
	out policyIds INTEGER[])
LANGUAGE PLPGSQL
AS $$
DECLARE
BEGIN



END $$;

-- returns the aggregate comment policy IDs
CREATE OR REPLACE FUNCTION extract_comments_policy
(specId INTEGER, parentPolicyIds INTEGER[],
	out policyIds INTEGER[])
LANGUAGE PLPGSQL
AS $$
DECLARE
BEGIN
END $$;
