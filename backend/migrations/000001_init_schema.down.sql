-- Drop triggers
DROP TRIGGER IF EXISTS update_activities_updated_at ON activities;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables (order matters due to foreign keys)
DROP TABLE IF EXISTS user_hr_zones;
DROP TABLE IF EXISTS user_power_zones;
DROP TABLE IF EXISTS activity_events;
DROP TABLE IF EXISTS activity_power_curve;
DROP TABLE IF EXISTS activity_laps;
DROP TABLE IF EXISTS activity_records;
DROP TABLE IF EXISTS activities;
DROP TABLE IF EXISTS users;

-- Drop extensions
DROP EXTENSION IF EXISTS "uuid-ossp";
