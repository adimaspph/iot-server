package repository

import (
	"context"
	"database/sql"
	"errors"
	"iot-subscriber/internal/entity"
	"iot-subscriber/internal/model"
	"time"

	"github.com/sirupsen/logrus"
)

const defaultQueryTimeout = 5 * time.Second

type SensorRepository struct {
	DB  *sql.DB
	Log *logrus.Logger
}

func NewSensorRepository(db *sql.DB, log *logrus.Logger) *SensorRepository {
	return &SensorRepository{
		DB:  db,
		Log: log,
	}
}

// helper
func pageMeta(page, pageSize int, total int64) *model.PageMetadata {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	totalPage := (total + int64(pageSize) - 1) / int64(pageSize)
	return &model.PageMetadata{
		Page:      page,
		Size:      pageSize,
		TotalItem: total,
		TotalPage: totalPage,
	}
}

func ctxWithTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithTimeout(parent, defaultQueryTimeout)
}

// Create
func (r *SensorRepository) CreateTx(ctx context.Context, tx *sql.Tx, sensor *entity.Sensor) error {
	const q = `
        INSERT INTO sensors (id1, id2, sensor_type)
        VALUES (?, ?, ?)
    `
	res, err := tx.ExecContext(ctx, q, sensor.ID1, sensor.ID2, sensor.SensorType)
	if err != nil {
		r.Log.WithError(err).Error("failed to insert sensor")
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		r.Log.WithError(err).Error("failed to get last insert id for sensor")
		return err
	}

	sensor.SensorID = id
	return nil
}

// Queries

func (r *SensorRepository) FindByUnique(ctx context.Context, id1 string, id2 int64, sensorType string) (*entity.Sensor, error) {
	ctx, cancel := ctxWithTimeout(ctx)
	defer cancel()

	const q = `
		SELECT sensor_id, id1, id2, sensor_type
		FROM sensors
		WHERE id1 = ? AND id2 = ? AND sensor_type = ?
		LIMIT 1
	`
	var s entity.Sensor
	err := r.DB.QueryRowContext(ctx, q, id1, id2, sensorType).Scan(
		&s.SensorID, &s.ID1, &s.ID2, &s.SensorType,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		r.Log.WithError(err).Errorf("failed to find sensor: id1=%s id2=%d sensorType=%s", id1, id2, sensorType)
		return nil, err
	}
	return &s, nil
}

func (r *SensorRepository) FindSensorRecordsByIdCombination(
	ctx context.Context,
	id1 string,
	id2 int64,
	page, pageSize int,
) (*entity.Sensor, *model.PageMetadata, error) {
	ctx, cancel := ctxWithTimeout(ctx)
	defer cancel()

	var s entity.Sensor

	offset := (page - 1) * pageSize

	// Query records + join sensor
	const qRecords = `
		SELECT sr.record_id, sr.sensor_id, sr.sensor_value, sr.timestamp, s.id1, s.id2, s.sensor_type
		FROM sensor_records sr
		JOIN sensors s ON s.sensor_id = sr.sensor_id
		WHERE id1 = ? AND id2 = ?
		ORDER BY timestamp ASC
		LIMIT ? OFFSET ?
	`
	rows, err := r.DB.QueryContext(ctx, qRecords, id1, id2, pageSize, offset)
	if err != nil {
		r.Log.WithError(err).Error("failed to retrieve sensor records")
		return nil, nil, err
	}
	defer rows.Close()

	s.Records = make([]entity.SensorRecord, 0, pageSize)
	for rows.Next() {
		var rec entity.SensorRecord
		if err := rows.Scan(&rec.RecordID, &rec.SensorID, &rec.SensorValue, &rec.Timestamp, &s.ID1, &s.ID2, &s.SensorType); err != nil {
			r.Log.WithError(err).Error("failed to scan sensor record row")
			return nil, nil, err
		}
		s.Records = append(s.Records, rec)
	}
	err = rows.Err()
	if err != nil {
		r.Log.WithError(err).Error("row iteration error for sensor records")
		return nil, nil, err
	}

	// Count total record
	const qCount = `
		SELECT COUNT(*)
		FROM sensor_records sr
		JOIN sensors s ON s.sensor_id = sr.sensor_id
		WHERE id1 = ? AND id2 = ?`
	var total int64
	err = r.DB.QueryRowContext(ctx, qCount, id1, id2).Scan(&total)
	if err != nil {
		r.Log.WithError(err).Error("failed to count sensor records")
		return nil, nil, err
	}

	return &s, pageMeta(page, pageSize, total), nil
}

