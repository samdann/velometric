ALTER TABLE activities DROP CONSTRAINT IF EXISTS unique_activity;

ALTER TABLE activities
    ADD CONSTRAINT unique_activity
    UNIQUE (user_id, start_time, sport, distance, duration);
