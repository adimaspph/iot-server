package repository

import (
	"iot-subscriber/internal/entity"

	"github.com/sirupsen/logrus"
)

type SensorRecordRepository struct {
	Repository[entity.SensorRecord]
	Log *logrus.Logger
}

func NewSensorRecordRepository(log *logrus.Logger) *SensorRecordRepository {
	return &SensorRecordRepository{
		Log: log,
	}
}
