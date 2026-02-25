-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    ftp INTEGER,                    -- Functional Threshold Power (watts)
    max_hr INTEGER,                 -- Maximum heart rate (bpm)
    weight DECIMAL(5,2),            -- Weight in kg
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Activities table
CREATE TABLE activities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    sport VARCHAR(50) NOT NULL DEFAULT 'cycling',
    start_time TIMESTAMPTZ NOT NULL,

    -- Basic metrics
    duration INTEGER NOT NULL,              -- seconds
    distance DECIMAL(10,2) NOT NULL,        -- meters
    elevation_gain DECIMAL(8,2),            -- meters

    -- Power metrics
    avg_power INTEGER,                      -- watts
    max_power INTEGER,                      -- watts
    normalized_power INTEGER,               -- watts (NP)
    tss DECIMAL(6,2),                       -- Training Stress Score
    intensity_factor DECIMAL(4,3),          -- IF
    variability_index DECIMAL(4,3),         -- VI (NP/AP)

    -- Heart rate metrics
    avg_hr INTEGER,                         -- bpm
    max_hr INTEGER,                         -- bpm

    -- Cadence metrics
    avg_cadence INTEGER,                    -- rpm
    max_cadence INTEGER,                    -- rpm

    -- Speed metrics
    avg_speed DECIMAL(6,2),                 -- m/s
    max_speed DECIMAL(6,2),                 -- m/s

    -- Other
    calories INTEGER,
    avg_temperature DECIMAL(4,1),           -- Celsius

    -- File storage
    fit_file_url VARCHAR(512),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_activities_user_id ON activities(user_id);
CREATE INDEX idx_activities_start_time ON activities(start_time DESC);

-- Activity records (time-series data) - TimescaleDB hypertable
CREATE TABLE activity_records (
    activity_id UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    timestamp TIMESTAMPTZ NOT NULL,

    -- Position
    lat DECIMAL(10,7),
    lon DECIMAL(10,7),
    altitude DECIMAL(7,2),                  -- meters

    -- Distance
    distance DECIMAL(10,2),                 -- cumulative meters

    -- Power
    power INTEGER,                          -- watts

    -- Heart rate
    heart_rate INTEGER,                     -- bpm

    -- Cadence
    cadence INTEGER,                        -- rpm

    -- Speed
    speed DECIMAL(6,2),                     -- m/s

    -- Temperature
    temperature DECIMAL(4,1),               -- Celsius

    -- Pedaling dynamics
    left_right_balance DECIMAL(4,1),        -- percentage (50 = balanced)
    left_torque_effectiveness DECIMAL(5,2), -- percentage
    right_torque_effectiveness DECIMAL(5,2),
    left_pedal_smoothness DECIMAL(5,2),     -- percentage
    right_pedal_smoothness DECIMAL(5,2),

    -- Gradient (computed from GPS)
    gradient DECIMAL(5,2),                  -- percentage

    PRIMARY KEY (activity_id, timestamp)
);

-- Convert to TimescaleDB hypertable (only if TimescaleDB is installed)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'timescaledb') THEN
        PERFORM create_hypertable('activity_records', 'timestamp',
            chunk_time_interval => INTERVAL '1 day',
            if_not_exists => TRUE
        );
    END IF;
END $$;

CREATE INDEX idx_activity_records_activity_id ON activity_records(activity_id, timestamp DESC);

-- Activity laps
CREATE TABLE activity_laps (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    activity_id UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    lap_number INTEGER NOT NULL,
    start_time TIMESTAMPTZ NOT NULL,

    -- Metrics
    duration INTEGER NOT NULL,              -- seconds
    distance DECIMAL(10,2) NOT NULL,        -- meters

    -- Power
    avg_power INTEGER,
    max_power INTEGER,
    normalized_power INTEGER,

    -- Heart rate
    avg_hr INTEGER,
    max_hr INTEGER,

    -- Cadence
    avg_cadence INTEGER,

    -- Speed
    avg_speed DECIMAL(6,2),
    max_speed DECIMAL(6,2),

    -- Elevation
    ascent DECIMAL(8,2),                    -- meters gained
    descent DECIMAL(8,2),                   -- meters lost

    -- Trigger (manual, auto, distance, etc.)
    trigger VARCHAR(50),

    UNIQUE(activity_id, lap_number)
);

CREATE INDEX idx_activity_laps_activity_id ON activity_laps(activity_id);

-- Power curve (best efforts at each duration)
CREATE TABLE activity_power_curve (
    activity_id UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    duration_seconds INTEGER NOT NULL,      -- 1, 5, 10, 30, 60, 300, 600, 1200, 3600, etc.
    best_power INTEGER NOT NULL,            -- watts

    PRIMARY KEY (activity_id, duration_seconds)
);

-- Activity events (gear changes, pauses, etc.)
CREATE TABLE activity_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    activity_id UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    timestamp TIMESTAMPTZ NOT NULL,
    event_type VARCHAR(50) NOT NULL,        -- gear_change, pause, resume, lap, etc.
    data JSONB,                             -- event-specific data

    UNIQUE(activity_id, timestamp, event_type)
);

CREATE INDEX idx_activity_events_activity_id ON activity_events(activity_id);

-- Power zones configuration per user
CREATE TABLE user_power_zones (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    zone_number INTEGER NOT NULL,           -- 1-7
    name VARCHAR(50) NOT NULL,              -- Recovery, Endurance, Tempo, etc.
    min_percentage DECIMAL(5,2) NOT NULL,   -- % of FTP
    max_percentage DECIMAL(5,2),            -- % of FTP (null for zone 7)
    color VARCHAR(7),                       -- hex color

    UNIQUE(user_id, zone_number)
);

-- Heart rate zones configuration per user
CREATE TABLE user_hr_zones (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    zone_number INTEGER NOT NULL,           -- 1-5
    name VARCHAR(50) NOT NULL,
    min_percentage DECIMAL(5,2) NOT NULL,   -- % of max HR
    max_percentage DECIMAL(5,2),            -- % of max HR
    color VARCHAR(7),

    UNIQUE(user_id, zone_number)
);

-- Updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply updated_at triggers
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_activities_updated_at
    BEFORE UPDATE ON activities
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
