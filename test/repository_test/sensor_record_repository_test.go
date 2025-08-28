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

func beginTx(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *sql.Tx) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("db.Begin: %v", err)
	}
	return db, mock, tx
}

func TestSensorRecordRepository_CreateTx_Success(t *testing.T) {
	db, mock, tx := beginTx(t)
	defer db.Close()
	defer func() {
		mock.ExpectRollback() // test controls the tx lifecycle
		_ = tx.Rollback()
	}()

	repo := repository.NewSensorRecordRepository(logrus.New())
	rec := &entity.SensorRecord{
		SensorID:    123,
		SensorValue: 45.67,
		Timestamp:   time.Now(),
	}

	query := regexp.QuoteMeta(`
        INSERT INTO sensor_records (sensor_id, sensor_value, timestamp)
        VALUES (?, ?, ?)
    `)
	mock.ExpectExec(query).
		// Timestamp may be driver-normalized; be lenient with AnyArg.
		WithArgs(rec.SensorID, rec.SensorValue, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(9876, 1))

	err := repo.CreateTx(context.Background(), tx, rec)
	if err != nil {
		t.Fatalf("CreateTx returned error: %v", err)
	}
	if rec.RecordID != 9876 {
		t.Fatalf("expected RecordID=9876, got %d", rec.RecordID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSensorRecordRepository_CreateTx_InsertError(t *testing.T) {
	db, mock, tx := beginTx(t)
	defer db.Close()
	defer func() {
		mock.ExpectRollback()
		_ = tx.Rollback()
	}()

	repo := repository.NewSensorRecordRepository(logrus.New())
	rec := &entity.SensorRecord{
		SensorID:    42,
		SensorValue: 1.23,
		Timestamp:   time.Now(),
	}

	query := regexp.QuoteMeta(`
        INSERT INTO sensor_records (sensor_id, sensor_value, timestamp)
        VALUES (?, ?, ?)
    `)
	mock.ExpectExec(query).
		WithArgs(rec.SensorID, rec.SensorValue, sqlmock.AnyArg()).
		WillReturnError(errors.New("insert failed"))

	err := repo.CreateTx(context.Background(), tx, rec)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if rec.RecordID != 0 {
		t.Fatalf("expected RecordID unchanged (0), got %d", rec.RecordID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSensorRecordRepository_CreateTx_LastInsertIDError(t *testing.T) {
	db, mock, tx := beginTx(t)
	defer db.Close()
	defer func() {
		mock.ExpectRollback()
		_ = tx.Rollback()
	}()

	repo := repository.NewSensorRecordRepository(logrus.New())
	rec := &entity.SensorRecord{
		SensorID:    7,
		SensorValue: 9.99,
		Timestamp:   time.Now(),
	}

	query := regexp.QuoteMeta(`
        INSERT INTO sensor_records (sensor_id, sensor_value, timestamp)
        VALUES (?, ?, ?)
    `)
	// Simulate driver failing on LastInsertId()
	mock.ExpectExec(query).
		WithArgs(rec.SensorID, rec.SensorValue, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewErrorResult(errors.New("no last insert id")))

	err := repo.CreateTx(context.Background(), tx, rec)
	if err == nil {
		t.Fatalf("expected error from LastInsertId, got nil")
	}
	if rec.RecordID != 0 {
		t.Fatalf("expected RecordID unchanged (0), got %d", rec.RecordID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
