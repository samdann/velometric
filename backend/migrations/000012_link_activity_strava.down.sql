DROP INDEX IF EXISTS idx_activities_strava_activity_id;

ALTER TABLE activities DROP COLUMN IF EXISTS strava_activity_id;
