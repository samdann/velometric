ALTER TABLE activity_power_curve
    DROP COLUMN IF EXISTS avg_speed,
    DROP COLUMN IF EXISTS avg_gradient,
    DROP COLUMN IF EXISTS avg_cadence,
    DROP COLUMN IF EXISTS avg_lr_balance,
    DROP COLUMN IF EXISTS avg_torque_effectiveness;
