-- record time of last update to blocks (add, edit, move, delete)
ALTER TABLE spec ADD COLUMN blocks_updated_at TIMESTAMPTZ;
ALTER TABLE spec_subspec ADD COLUMN blocks_updated_at TIMESTAMPTZ;

UPDATE spec SET blocks_updated_at = (
	SELECT updated_at FROM spec_block
	WHERE spec_id = spec.id
	ORDER BY updated_at DESC
	LIMIT 1
);
UPDATE spec_subspec SET blocks_updated_at = (
	SELECT updated_at FROM spec_block
	WHERE subspec_id = spec_subspec.id
	ORDER BY updated_at DESC
	LIMIT 1
);
