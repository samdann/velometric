ALTER TABLE activity_records
    ALTER COLUMN gradient TYPE DECIMAL(5,2) USING gradient::DECIMAL(5,2);
