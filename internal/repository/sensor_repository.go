package repository

import (
	"errors"
	"iot-subscriber/internal/entity"
	"iot-subscriber/internal/model"

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

func (r *SensorRepository) FindSensorRecordsByIdCombination(
	tx *gorm.DB,
	id1 string,
	id2 int64,
	page, pageSize int,
) (*entity.Sensor, *model.PageMetadata, error) {
	var sensor *entity.Sensor

	offset := (page - 1) * pageSize

	// load sensors and apply pagination for sensor records
	err := tx.
		Where("id1 = ? AND id2 = ?", id1, id2).
		Order("sensor_id ASC").
		Preload("Records", func(db *gorm.DB) *gorm.DB {
			return db.
				Order("timestamp ASC").
				Limit(pageSize).
				Offset(offset)
		}).
		First(&sensor).Error
	if err != nil {
		r.Log.WithError(err).Error("failed to load sensor record with paginated records")
		return nil, nil, err
	}

	// count total record
	var totalItem int64
	err = tx.Model(&entity.SensorRecord{}).
		Where("sensor_id = ?", sensor.SensorID).
		Count(&totalItem).Error
	if err != nil {
		r.Log.WithError(err).Error("failed to count record with paginated records")
		return nil, nil, err
	}

	metadata := &model.PageMetadata{
		Page:      page,
		Size:      pageSize,
		TotalItem: totalItem,
		TotalPage: int64(int((totalItem + int64(pageSize) - 1) / int64(pageSize))),
	}

	return sensor, metadata, nil
}
