package entity

import "time"

type Sensor struct {
	SensorID   int64     `json:"sensor_id" gorm:"column:sensor_id;primaryKey;autoIncrement"`
	ID1        string    `json:"id1" gorm:"column:id1;size:20;not null"`
	ID2        int64     `json:"id2" gorm:"column:id2;not null"`
	SensorType string    `json:"sensor_type" gorm:"column:sensor_type;size:50;not null"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt  time.Time `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`

	Records []SensorRecord `json:"records,omitempty" gorm:"foreignKey:sensor_id;references:sensor_id"`

	// Unique constraint across id1, id2, sensor_type
}

func (Sensor) TableName() string {
	return "sensors"
}
