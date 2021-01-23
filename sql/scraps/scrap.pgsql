

-- find ID of policy pertaining to target block
--CREATE FUNCTION find_block_policy (INTEGER);

-- find ID of policy pertaining to target comment
--CREATE FUNCTION find_comment_policy (INTEGER);

-- include base
-- include self tags
-- include all matching conditional policies
-- tags appear in order according to first appearance in the include stack,
-- grouped with their base tags expansion group
--CREATE FUNCTION expand_policy_tags (INTEGER);

-- TODO guard against exclusive conditional policy basing itself on its parent;
-- when that happens treat it as non-exclusive, and take next matching exclusive policy from base or combine with other matching non-exclusive conditional policies from base if no other exclusive policy takes precedence
-- or, include matching conditional base policies but skip exclusive ones
-- what about exclusive policies from base if base is not parent?

/*
Example 1:

policy A
	conditional policy C
		exclusive
		if pin tagged locked
		community disabled
	subs policy B
		base policy A

- block
	set to policy A
	- block
		pin tagged locked
		matches policy C through base of subs policy B
		- block
			inherits policy C
			- block
				inherits policy C

Example 2:

policy A
	conditional policy C
		non-exclusive
		if pin tagged locked
		community disabled
	subs policy B
		base policy A
		conditional policy D
			non-exclusive
			if pin tagged unlocked
			community enabled

- block
	set to policy A
	- block
		pin tagged locked
		natches policy B with C enabled through base A (community disabled)
		- block
			pin tagged unlocked
			matches policy B with D enabled (community enabled)
			- block
				matches policy B with community enabled inherited

*/

--
CREATE FUNCTION expand_policy_tags
(policyId INTEGER, specId INT, subspecId INT, blockIds INT[], commentIds INT[], out tagIds INT[])
LANGUAGE PLPGSQL
AS $$
DECLARE
	levelType spec_community_target_type;
	levelId INT;
	levelTagIds INT[];
	renderPolicyIds INT[];
	replacementPolicyId INT;
BEGIN

	-- determine initial level
	IF specId IS NOT NULL THEN
		-- load spec tags
		levelType := 'spec';
		levelId := specId;
	ELSE IF subspecId IS NOT NULL THEN
		-- load subspec tags
		levelType := 'subspec';
		levelId = subspecId;
	ELSE IF array_length(blockIds) > 0 THEN
		-- load first block tags
		levelType := 'block';
		levelId := blockIds[0];
	ELSE
		-- load first comment tags
		levelType := 'comment';
		levelId := commentIds[0];
	END IF;

	renderPolicyIds = ARRAY[policyId];

	<<step_levels>>
	LOOP

		-- select level tags
		SELECT tag_id
		FROM spec_community_tag
		INTO levelTagIds
		WHERE target_type = levelType
		AND tag_id = levelId
		AND admin_pin IS NOT NULL;

		-- check for matching exclusive conditional policy
		replacementPolicyId := expand_policy_first_exclusive_policy(policyId, levelTagIds)
		IF replacementPolicyId IS NOT NULL THEN
			renderPolicyIds := ARRAY[replacementPolicyId];
		END IF;

	-- TODO
	-- if spec, subspec or block take subs policy
	-- if comment, take comments policy

	END LOOP step_levels;


END $$;
