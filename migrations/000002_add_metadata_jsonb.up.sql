ALTER TABLE destinations
    ADD COLUMN metadata JSONB NOT NULL DEFAULT '{}';

ALTER TABLE destinations
    DROP COLUMN IF EXISTS description;
