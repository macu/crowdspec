-- start recording block created_at
ALTER TABLE spec_block ADD COLUMN created_at TIMESTAMPTZ;

-- initialize existing blocks created_at to spec created_at
UPDATE spec_block SET created_at = (
	SELECT created_at FROM spec WHERE spec.id = spec_block.spec_id
);

ALTER TABLE spec_block ALTER COLUMN created_at SET NOT NULL;


-- start recording spec_url created_at
ALTER TABLE spec_url ADD COLUMN created_at TIMESTAMPTZ;

-- initialize existing blocks created_at to spec created_at
UPDATE spec_url SET created_at = (
	SELECT created_at FROM spec WHERE spec.id = spec_url.spec_id
);

ALTER TABLE spec_url ALTER COLUMN created_at SET NOT NULL;


-- start recording last update timestamps
ALTER TABLE spec ADD COLUMN updated_at TIMESTAMPTZ;
ALTER TABLE spec_subspec ADD COLUMN updated_at TIMESTAMPTZ;
ALTER TABLE spec_block ADD COLUMN updated_at TIMESTAMPTZ;

UPDATE spec SET updated_at = CURRENT_TIMESTAMP(0);
UPDATE spec_subspec SET updated_at = CURRENT_TIMESTAMP(0);
UPDATE spec_block SET updated_at = CURRENT_TIMESTAMP(0);

ALTER TABLE spec ALTER COLUMN updated_at SET NOT NULL;
ALTER TABLE spec_subspec ALTER COLUMN updated_at SET NOT NULL;
ALTER TABLE spec_block ALTER COLUMN updated_at SET NOT NULL;

-- update existing timestamp types to TIMESTAMPTZ
ALTER TABLE user_account ALTER COLUMN created_at SET DATA TYPE TIMESTAMPTZ;
ALTER TABLE user_session ALTER COLUMN expires SET DATA TYPE TIMESTAMPTZ;
ALTER TABLE user_group ALTER COLUMN created_at SET DATA TYPE TIMESTAMPTZ;
ALTER TABLE spec ALTER COLUMN created_at SET DATA TYPE TIMESTAMPTZ;
ALTER TABLE spec_subspec ALTER COLUMN created_at SET DATA TYPE TIMESTAMPTZ;
ALTER TABLE spec_url ALTER COLUMN updated_at SET DATA TYPE TIMESTAMPTZ;
