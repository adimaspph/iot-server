package entity

import "time"

type SensorRecord struct {
	RecordID    int64     `json:"record_id" gorm:"column:record_id;primaryKey;autoIncrement"`
	SensorID    int64     `json:"sensor_id" gorm:"column:sensor_id;not null;index"`
	SensorValue float64   `json:"sensor_value" gorm:"column:sensor_value;not null"`
	Timestamp   time.Time `json:"timestamp" gorm:"column:timestamp;not null;precision:6"`

	// Relations
	Sensor Sensor `json:"sensor,omitempty" gorm:"foreignKey:sensor_id;references:sensor_id"`
}

func (SensorRecord) TableName() string {
	return "sensor_records"
}
