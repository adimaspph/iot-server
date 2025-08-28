CREATE INDEX idx_sensor_time ON sensor_records(sensor_id, timestamp);

CREATE INDEX idx_sensors ON sensors(id1, id2);