package entity

// User is a struct that represents a user entity in database
type User struct {
	ID        string
	Password  string
	Name      string
	CreatedAt int64
	UpdatedAt int64
}
