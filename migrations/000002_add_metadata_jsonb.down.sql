ALTER TABLE destinations
    ADD COLUMN description TEXT NOT NULL DEFAULT '';

ALTER TABLE destinations
    DROP COLUMN IF EXISTS metadata;
