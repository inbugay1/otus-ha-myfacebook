package sqlx

import (
	"context"
	"fmt"

	"myfacebook/internal/db"
	"myfacebook/internal/repository"
)

type UserRepository struct {
	db *db.DB
}

func NewUserRepository(db *db.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) Add(ctx context.Context, user repository.User) error {
	dbConn := r.db.GetConnection()

	sqlQuery := `INSERT INTO users (id, first_name, last_name, birthdate, biography, city, password) 
				VALUES (:id, :first_name, :last_name, :birthdate, :biography, :city, :password)`

	_, err := dbConn.NamedExecContext(ctx, sqlQuery, user)
	if err != nil {
		return fmt.Errorf("failed to add user to db: %w", err)
	}

	return nil
}
