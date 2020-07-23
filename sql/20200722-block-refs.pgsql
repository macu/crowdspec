-- create custom collation
CREATE COLLATION case_insensitive (provider = icu, locale = 'und@colStrength=primary;colNumeric=yes', deterministic = false);

-- rename subspace to subspec and set collation
ALTER TABLE spec_subspace RENAME TO spec_subspec;
ALTER TABLE spec_block RENAME subspace_id TO subspec_id;
ALTER TABLE spec_subspec RENAME subspace_name TO subspec_name;
ALTER TABLE spec_subspec ALTER COLUMN subspec_name SET DATA TYPE VARCHAR(255) COLLATE case_insensitive;
ALTER TABLE spec_subspec RENAME subspace_desc TO subspec_desc;
CREATE INDEX spec_subspec_name ON spec_subspec (spec_id, subspec_name);
ALTER TYPE spec_block_ref_type RENAME VALUE 'subspace' TO 'subspec';

-- detach URL from block, attach to spec
ALTER TABLE spec_block_url ADD COLUMN spec_id INTEGER REFERENCES spec (id) ON DELETE CASCADE;
UPDATE spec_block_url SET spec_id=spec_block.spec_id FROM spec_block WHERE spec_block.id=spec_block_url.block_id;
ALTER TABLE spec_block_url ALTER COLUMN spec_id SET NOT NULL;
ALTER TABLE spec_block_url DROP COLUMN block_id;
ALTER TABLE spec_block_url RENAME TO spec_url;
ALTER TABLE spec_url ALTER COLUMN url_title SET DATA TYPE VARCHAR(255) COLLATE case_insensitive;
ALTER TABLE spec_url ADD COLUMN updated_at TIMESTAMP;
UPDATE spec_url SET updated_at = CURRENT_TIMESTAMP; -- set timestamps so they can be non null
ALTER TABLE spec_url ALTER COLUMN updated_at SET NOT NULL;
CREATE INDEX spec_url_url ON spec_url (spec_id, url);
CREATE INDEX spec_url_title ON spec_url (spec_id, url_title);
