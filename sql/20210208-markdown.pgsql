ALTER TABLE spec_block ADD COLUMN rendered_html TEXT;

-- make content_type mandatory
UPDATE spec_block SET content_type = 'plaintext';
ALTER TABLE spec_block ALTER COLUMN content_type SET NOT NULL;
