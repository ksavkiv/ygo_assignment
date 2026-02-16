CREATE TABLE IF NOT EXISTS destinations (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    country     TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    latitude    DOUBLE PRECISION NOT NULL DEFAULT 0,
    longitude   DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
