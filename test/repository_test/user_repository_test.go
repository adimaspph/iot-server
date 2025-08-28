package repository_test_test

import (
	"context"
	"database/sql"
	"errors"
	"iot-server/internal/repository"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirupsen/logrus"
	"iot-server/internal/entity"
)

func newMockRepo(t *testing.T) (*repository.UserRepository, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	log := logrus.New()
	return repository.NewUserRepository(db, log), mock, db
}

func TestUserRepository_Create_SetsTimestampsAndInserts(t *testing.T) {
	repo, mock, db := newMockRepo(t)
	defer db.Close()

	u := &entity.User{
		ID:       "u1",
		Name:     "Alice",
		Password: "hash",
		Role:     "admin",
		// CreatedAt & UpdatedAt are 0 -> should be set by repo
	}

	query := regexp.QuoteMeta(`
		INSERT INTO users (id, name, password, role, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	mock.ExpectExec(query).
		WithArgs(u.ID, u.Name, u.Password, u.Role, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Create(context.Background(), u)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if u.CreatedAt == 0 || u.UpdatedAt == 0 {
		t.Fatalf("timestamps not set: created_at=%d updated_at=%d", u.CreatedAt, u.UpdatedAt)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepository_Create_RespectsProvidedTimestamps(t *testing.T) {
	repo, mock, db := newMockRepo(t)
	defer db.Close()

	pre := time.Now().Add(-time.Hour).UnixMilli()
	u := &entity.User{
		ID:        "u2",
		Name:      "Bob",
		Password:  "pw",
		Role:      "user",
		CreatedAt: pre,
		UpdatedAt: pre,
	}

	query := regexp.QuoteMeta(`
		INSERT INTO users (id, name, password, role, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	mock.ExpectExec(query).
		WithArgs(u.ID, u.Name, u.Password, u.Role, u.CreatedAt, u.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.Create(context.Background(), u); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if u.CreatedAt != pre || u.UpdatedAt != pre {
		t.Fatalf("timestamps overwritten")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepository_FindByID_Success(t *testing.T) {
	repo, mock, db := newMockRepo(t)
	defer db.Close()

	id := "u1"
	rows := sqlmock.NewRows([]string{"id", "name", "password", "role", "created_at", "updated_at"}).
		AddRow(id, "Alice", "hash", "admin", int64(1), int64(2))

	query := regexp.QuoteMeta(`
		SELECT id, name, password, role, created_at, updated_at
		FROM users
		WHERE id = ?
		LIMIT 1
	`)
	mock.ExpectQuery(query).WithArgs(id).WillReturnRows(rows)

	got, err := repo.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.ID != id || got.Name != "Alice" || got.Role != "admin" {
		t.Fatalf("unexpected user: %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	repo, mock, db := newMockRepo(t)
	defer db.Close()

	id := "nope"
	query := regexp.QuoteMeta(`
		SELECT id, name, password, role, created_at, updated_at
		FROM users
		WHERE id = ?
		LIMIT 1
	`)
	mock.ExpectQuery(query).WithArgs(id).WillReturnError(sql.ErrNoRows)

	got, err := repo.FindByID(context.Background(), id)
	if err == nil || !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got: %v (user=%v)", err, got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepository_Update_Success(t *testing.T) {
	repo, mock, db := newMockRepo(t)
	defer db.Close()

	u := &entity.User{ID: "u1", Name: "NewName", Password: "newpw"}

	query := regexp.QuoteMeta(`
		UPDATE users
		SET name = ?, password = ?, updated_at = ?
		WHERE id = ?
	`)
	mock.ExpectExec(query).
		WithArgs(u.Name, u.Password, sqlmock.AnyArg(), u.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	affected, err := repo.Update(context.Background(), u)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if affected != 1 {
		t.Fatalf("expected affected=1, got %d", affected)
	}
	if u.UpdatedAt == 0 {
		t.Fatalf("expected UpdatedAt to be set")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepository_Update_NoRows(t *testing.T) {
	repo, mock, db := newMockRepo(t)
	defer db.Close()

	u := &entity.User{ID: "missing", Name: "X", Password: "Y"}
	query := regexp.QuoteMeta(`
		UPDATE users
		SET name = ?, password = ?, updated_at = ?
		WHERE id = ?
	`)
	mock.ExpectExec(query).
		WithArgs(u.Name, u.Password, sqlmock.AnyArg(), u.ID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	affected, err := repo.Update(context.Background(), u)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if affected != 0 {
		t.Fatalf("expected 0 affected, got %d", affected)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepository_Update_RowsAffectedError(t *testing.T) {
	repo, mock, db := newMockRepo(t)
	defer db.Close()

	u := &entity.User{ID: "u1", Name: "N", Password: "P"}
	query := regexp.QuoteMeta(`
		UPDATE users
		SET name = ?, password = ?, updated_at = ?
		WHERE id = ?
	`)
	mock.ExpectExec(query).
		WithArgs(u.Name, u.Password, sqlmock.AnyArg(), u.ID).
		WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected failed")))

	_, err := repo.Update(context.Background(), u)
	if err == nil {
		t.Fatalf("expected error from RowsAffected")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepository_UpdateToken_Success(t *testing.T) {
	repo, mock, db := newMockRepo(t)
	defer db.Close()

	id := "u1"
	token := "tok"

	query := regexp.QuoteMeta(`
		UPDATE users
		SET token = ?, updated_at = ?
		WHERE id = ?
	`)
	mock.ExpectExec(query).
		WithArgs(token, sqlmock.AnyArg(), id).
		WillReturnResult(sqlmock.NewResult(0, 1))

	affected, err := repo.UpdateToken(context.Background(), id, token)
	if err != nil {
		t.Fatalf("UpdateToken: %v", err)
	}
	if affected != 1 {
		t.Fatalf("expected affected=1, got %d", affected)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepository_UpdateToken_RowsAffectedError(t *testing.T) {
	repo, mock, db := newMockRepo(t)
	defer db.Close()

	id := "u1"
	token := "tok"

	query := regexp.QuoteMeta(`
		UPDATE users
		SET token = ?, updated_at = ?
		WHERE id = ?
	`)
	mock.ExpectExec(query).
		WithArgs(token, sqlmock.AnyArg(), id).
		WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected failed")))

	_, err := repo.UpdateToken(context.Background(), id, token)
	if err == nil {
		t.Fatalf("expected error from RowsAffected")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
