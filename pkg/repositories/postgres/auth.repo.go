package postgres

import (
	"database/sql"
	"fmt"

	"github.com/slim-crown/issue-1-REST/pkg/services/auth"
)

type jWtAuthRepository repository

// NewAuthRepository returns a struct that implements the auth.Repository using
// a PostgresSQL database.
// A database connection needs to be passed so that it can function.
func NewAuthRepository(DB *sql.DB, allRepos *map[string]interface{}) auth.Repository {
	return &jWtAuthRepository{DB, allRepos}
}

// Authenticate checks the given pass hash against the pass hash found in the database for the username.
func (repo *jWtAuthRepository) Authenticate(u *auth.User) (bool, error) {
	/*	var hashedPassword string
		{
			// this block hashes the password
			cat := u.Password + u.Username
			hashedPasswordArr := sha512.Sum512([]byte(cat))
			hashedPassword = hex.EncodeToString(hashedPasswordArr[:])
		}
	*/
	query := `
SELECT EXISTS(
               SELECT username
               FROM users
               WHERE username = $2
                 AND pass_hash = sha512(
                       ($1 || $2)::bytea
                   )::text
           )
           OR
       EXISTS(
               SELECT username
               FROM users
               WHERE email = $3
                 AND pass_hash = sha512(
                       ($1
                           ||
                        (
                            SELECT username
                            from users
                            where email = $3
                        )
                           )::bytea
                   )::text
           )`
	var accepted bool
	err := repo.db.QueryRow(query, u.Password, u.Username, u.Email).Scan(&accepted)
	if err != nil {
		return false, fmt.Errorf("couldn't authenticate user beacause: %w", err)
	}
	return accepted, nil
}

// AddToBlacklist - this method is only implemented in memory cache as of now.
func (repo *jWtAuthRepository) AddToBlacklist(tokenString string) error {
	panic("not implemented")
}

// IsInBlacklist - this method is only implemented in memory cache as of now.
func (repo *jWtAuthRepository) IsInBlacklist(token string) (bool, error) {
	panic("not implemented")
}