func (r *SensorRepository) FindSensorRecordsByTimeRange(
	ctx context.Context,
	startTime, endTime time.Time,
	page, pageSize int,
) ([]entity.SensorRecord, *model.PageMetadata, error) {
	ctx, cancel := ctxWithTimeout(ctx)
	defer cancel()

	offset := (page - 1) * pageSize

	// Query records + join sensor
	const q = `
		SELECT
			r.record_id, r.sensor_id, r.sensor_value, r.timestamp,
			s.sensor_id, s.id1, s.id2, s.sensor_type
		FROM sensor_records r
		JOIN sensors s ON s.sensor_id = r.sensor_id
		WHERE r.timestamp BETWEEN ? AND ?
		ORDER BY r.timestamp ASC
		LIMIT ? OFFSET ?
	`
	rows, err := r.DB.QueryContext(ctx, q, startTime, endTime, pageSize, offset)
	if err != nil {
		r.Log.WithError(err).Errorf("failed to retrieve sensors with records by time range")
		return nil, nil, err
	}
	defer rows.Close()

	out := make([]entity.SensorRecord, 0, pageSize)
	for rows.Next() {
		var rec entity.SensorRecord
		var sens entity.Sensor
		if err := rows.Scan(
			&rec.RecordID, &rec.SensorID, &rec.SensorValue, &rec.Timestamp, &sens.SensorID, &sens.ID1, &sens.ID2, &sens.SensorType,
		); err != nil {
			r.Log.WithError(err).Error("failed to scan time-range row")
			return nil, nil, err
		}
		rec.Sensor = sens
		out = append(out, rec)
	}
	err = rows.Err()
	if err != nil {
		r.Log.WithError(err).Error("row iteration error for time-range records")
		return nil, nil, err
	}

	// Count total record
	const qCount = `
		SELECT COUNT(*)
		FROM sensor_records
		WHERE timestamp BETWEEN ? AND ?
	`
	var total int64
	err = r.DB.QueryRowContext(ctx, qCount, startTime, endTime).Scan(&total)
	if err != nil {
		r.Log.WithError(err).Error("failed to count sensors by time range")
		return nil, nil, err
	}

	return out, pageMeta(page, pageSize, total), nil
}

func (r *SensorRepository) FindSensorRecordsByIdAndTimeRange(
	ctx context.Context,
	id1 string,
	id2 int64,
	startTime, endTime time.Time,
	page, pageSize int,
) (*entity.Sensor, *model.PageMetadata, error) {
	ctx, cancel := ctxWithTimeout(ctx)
	defer cancel()

	var s entity.Sensor

	offset := (page - 1) * pageSize

	// Query records + join sensor
	const qRecords = `
		SELECT sr.record_id, sr.sensor_id, sr.sensor_value, sr.timestamp, s.id1, s.id2, s.sensor_type
		FROM sensor_records sr
		JOIN sensors s ON s.sensor_id = sr.sensor_id
		WHERE id1 = ? AND id2 = ?
		AND timestamp BETWEEN ? AND ?
		ORDER BY timestamp ASC
		LIMIT ? OFFSET ?
	`
	result, err := r.DB.QueryContext(ctx, qRecords, id1, id2, startTime, endTime, pageSize, offset)
	if err != nil {
		r.Log.WithError(err).Error("failed to retrieve records for id+time range")
		return nil, nil, err
	}
	defer result.Close()

	s.Records = make([]entity.SensorRecord, 0, pageSize)
	for result.Next() {
		var rec entity.SensorRecord
		err := result.Scan(&rec.RecordID, &rec.SensorID, &rec.SensorValue, &rec.Timestamp, &s.ID1, &s.ID2, &s.SensorType)
		if err != nil {
			r.Log.WithError(err).Error("failed to scan id+time range row")
			return nil, nil, err
		}
		s.Records = append(s.Records, rec)
	}
	err = result.Err()
	if err != nil {
		r.Log.WithError(err).Error("row iteration error for id+time range")
		return nil, nil, err
	}

	// Count total record
	const qCount = `
		SELECT COUNT(*)
		FROM sensor_records sr
		JOIN sensors s ON s.sensor_id = sr.sensor_id
		WHERE id1 = ? AND id2 = ?
		AND timestamp BETWEEN ? AND ?
	`
	var total int64
	err = r.DB.QueryRowContext(ctx, qCount, id1, id2, startTime, endTime).Scan(&total)
	if err != nil {
		r.Log.WithError(err).Error("failed to count records for id+time range")
		return nil, nil, err
	}

	return &s, pageMeta(page, pageSize, total), nil
}

func (r *SensorRepository) DeleteRecordsByIdCombination(
	ctx context.Context,
	id1 string,
	id2 int64,
) (int64, error) {
	ctx, cancel := ctxWithTimeout(ctx)
	defer cancel()

	const q = `
        DELETE sr
        FROM sensor_records sr
        JOIN sensors s ON s.sensor_id = sr.sensor_id
        WHERE s.id1 = ? AND s.id2 = ?
    `
	res, err := r.DB.ExecContext(ctx, q, id1, id2)
	if err != nil {
		r.Log.WithError(err).Error("failed to delete records by id1+id2")
		return 0, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return affected, nil
}
