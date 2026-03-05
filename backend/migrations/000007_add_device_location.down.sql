ALTER TABLE activities
    DROP COLUMN IF EXISTS device_name,
    DROP COLUMN IF EXISTS location;
