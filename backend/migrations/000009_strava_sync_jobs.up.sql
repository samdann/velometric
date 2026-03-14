CREATE TABLE strava_sync_jobs (
    id            UUID PRIMARY KEY,
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status        VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    limit_count   INTEGER NOT NULL DEFAULT 0,
    fetched_count INTEGER,
    updated_count INTEGER,
    created_count INTEGER,
    error_message TEXT,
    started_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    fetched_at    TIMESTAMPTZ,
    completed_at  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON strava_sync_jobs(user_id);

ALTER TABLE strava_activities ADD COLUMN job_id UUID REFERENCES strava_sync_jobs(id);

CREATE INDEX ON strava_activities(job_id);
