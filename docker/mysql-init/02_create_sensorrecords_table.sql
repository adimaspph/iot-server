CREATE TABLE IF NOT EXISTS sensor_records
(
    record_id    BIGINT AUTO_INCREMENT PRIMARY KEY,
    sensor_id    BIGINT       NOT NULL,
    sensor_value DOUBLE       NOT NULL,
    timestamp    TIMESTAMP(6) NOT NULL,
    UNIQUE KEY unique_sensor (sensor_id, sensor_value, timestamp),
    CONSTRAINT fk_sensors_records FOREIGN KEY (sensor_id)
        REFERENCES sensors (sensor_id)
        ON DELETE CASCADE
        ON UPDATE CASCADE
);