ALTER TABLE activities
    ADD CONSTRAINT unique_activity
    UNIQUE (user_id, start_time, sport, distance, duration);
