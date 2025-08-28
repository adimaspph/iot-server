package entity

type Sensor struct {
	SensorID   int64  `json:"sensor_id" gorm:"column:sensor_id;primaryKey;autoIncrement"`
	ID1        string `json:"id1" gorm:"column:id1;size:20;not null"`
	ID2        int64  `json:"id2" gorm:"column:id2;not null"`
	SensorType string `json:"sensor_type" gorm:"column:sensor_type;size:50;not null"`

	Records []SensorRecord `json:"records,omitempty" gorm:"foreignKey:sensor_id;references:sensor_id"`

	// Unique constraint across id1, id2
}

func (Sensor) TableName() string {
	return "sensors"
}
