package model

import (
	"database/sql"
	"fmt"
	"log/slog"
)

type User struct {
	// PRIMARY KEY
	ID int
	// XTB user id
	Username int
}

type UserService struct {
	Logger *slog.Logger
	DB     *sql.DB
}

func (us UserService) Insert(userID int) error {
	var exist bool
	row := us.DB.QueryRow(`
			SELECT 
		    	EXISTS (SELECT 1 FROM users WHERE username = $1)
	`, userID)
	err := row.Scan(&exist)
	if err != nil {
		return fmt.Errorf("select user while inserting: %w", err)
	}

	if exist {
		us.Logger.Info(fmt.Sprintf("User with username: %d already exsists in database", userID))
		return nil
	}

	var userDatabaseID int
	row = us.DB.QueryRow(`
	INSERT INTO users 
	    (username)
	    VALUES ($1)
	    RETURNING id`, userID)
	err = row.Scan(&userDatabaseID)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}

	us.Logger.Info(fmt.Sprintf("Inserted user with username: %d, assigned id: %d", userID, userDatabaseID))
	return nil
}

func (us UserService) SelectUserByUsername(userID int) (User, error) {
	var user User
	row := us.DB.QueryRow(`
		SELECT *
		FROM users
		WHERE username = $1`, userID)
	err := row.Scan(&user.ID, &user.Username)
	if err != nil {
		return User{}, fmt.Errorf("select user by username: %w", err)
	}
	return user, nil
}
