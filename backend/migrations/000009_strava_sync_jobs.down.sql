DROP INDEX IF EXISTS strava_activities_job_id_idx;
ALTER TABLE strava_activities DROP COLUMN IF EXISTS job_id;
DROP INDEX IF EXISTS strava_sync_jobs_user_id_idx;
DROP TABLE IF EXISTS strava_sync_jobs;
