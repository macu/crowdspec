ALTER TABLE user_account ALTER COLUMN username SET DATA TYPE VARCHAR(25) COLLATE case_insensitive;

-- collation version changed
REINDEX DATABASE crowdspec;
ALTER COLLATION public.case_insensitive REFRESH VERSION;
