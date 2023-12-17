package sqlx

import (
	"context"
	"database/sql"
	"errors"
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

func (r *UserRepository) GetUserByID(ctx context.Context, userID string) (*repository.User, error) {
	dbConn := r.db.GetConnection()

	var user repository.User

	sqlQuery := `SELECT id, first_name, last_name, birthdate, city, biography, password, token FROM users WHERE id = $1`

	err := dbConn.GetContext(ctx, &user, sqlQuery, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}

		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) UpdateUserToken(ctx context.Context, userID, token string) error {
	dbConn := r.db.GetConnection()

	sqlQuery := `UPDATE users SET token=$2 WHERE id=$1`

	_, err := dbConn.ExecContext(ctx, sqlQuery, userID, token)
	if err != nil {
		return fmt.Errorf("failed to update user token in db: %w", err)
	}

	return nil
}
