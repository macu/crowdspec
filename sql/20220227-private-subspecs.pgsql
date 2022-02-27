-- allow private subspecs within public specs
DROP INDEX spec_subspec_by_name;
ALTER TABLE spec_subspec ADD COLUMN is_private BOOLEAN NOT NULL DEFAULT FALSE;
CREATE INDEX spec_subspec_by_name ON spec_subspec (spec_id, is_private, subspec_name);
