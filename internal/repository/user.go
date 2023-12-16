package repository

import "context"

type User struct {
	ID        string `db:"id"`
	FirstName string `db:"first_name"`
	LastName  string `db:"last_name"`
	BirthDate string `db:"birthdate"`
	Biography string `db:"biography"`
	City      string `db:"city"`
	Password  string `db:"password"`
}

type UserRepository interface {
	Add(ctx context.Context, user User) error
}
