CREATE TABLE strava_activities (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    strava_id     BIGINT NOT NULL,
    title         VARCHAR(255),
    activity_type VARCHAR(50),
    start_time    TIMESTAMPTZ NOT NULL,
    distance      NUMERIC(10, 2), -- meters
    is_private    BOOLEAN DEFAULT FALSE,
    is_flagged    BOOLEAN DEFAULT FALSE,
    raw_json      JSONB,
    synced_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(user_id, strava_id)
);

CREATE INDEX idx_strava_activities_user_id ON strava_activities(user_id);
CREATE INDEX idx_strava_activities_start_time ON strava_activities(start_time);