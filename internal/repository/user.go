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
	Token     string `db:"token"`
}

type UserRepository interface {
	Add(ctx context.Context, user User) error
	GetUserByID(ctx context.Context, userID string) (*User, error)
	GetUsersByFirstnameAndLastname(ctx context.Context, firstName, lastName string) ([]User, error)
	UpdateUserToken(ctx context.Context, userID, token string) error
	GetUserByToken(ctx context.Context, token string) (*User, error)
	AddFriend(ctx context.Context, userID, friendID string) error
	DeleteFriend(ctx context.Context, userID, friendID string) error
	GetUsersIDsByFriendID(ctx context.Context, friendID string) ([]string, error)
	GetPopularFriendsIDsByUserID(ctx context.Context, userID string, popularFriendUsersCount int) ([]string, error)
	GetUsersCountByFriendID(ctx context.Context, friendID string) (int, error)
}
