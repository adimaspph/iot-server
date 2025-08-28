package repository_test_test

import (
	"context"
	"database/sql"
	"errors"
	"iot-server/internal/repository"
	"regexp"
	"testing"
	"time"

	"iot-server/internal/entity"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirupsen/logrus"
)

func newRepoWithDB(t *testing.T) (*repository.SensorRepository, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	return repository.NewSensorRepository(db, logrus.New()), mock, db
}

func sensorRepoBeginTx(t *testing.T) (*repository.SensorRepository, sqlmock.Sqlmock, *sql.DB, *sql.Tx) {
	t.Helper()
	repo, mock, db := newRepoWithDB(t)
	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("db.Begin: %v", err)
	}
	return repo, mock, db, tx
}

//
// CreateTx
//

func TestSensorRepository_CreateTx_Success(t *testing.T) {
	repo, mock, db, tx := sensorRepoBeginTx(t)
	defer db.Close()
	defer func() {
		mock.ExpectRollback()
		_ = tx.Rollback()
	}()

	s := &entity.Sensor{ID1: "S1", ID2: 2, SensorType: "temp"}

	query := regexp.QuoteMeta(`
        INSERT INTO sensors (id1, id2, sensor_type)
        VALUES (?, ?, ?)
    `)
	mock.ExpectExec(query).
		WithArgs(s.ID1, s.ID2, s.SensorType).
		WillReturnResult(sqlmock.NewResult(1234, 1))

	err := repo.CreateTx(context.Background(), tx, s)
	if err != nil {
		t.Fatalf("CreateTx: %v", err)
	}
	if s.SensorID != 1234 {
		t.Fatalf("expected SensorID=1234, got %d", s.SensorID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSensorRepository_CreateTx_InsertError(t *testing.T) {
	repo, mock, db, tx := sensorRepoBeginTx(t)
	defer db.Close()
	defer func() {
		mock.ExpectRollback()
		_ = tx.Rollback()
	}()

	s := &entity.Sensor{ID1: "S1", ID2: 2, SensorType: "temp"}

	query := regexp.QuoteMeta(`
        INSERT INTO sensors (id1, id2, sensor_type)
        VALUES (?, ?, ?)
    `)
	mock.ExpectExec(query).
		WithArgs(s.ID1, s.ID2, s.SensorType).
		WillReturnError(errors.New("insert failed"))

	err := repo.CreateTx(context.Background(), tx, s)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if s.SensorID != 0 {
		t.Fatalf("expected SensorID unchanged, got %d", s.SensorID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSensorRepository_CreateTx_LastInsertIDError(t *testing.T) {
	repo, mock, db, tx := sensorRepoBeginTx(t)
	defer db.Close()
	defer func() {
		mock.ExpectRollback()
		_ = tx.Rollback()
	}()

	s := &entity.Sensor{ID1: "S1", ID2: 2, SensorType: "temp"}

	query := regexp.QuoteMeta(`
        INSERT INTO sensors (id1, id2, sensor_type)
        VALUES (?, ?, ?)
    `)
	mock.ExpectExec(query).
		WithArgs(s.ID1, s.ID2, s.SensorType).
		WillReturnResult(sqlmock.NewErrorResult(errors.New("no last insert id")))

	err := repo.CreateTx(context.Background(), tx, s)
	if err == nil {
		t.Fatalf("expected error from LastInsertId, got nil")
	}
	if s.SensorID != 0 {
		t.Fatalf("expected SensorID unchanged, got %d", s.SensorID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

//
// FindByUnique
//

func TestSensorRepository_FindByUnique_Success(t *testing.T) {
	repo, mock, db := newRepoWithDB(t)
	defer db.Close()

	id1, id2, st := "S1", int64(2), "temp"

	query := regexp.QuoteMeta(`
		SELECT sensor_id, id1, id2, sensor_type
		FROM sensors
		WHERE id1 = ? AND id2 = ? AND sensor_type = ?
		LIMIT 1
	`)
	rows := sqlmock.NewRows([]string{"sensor_id", "id1", "id2", "sensor_type"}).
		AddRow(int64(10), id1, id2, st)
	mock.ExpectQuery(query).WithArgs(id1, id2, st).WillReturnRows(rows)

	got, err := repo.FindByUnique(context.Background(), id1, id2, st)
	if err != nil {
		t.Fatalf("FindByUnique: %v", err)
	}
	if got.SensorID != 10 || got.ID1 != id1 || got.ID2 != id2 || got.SensorType != st {
		t.Fatalf("unexpected sensor: %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSensorRepository_FindByUnique_NotFound(t *testing.T) {
	repo, mock, db := newRepoWithDB(t)
	defer db.Close()

	query := regexp.QuoteMeta(`
		SELECT sensor_id, id1, id2, sensor_type
		FROM sensors
		WHERE id1 = ? AND id2 = ? AND sensor_type = ?
		LIMIT 1
	`)
	mock.ExpectQuery(query).
		WithArgs("S1", int64(2), "temp").
		WillReturnError(sql.ErrNoRows)

	_, err := repo.FindByUnique(context.Background(), "S1", 2, "temp")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

//
// FindSensorRecordsByIdCombination
//

func TestSensorRepository_FindSensorRecordsByIdCombination_Success(t *testing.T) {
	repo, mock, db := newRepoWithDB(t)
	defer db.Close()

	id1, id2 := "S1", int64(2)
	page, pageSize := 2, 3
	offset := (page - 1) * pageSize
	now := time.Now()

	qRecords := regexp.QuoteMeta(`
		SELECT sr.record_id, sr.sensor_id, sr.sensor_value, sr.timestamp, s.id1, s.id2, s.sensor_type
		FROM sensor_records sr
		JOIN sensors s ON s.sensor_id = sr.sensor_id
		WHERE id1 = ? AND id2 = ?
		ORDER BY timestamp ASC
		LIMIT ? OFFSET ?
	`)
	rows := sqlmock.NewRows([]string{"record_id", "sensor_id", "sensor_value", "timestamp", "id1", "id2", "sensor_type"}).
		AddRow(int64(1), int64(10), 11.1, now, id1, id2, "temp").
		AddRow(int64(2), int64(10), 12.2, now.Add(time.Second), id1, id2, "temp")
	mock.ExpectQuery(qRecords).WithArgs(id1, id2, pageSize, offset).WillReturnRows(rows)

	qCount := regexp.QuoteMeta(`
		SELECT COUNT(*)
		FROM sensor_records sr
		JOIN sensors s ON s.sensor_id = sr.sensor_id
		WHERE id1 = ? AND id2 = ?`)
	mock.ExpectQuery(qCount).WithArgs(id1, id2).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(2)))

	s, meta, err := repo.FindSensorRecordsByIdCombination(context.Background(), id1, id2, page, pageSize)
	if err != nil {
		t.Fatalf("FindSensorRecordsByIdCombination: %v", err)
	}
	if s == nil || meta == nil {
		t.Fatalf("expected non-nil sensor & meta")
	}
	if len(s.Records) != 2 || s.ID1 != id1 || s.ID2 != id2 || s.SensorType != "temp" {
		t.Fatalf("unexpected data: %+v", s)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSensorRepository_FindSensorRecordsByIdCombination_QueryError(t *testing.T) {
	repo, mock, db := newRepoWithDB(t)
	defer db.Close()

	qRecords := regexp.QuoteMeta(`
		SELECT sr.record_id, sr.sensor_id, sr.sensor_value, sr.timestamp, s.id1, s.id2, s.sensor_type
		FROM sensor_records sr
		JOIN sensors s ON s.sensor_id = sr.sensor_id
		WHERE id1 = ? AND id2 = ?
		ORDER BY timestamp ASC
		LIMIT ? OFFSET ?
	`)
	mock.ExpectQuery(qRecords).
		WithArgs("S1", int64(2), 10, 0).
		WillReturnError(errors.New("db error"))

	_, _, err := repo.FindSensorRecordsByIdCombination(context.Background(), "S1", 2, 1, 10)
	if err == nil {
		t.Fatalf("expected error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSensorRepository_FindSensorRecordsByIdCombination_ScanError(t *testing.T) {
	repo, mock, db := newRepoWithDB(t)
	defer db.Close()

	qRecords := regexp.QuoteMeta(`
		SELECT sr.record_id, sr.sensor_id, sr.sensor_value, sr.timestamp, s.id1, s.id2, s.sensor_type
		FROM sensor_records sr
		JOIN sensors s ON s.sensor_id = sr.sensor_id
		WHERE id1 = ? AND id2 = ?
		ORDER BY timestamp ASC
		LIMIT ? OFFSET ?
	`)
	// Cause scan error: put string where int is expected (id2)
	rows := sqlmock.NewRows([]string{"record_id", "sensor_id", "sensor_value", "timestamp", "id1", "id2", "sensor_type"}).
		AddRow(int64(1), int64(10), 11.1, time.Now(), "S1", "oops", "temp")
	mock.ExpectQuery(qRecords).WithArgs("S1", int64(2), 5, 0).WillReturnRows(rows)

	_, _, err := repo.FindSensorRecordsByIdCombination(context.Background(), "S1", 2, 1, 5)
	if err == nil {
		t.Fatalf("expected scan error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSensorRepository_FindSensorRecordsByIdCombination_CountError(t *testing.T) {
	repo, mock, db := newRepoWithDB(t)
	defer db.Close()

	id1, id2 := "S1", int64(2)
	qRecords := regexp.QuoteMeta(`
		SELECT sr.record_id, sr.sensor_id, sr.sensor_value, sr.timestamp, s.id1, s.id2, s.sensor_type
		FROM sensor_records sr
		JOIN sensors s ON s.sensor_id = sr.sensor_id
		WHERE id1 = ? AND id2 = ?
		ORDER BY timestamp ASC
		LIMIT ? OFFSET ?
	`)
	mock.ExpectQuery(qRecords).
		WithArgs(id1, id2, 2, 0).
		WillReturnRows(sqlmock.NewRows([]string{"record_id", "sensor_id", "sensor_value", "timestamp", "id1", "id2", "sensor_type"}))

	qCount := regexp.QuoteMeta(`
		SELECT COUNT(*)
		FROM sensor_records sr
		JOIN sensors s ON s.sensor_id = sr.sensor_id
		WHERE id1 = ? AND id2 = ?`)
	mock.ExpectQuery(qCount).WithArgs(id1, id2).WillReturnError(errors.New("count fail"))

	_, _, err := repo.FindSensorRecordsByIdCombination(context.Background(), id1, id2, 1, 2)
	if err == nil {
		t.Fatalf("expected error from count")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

//
// FindSensorRecordsByTimeRange
//

func TestSensorRepository_FindSensorRecordsByTimeRange_Success(t *testing.T) {
	repo, mock, db := newRepoWithDB(t)
	defer db.Close()

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now()
	page, pageSize := 1, 3

	q := regexp.QuoteMeta(`
		SELECT
			r.record_id, r.sensor_id, r.sensor_value, r.timestamp,
			s.sensor_id, s.id1, s.id2, s.sensor_type
		FROM sensor_records r
		JOIN sensors s ON s.sensor_id = r.sensor_id
		WHERE r.timestamp BETWEEN ? AND ?
		ORDER BY r.timestamp ASC
		LIMIT ? OFFSET ?
	`)
	rows := sqlmock.NewRows([]string{
		"r_record_id", "r_sensor_id", "r_sensor_value", "r_timestamp",
		"s_sensor_id", "s_id1", "s_id2", "s_sensor_type",
	}).
		AddRow(int64(1), int64(10), 9.9, start, int64(10), "S1", int64(2), "temp")
	mock.ExpectQuery(q).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), pageSize, 0).
		WillReturnRows(rows)

	qCount := regexp.QuoteMeta(`
		SELECT COUNT(*)
		FROM sensor_records
		WHERE timestamp BETWEEN ? AND ?
	`)
	mock.ExpectQuery(qCount).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	recs, meta, err := repo.FindSensorRecordsByTimeRange(context.Background(), start, end, page, pageSize)
	if err != nil {
		t.Fatalf("FindSensorRecordsByTimeRange: %v", err)
	}
	if len(recs) != 1 || meta == nil {
		t.Fatalf("unexpected result")
	}
	if recs[0].Sensor.ID1 != "S1" || recs[0].Sensor.SensorID != 10 {
		t.Fatalf("unexpected join data: %+v", recs[0].Sensor)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSensorRepository_FindSensorRecordsByTimeRange_CountError(t *testing.T) {
	repo, mock, db := newRepoWithDB(t)
	defer db.Close()

	q := regexp.QuoteMeta(`
		SELECT
			r.record_id, r.sensor_id, r.sensor_value, r.timestamp,
			s.sensor_id, s.id1, s.id2, s.sensor_type
		FROM sensor_records r
		JOIN sensors s ON s.sensor_id = r.sensor_id
		WHERE r.timestamp BETWEEN ? AND ?
		ORDER BY r.timestamp ASC
		LIMIT ? OFFSET ?
	`)
	mock.ExpectQuery(q).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), 5, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"r_record_id", "r_sensor_id", "r_sensor_value", "r_timestamp",
			"s_sensor_id", "s_id1", "s_id2", "s_sensor_type",
		}))

	qCount := regexp.QuoteMeta(`
		SELECT COUNT(*)
		FROM sensor_records
		WHERE timestamp BETWEEN ? AND ?
	`)
	mock.ExpectQuery(qCount).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("count err"))

	_, _, err := repo.FindSensorRecordsByTimeRange(context.Background(), time.Now().Add(-time.Hour), time.Now(), 1, 5)
	if err == nil {
		t.Fatalf("expected error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

//
// FindSensorRecordsByIdAndTimeRange
//

func TestSensorRepository_FindSensorRecordsByIdAndTimeRange_Success(t *testing.T) {
	repo, mock, db := newRepoWithDB(t)
	defer db.Close()

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now()
	id1, id2 := "S1", int64(2)
	page, pageSize := 1, 2

	q := regexp.QuoteMeta(`
		SELECT sr.record_id, sr.sensor_id, sr.sensor_value, sr.timestamp, s.id1, s.id2, s.sensor_type
		FROM sensor_records sr
		JOIN sensors s ON s.sensor_id = sr.sensor_id
		WHERE id1 = ? AND id2 = ?
		AND timestamp BETWEEN ? AND ?
		ORDER BY timestamp ASC
		LIMIT ? OFFSET ?
	`)
	rows := sqlmock.NewRows([]string{"record_id", "sensor_id", "sensor_value", "timestamp", "id1", "id2", "sensor_type"}).
		AddRow(int64(11), int64(99), 1.1, start, id1, id2, "temp")
	mock.ExpectQuery(q).
		WithArgs(id1, id2, sqlmock.AnyArg(), sqlmock.AnyArg(), pageSize, 0).
		WillReturnRows(rows)

	qCount := regexp.QuoteMeta(`
		SELECT COUNT(*)
		FROM sensor_records sr
		JOIN sensors s ON s.sensor_id = sr.sensor_id
		WHERE id1 = ? AND id2 = ?
		AND timestamp BETWEEN ? AND ?
	`)
	mock.ExpectQuery(qCount).WithArgs(id1, id2, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	s, meta, err := repo.FindSensorRecordsByIdAndTimeRange(context.Background(), id1, id2, start, end, page, pageSize)
	if err != nil {
		t.Fatalf("FindSensorRecordsByIdAndTimeRange: %v", err)
	}
	if s == nil || meta == nil || len(s.Records) != 1 {
		t.Fatalf("unexpected result")
	}
	if s.ID1 != id1 || s.ID2 != id2 || s.SensorType != "temp" {
		t.Fatalf("unexpected sensor fields: %+v", s)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

//
// Deletes
//

func TestSensorRepository_DeleteRecordsByIdCombination_Success(t *testing.T) {
	repo, mock, db := newRepoWithDB(t)
	defer db.Close()

	q := regexp.QuoteMeta(`
        DELETE sr
        FROM sensor_records sr
        JOIN sensors s ON s.sensor_id = sr.sensor_id
        WHERE s.id1 = ? AND s.id2 = ?
    `)
	mock.ExpectExec(q).WithArgs("S1", int64(2)).
		WillReturnResult(sqlmock.NewResult(0, 5))

	n, err := repo.DeleteRecordsByIdCombination(context.Background(), "S1", 2)
	if err != nil {
		t.Fatalf("DeleteRecordsByIdCombination: %v", err)
	}
	if n != 5 {
		t.Fatalf("expected 5 rows affected, got %d", n)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSensorRepository_DeleteRecordsByTimeRange_RowsAffectedError(t *testing.T) {
	repo, mock, db := newRepoWithDB(t)
	defer db.Close()

	q := regexp.QuoteMeta(`
        DELETE FROM sensor_records
        WHERE timestamp BETWEEN ? AND ?
    `)
	mock.ExpectExec(q).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected failed")))

	_, err := repo.DeleteRecordsByTimeRange(context.Background(), time.Now().Add(-time.Hour), time.Now())
	if err == nil {
		t.Fatalf("expected rows affected error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

//
// Updates
//

func TestSensorRepository_UpdateSensorValuesByIdCombination_Success(t *testing.T) {
	repo, mock, db := newRepoWithDB(t)
	defer db.Close()

	q := regexp.QuoteMeta(`
		UPDATE sensor_records sr
		JOIN sensors s ON s.sensor_id = sr.sensor_id
		SET sr.sensor_value = ?
		WHERE s.id1 = ? AND s.id2 = ?`)
	mock.ExpectExec(q).
		WithArgs(12.34, "S1", int64(2)).
		WillReturnResult(sqlmock.NewResult(0, 3))

	n, err := repo.UpdateSensorValuesByIdCombination(context.Background(), "S1", 2, 12.34)
	if err != nil {
		t.Fatalf("UpdateSensorValuesByIdCombination: %v", err)
	}
	if n != 3 {
		t.Fatalf("expected 3 rows affected, got %d", n)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSensorRepository_UpdateSensorValuesByTimeRange_ExecError(t *testing.T) {
	repo, mock, db := newRepoWithDB(t)
	defer db.Close()

	q := regexp.QuoteMeta(`
		UPDATE sensor_records
		SET sensor_value = ?
		WHERE timestamp BETWEEN ? AND ?
	`)
	mock.ExpectExec(q).
		WithArgs(9.99, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("update failed"))

	_, err := repo.UpdateSensorValuesByTimeRange(context.Background(), time.Now().Add(-time.Hour), time.Now(), 9.99)
	if err == nil {
		t.Fatalf("expected error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSensorRepository_UpdateSensorValueByIdAndTimeRange_RowsAffectedError(t *testing.T) {
	repo, mock, db := newRepoWithDB(t)
	defer db.Close()

	q := regexp.QuoteMeta(`
		UPDATE sensor_records sr
		JOIN sensors s ON s.sensor_id = sr.sensor_id
		SET sr.sensor_value = ?
		WHERE s.id1 = ? AND s.id2 = ?
		  AND sr.timestamp BETWEEN ? AND ?
	`)
	mock.ExpectExec(q).
		WithArgs(7.77, "S1", int64(2), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected failed")))

	_, err := repo.UpdateSensorValueByIdAndTimeRange(context.Background(), "S1", 2, time.Now().Add(-time.Hour), time.Now(), 7.77)
	if err == nil {
		t.Fatalf("expected rows affected error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
