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
	ID1         string    `json:"id1" validate:"required,uppercase"`
	ID2         int64     `json:"id2" validate:"required"`
	SensorType  string    `json:"sensor_type" validate:"required"`
	SensorValue float64   `json:"sensor_value" validate:"required"`
	Timestamp   time.Time `json:"timestamp" validate:"required"`
}

type SensorSearchByIdRequest struct {
	ID1      string `query:"id1" validate:"required,uppercase"`
	ID2      int64  `query:"id2" validate:"required"`
	Page     int    `query:"page" validate:"omitempty,min=1"`             // optional, must be >= 1 if provided
	PageSize int    `query:"pageSize" validate:"omitempty,min=1,max=100"` // optional, must be between 1–100
}

type SensorSearchByTimeRangeRequest struct {
	Start    time.Time `query:"start" validate:"required"`
	End      time.Time `query:"end" validate:"required"`
	Page     int       `query:"page" validate:"omitempty,min=1"`             // optional, must be >= 1 if provided
	PageSize int       `query:"pageSize" validate:"omitempty,min=1,max=100"` // optional, must be between 1–100
}

type SensorSearchByIdAndTimeRangeRequest struct {
	ID1      string    `query:"id1" validate:"required,uppercase"`
	ID2      int64     `query:"id2" validate:"required"`
	Start    time.Time `query:"start" validate:"required"`
	End      time.Time `query:"end" validate:"required"`
	Page     int       `query:"page" validate:"omitempty,min=1"`             // optional, must be >= 1 if provided
	PageSize int       `query:"pageSize" validate:"omitempty,min=1,max=100"` // optional, must be between 1–100
}

type SensorDeleteResponse struct {
	Deleted int64 `json:"deleted"`
}

type SensorUpdateResponse struct {
	Updated int64 `json:"updated"`
}

type SensorUpdateByIdRequest struct {
	ID1         string  `json:"id1" validate:"required,uppercase"`
	ID2         int64   `json:"id2" validate:"required"`
	SensorValue float64 `json:"sensor_value" validate:"required"`
}

type SensorUpdateByTimeRangeRequest struct {
	Start       time.Time `json:"start" validate:"required"`
	End         time.Time `json:"end" validate:"required"`
	SensorValue float64   `json:"sensor_value" validate:"required"`
}

type SensorUpdateByIdAndTimeRangeRequest struct {
	ID1         string    `json:"id1" validate:"required,uppercase"`
	ID2         int64     `json:"id2" validate:"required"`
	Start       time.Time `json:"start" validate:"required"`
	End         time.Time `json:"end" validate:"required"`
	SensorValue float64   `json:"sensor_value" validate:"required"`
}
