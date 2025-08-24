package repository

import (
	"errors"
	"iot-subscriber/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SensorRepository struct {
	Repository[entity.Sensor]
	Log *logrus.Logger
}

func NewSensorRepository(log *logrus.Logger) *SensorRepository {
	return &SensorRepository{
		Log: log,
	}
}

func (r *SensorRepository) FindByUnique(tx *gorm.DB, id1 string, id2 int64, sensorType string) (*entity.Sensor, error) {
	var sensor entity.Sensor
	err := tx.Where("id1 = ? AND id2 = ? AND sensor_type = ?", id1, id2, sensorType).First(&sensor).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		r.Log.WithError(err).Errorf("failed to find sensor: id1=%s id2=%d sensorType=%s", id1, id2, sensorType)
		return nil, err
	}
	return &sensor, nil
}
