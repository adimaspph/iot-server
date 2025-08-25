package repository

import (
	"context"
	"database/sql"
	"iot-subscriber/internal/entity"

	"github.com/sirupsen/logrus"
)

type SensorRecordRepository struct {
	Log *logrus.Logger
}

func NewSensorRecordRepository(log *logrus.Logger) *SensorRecordRepository {
	return &SensorRecordRepository{
		Log: log,
	}
}

func (r *SensorRecordRepository) CreateTx(ctx context.Context, tx *sql.Tx, record *entity.SensorRecord) error {
	const q = `
        INSERT INTO sensor_records (sensor_id, sensor_value, timestamp)
        VALUES (?, ?, ?)
    `
	res, err := tx.ExecContext(ctx, q, record.SensorID, record.SensorValue, record.Timestamp)
	if err != nil {
		r.Log.WithError(err).Error("failed to insert sensor record")
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		r.Log.WithError(err).Error("failed to get last insert id for record")
		return err
	}

	record.RecordID = id
	return nil
}
