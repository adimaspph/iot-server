package model

import "time"

type SensorRecord struct {
	SensorValue float64   `json:"sensor_value"`
	Timestamp   time.Time `json:"timestamp"`
}

type SensorResponse struct {
	ID1            string         `json:"id1"`
	ID2            int64          `json:"id2"`
	SensorType     string         `json:"sensor_type"`
	SensorsRecords []SensorRecord `json:"sensor_records"`
}

type CreateSensorRequest struct {
	ID1         string  `json:"id1" validate:"required,uppercase"`
	ID2         int64   `json:"id2" validate:"required"`
	SensorType  string  `json:"sensor_type" validate:"required"`
	SensorValue float64 `json:"sensor_value" validate:"required"`
	Timestamp   string  `json:"timestamp" validate:"required,datetime=2006-01-02T15:04:05.000Z"`
}
