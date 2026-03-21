ALTER TABLE activities
    ADD COLUMN strava_activity_id UUID REFERENCES strava_activities(id) ON DELETE SET NULL;

CREATE INDEX idx_activities_strava_activity_id ON activities(strava_activity_id);
