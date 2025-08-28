CREATE TABLE IF NOT EXISTS sensors
(
    sensor_id   BIGINT AUTO_INCREMENT PRIMARY KEY,
    id1         VARCHAR(20)  NOT NULL,
    id2         BIGINT       NOT NULL,
    sensor_type VARCHAR(50)  NOT NULL,
    UNIQUE KEY unique_sensor (id1, id2)
);