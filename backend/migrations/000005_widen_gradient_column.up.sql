ALTER TABLE activity_records
    ALTER COLUMN gradient TYPE FLOAT USING gradient::FLOAT;
