package repository

import (
	"context"
	"database/sql"
	"iot-server/internal/entity"
	"time"

	"github.com/sirupsen/logrus"
)

type UserRepository struct {
	DB  *sql.DB
	Log *logrus.Logger
}

func NewUserRepository(db *sql.DB, log *logrus.Logger) *UserRepository {
	return &UserRepository{
		DB:  db,
		Log: log,
	}
}

// Create inserts a new user row
func (r *UserRepository) Create(ctx context.Context, u *entity.User) error {
	ctx, cancel := ctxWithTimeout(ctx)
	defer cancel()

	now := time.Now().UnixMilli()
	if u.CreatedAt == 0 {
		u.CreatedAt = now
	}
	if u.UpdatedAt == 0 {
		u.UpdatedAt = now
	}

	const q = `
		INSERT INTO users (id, name, password, role, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := r.DB.ExecContext(ctx, q,
		u.ID,
		u.Name,
		u.Password,
		u.Role,
		u.CreatedAt,
		u.UpdatedAt,
	)
	return err
}

// FindByID returns a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	ctx, cancel := ctxWithTimeout(ctx)
	defer cancel()

	const q = `
		SELECT id, name, password, role, created_at, updated_at
		FROM users
		WHERE id = ?
		LIMIT 1
	`
	var user entity.User
	err := r.DB.QueryRowContext(ctx, q, id).Scan(
		&user.ID, &user.Name, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update overwrites name/password/token for a user by ID and bumps updated_at.
// Returns the number of affected rows.
func (r *UserRepository) Update(ctx context.Context, user *entity.User) (int64, error) {
	ctx, cancel := ctxWithTimeout(ctx)
	defer cancel()

	now := time.Now().UnixMilli()
	user.UpdatedAt = now

	const q = `
		UPDATE users
		SET name = ?, password = ?, updated_at = ?
		WHERE id = ?
	`
	res, err := r.DB.ExecContext(ctx, q,
		user.Name,
		user.Password,
		user.UpdatedAt,
		user.ID,
	)
	if err != nil {
		return 0, err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return aff, nil
}

// UpdateToken updates a user token
func (r *UserRepository) UpdateToken(ctx context.Context, id, token string) (int64, error) {
	ctx, cancel := ctxWithTimeout(ctx)
	defer cancel()

	now := time.Now().UnixMilli()
	const q = `
		UPDATE users
		SET token = ?, updated_at = ?
		WHERE id = ?
	`
	res, err := r.DB.ExecContext(ctx, q, token, now, id)

	if err != nil {
		return 0, err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return aff, nil
}
