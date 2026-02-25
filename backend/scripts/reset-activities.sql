-- Reset script: Deletes all activity data while preserving users and zones
-- Run with: psql -U postgres -d velometric -f scripts/reset-activities.sql

BEGIN;

-- Delete in order respecting foreign keys
DELETE FROM activity_events;
DELETE FROM activity_power_curve;
DELETE FROM activity_laps;
DELETE FROM activity_records;
DELETE FROM activities;

-- Verify counts
SELECT 'activities' as table_name, COUNT(*) as count FROM activities
UNION ALL
SELECT 'activity_records', COUNT(*) FROM activity_records
UNION ALL
SELECT 'activity_laps', COUNT(*) FROM activity_laps
UNION ALL
SELECT 'activity_power_curve', COUNT(*) FROM activity_power_curve
UNION ALL
SELECT 'activity_events', COUNT(*) FROM activity_events;

-- Show preserved data
SELECT 'users' as table_name, COUNT(*) as count FROM users
UNION ALL
SELECT 'user_power_zones', COUNT(*) FROM user_power_zones
UNION ALL
SELECT 'user_hr_zones', COUNT(*) FROM user_hr_zones;

COMMIT;

\echo 'Activity data reset complete. Users and zones preserved.'
