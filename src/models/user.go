package models

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"time"
)

type User struct {
	Id                  string    `json:"id" binding:"required"`
	Username            string    `json:"username" binding:"required"`
	Pic                 string    `json:"pic"`
	TwitterAccessToken  string    `json:"-" binding:"required"`
	TwitterAccessSecret string    `json:"-" binding:"required"`
	TimeCreated         time.Time `json:"timeCreated"`
	TimeUpdated         time.Time `json:"timeUpdated"`
	IsAdmin             bool      `json:"isAdmin"`
}

var ADMIN_USER_IDS []string = []string{"14087951"}

func (user *User) CreateTableSQL() string {
	return `
    CREATE TABLE auth_user (
        id TEXT PRIMARY KEY,
        username TEXT NOT NULL,
        pic TEXT NOT NULL,
        twitter_access_token TEXT NOT NULL,
        twitter_access_secret TEXT NOT NULL,
        time_created timestamp NOT NULL,
        time_updated timestamp NOT NULL,
        UNIQUE (id),
        UNIQUE (username)
    );
    `
}

func (user *User) GetIsAdmin() bool {
	for _, adminUserId := range ADMIN_USER_IDS {
		if adminUserId == user.Id {
			return true
		}
	}
	return false
}

// This is needed because we don't want to expose the access token/secret
// normally, but we do want to ingest them, and martini's binding module isn't
// expressive enough to represent that concisely.
type UserForm struct {
	TwitterAccessToken  string `json:"twitterAccessToken" binding:"required"`
	TwitterAccessSecret string `json:"twitterAccessSecret" binding:"required"`
}

// DATABASE ACCESS STUFF

func UserRowReturn(err error, user *User) (*User, error) {
	switch {
	case err != nil:
		return nil, err
	default:
		user.IsAdmin = user.GetIsAdmin()
		return user, nil
	}
}

func (db *DB) GetUserWithId(id string) (*User, error) {
	var user User
	err := db.SQLDB.QueryRow(`
        SELECT id, username, pic, twitter_access_token, twitter_access_secret,
               time_created, time_updated
        FROM auth_user WHERE id = $1`, id).Scan(
		&user.Id,
		&user.Username,
		&user.Pic,
		&user.TwitterAccessToken,
		&user.TwitterAccessSecret,
		&user.TimeCreated,
		&user.TimeUpdated,
	)
	return UserRowReturn(err, &user)
}

func (db *DB) GetUserWithUsername(username string) (*User, error) {
	var user User
	err := db.SQLDB.QueryRow(`
        SELECT id, username, pic, twitter_access_token, twitter_access_secret,
               time_created, time_updated
        FROM auth_user WHERE UPPER(username) = UPPER($1)`, username).Scan(
		&user.Id,
		&user.Username,
		&user.Pic,
		&user.TwitterAccessToken,
		&user.TwitterAccessSecret,
		&user.TimeCreated,
		&user.TimeUpdated,
	)
	return UserRowReturn(err, &user)
}

func (db *DB) UpdateUserAuth(id, twitterAccessToken, twitterAccessSecret string) (*User, error) {
	user, err := db.GetUserWithId(id)
	if err != nil {
		return user, err
	}
	user.TwitterAccessToken = twitterAccessToken
	user.TwitterAccessSecret = twitterAccessSecret
	user.TimeUpdated = time.Now().UTC()
	_, err = db.SQLDB.Exec(`
		UPDATE auth_user
		SET twitter_access_token = $1, twitter_access_secret = $2,
		    time_updated = $3
		WHERE id = $4
		`,
		user.TwitterAccessToken,
		user.TwitterAccessSecret,
		user.TimeUpdated,
		user.Id,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (db *DB) CreateUser(id, username, pic, twitterAccessToken, twitterAccessSecret string) (*User, error) {
	t := time.Now().UTC()
	user := User{
		Id:                  id,
		Username:            username,
		Pic:                 pic,
		TwitterAccessToken:  twitterAccessToken,
		TwitterAccessSecret: twitterAccessSecret,
		TimeCreated:         t,
		TimeUpdated:         t,
	}

	if id != "" {
		tmpUser, tmpErr := db.GetUserWithId(id)
		if tmpErr != nil && tmpErr != sql.ErrNoRows {
			return nil, tmpErr
		}
		if tmpUser != nil {
			reason := fmt.Sprintf("User with id %v already exists.", id)
			return nil, errors.New(reason)
		}
	}
	if username != "" {
		tmpUser, tmpErr := db.GetUserWithUsername(username)
		if tmpErr != nil && tmpErr != sql.ErrNoRows {
			return nil, tmpErr
		}
		if tmpUser != nil {
			reason := fmt.Sprintf("User with username %v already exists.",
				username)
			return nil, errors.New(reason)
		}
	}

	_, err := db.SQLDB.Exec(`
        INSERT INTO auth_user(
            id, username, pic, twitter_access_token, twitter_access_secret,
            time_created, time_updated
        ) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		user.Id,
		user.Username,
		user.Pic,
		user.TwitterAccessToken,
		user.TwitterAccessSecret,
		user.TimeCreated,
		user.TimeUpdated,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
