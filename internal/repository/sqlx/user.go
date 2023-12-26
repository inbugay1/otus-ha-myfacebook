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
	writeDB *db.DB
	readDB  *db.DB
}

func NewUserRepository(writeDB *db.DB, readDB *db.DB) *UserRepository {
	return &UserRepository{
		writeDB: writeDB,
		readDB:  readDB,
	}
}

func (r *UserRepository) Add(ctx context.Context, user repository.User) error {
	dbConn := r.writeDB.GetConnection()

	sqlQuery := `INSERT INTO users (id, first_name, last_name, birthdate, biography, city, password) 
				VALUES (:id, :first_name, :last_name, :birthdate, :biography, :city, :password)`

	_, err := dbConn.NamedExecContext(ctx, sqlQuery, user)
	if err != nil {
		return fmt.Errorf("failed to add user to writeDB: %w", err)
	}

	return nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, userID string) (*repository.User, error) {
	dbConn := r.readDB.GetConnection()

	var user repository.User

	sqlQuery := `SELECT id, first_name, last_name, TO_CHAR(birthdate, 'YYYY-MM-DD') as birthdate, city, biography, password, token, friend_id FROM users WHERE id = $1`

	err := dbConn.GetContext(ctx, &user, sqlQuery, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}

		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetUsersByFirstnameAndLastname(ctx context.Context, firstName, lastName string) ([]repository.User, error) {
	dbConn := r.readDB.GetConnection()

	var users []repository.User

	sqlQuery := `SELECT id, first_name, last_name, TO_CHAR(birthdate, 'YYYY-MM-DD') as birthdate, city, biography, password, token 
		FROM users WHERE first_name LIKE $1 AND last_name LIKE $2 ORDER BY id`

	err := dbConn.SelectContext(ctx, &users, sqlQuery, firstName+"%", lastName+"%")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}

		return nil, fmt.Errorf("failed to find users by firstname and lastname: %w", err)
	}

	return users, nil
}

func (r *UserRepository) UpdateUserToken(ctx context.Context, userID, token string) error {
	dbConn := r.writeDB.GetConnection()

	sqlQuery := `UPDATE users SET token=$2 WHERE id=$1`

	_, err := dbConn.ExecContext(ctx, sqlQuery, userID, token)
	if err != nil {
		return fmt.Errorf("failed to update user token in writeDB: %w", err)
	}

	return nil
}

func (r *UserRepository) SetUserFriend(ctx context.Context, userID, friendID string) error {
	dbConn := r.writeDB.GetConnection()

	sqlQuery := `UPDATE users SET friend_id=$2 WHERE id=$1`

	_, err := dbConn.ExecContext(ctx, sqlQuery, userID, friendID)
	if err != nil {
		return fmt.Errorf("failed to update user friend: %w", err)
	}

	return nil
}

func (r *UserRepository) GetUserByToken(ctx context.Context, token string) (*repository.User, error) {
	dbConn := r.readDB.GetConnection()

	var user repository.User

	sqlQuery := `SELECT id, first_name, last_name, TO_CHAR(birthdate, 'YYYY-MM-DD') as birthdate, city, biography, password, token FROM users WHERE token = $1`

	err := dbConn.GetContext(ctx, &user, sqlQuery, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}

		return nil, fmt.Errorf("failed to get user by token: %w", err)
	}

	return &user, nil
}
