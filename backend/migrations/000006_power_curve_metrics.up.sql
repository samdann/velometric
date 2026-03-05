ALTER TABLE activity_power_curve
    ADD COLUMN avg_speed DECIMAL(6,2),
    ADD COLUMN avg_gradient DECIMAL(5,2),
    ADD COLUMN avg_cadence INTEGER,
    ADD COLUMN avg_lr_balance DECIMAL(4,1),
    ADD COLUMN avg_torque_effectiveness DECIMAL(5,2);
