package entity

// User is a struct that represents a user entity in database
type User struct {
	ID        string
	Password  string
	Name      string
	Role      string
	CreatedAt int64
	UpdatedAt int64
}

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)
