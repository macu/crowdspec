DROP TABLE IF EXISTS spec_block_url;
UPDATE spec_block SET ref_type = NULL, ref_id = NULL WHERE ref_type = 'url'; -- clear existing references

CREATE TABLE spec_block_url (
	id SERIAL PRIMARY KEY,
	block_id INTEGER NOT NULL REFERENCES spec_block (id) ON DELETE CASCADE,
	url VARCHAR(1024) NOT NULL,
	url_title VARCHAR(255),
	url_desc VARCHAR(255),
	url_image_data TEXT
);
