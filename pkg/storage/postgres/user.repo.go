package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/slim-crown/Issue-1/pkg/domain/user"
)

// NewUserRepository returns a new in PostgreSQL implementation of user.Repository.
// the databse connection must be passed as the first argument
// since for the repo to work.
// A map of all the other PostgreSQL based implementations of the Repository interfaces
// found in the different services of the project must be passed as a second argument as
// the Repository might make user of them to fetch objects instead of implementing redundant logic.
func NewUserRepository(DB *sql.DB, allRepos *map[string]interface{}) *UserRepository {
	return &UserRepository{DB, allRepos}
}

// AddUser takes in a user.User struct and persists it in the database.
func (repo *UserRepository) AddUser(u *user.User) error {
	var err error
	_, err = repo.db.Exec(`INSERT INTO "issue#1".users (username, email, pass_hash)
							VALUES ($1, $2, $3)`, u.Username, u.Email, u.PassHash)
	if err != nil {
		return fmt.Errorf("insertion of user failed because of: %s", err.Error())
	}

	// set the username to zero to avoid call to UpdateUser won't do redudnant updating of username
	username := u.Username
	u.Username = ""

	// using UpdateUser to set the rest of the values so that null values will be preserved
	// (instead of columns with go's zero value of "")
	err = repo.UpdateUser(username, u)
	if err != nil {
		return fmt.Errorf("the user %s was created with the email %s but some data wasn't persisted because of: %s", u.Username, u.Email, err.Error())
	}
	return nil
}

// GetUser retrives a user.User based on the username passed.
func (repo *UserRepository) GetUser(username string) (*user.User, error) {
	var err error
	var u *user.User = new(user.User)

	var creationTimeString string
	err = repo.db.QueryRow(`SELECT email, COALESCE(first_name, ''), COALESCE(middle_name, ''), COALESCE(last_name, ''), creation_time
							FROM "issue#1".users
							WHERE username = $1`, username).Scan(&u.Email, &u.FirstName, &u.MiddleName, &u.LastName, &creationTimeString)
	if err != nil {
		return nil, fmt.Errorf("scanning Row from users failed because of: %s", err.Error())
	}

	bio, err := repo.getBio(username)
	if err != nil {
		return nil, err
	}
	u.Bio = bio

	creationTime, err := time.Parse(time.RFC3339, creationTimeString)
	if err != nil {
		return nil, fmt.Errorf("parsing of timestamp to time.Time failed because of: %s", err.Error())
	}
	u.CreationTime = creationTime

	bookmarkedPosts, err := repo.getBookmarkedPosts(username)
	if err != nil {
		return nil, fmt.Errorf("unable to get bookmarked posts because of: %s", err.Error())
	}
	u.BookmarkedPosts = bookmarkedPosts

	u.Username = username
	return u, nil
}

// getBookmarkedPosts is just a helper function
func (repo *UserRepository) getBio(username string) (string, error) {
	var bio string
	err := repo.db.QueryRow(`SELECT COALESCE(bio, '') 
							FROM "issue#1".users_bio 
							WHERE username = $1`, username).Scan(&bio)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("scanning Row from bio_users failed because of: %s", err.Error())

	}
	return bio, nil
}

// getBookmarkedPosts is just a helper function
func (repo *UserRepository) getBookmarkedPosts(username string) (map[int]time.Time, error) {
	// TODO test this method
	var bookmarkedPosts = make(map[int]time.Time)
	var (
		postID             int
		creationTimeString string
	)

	rows, err := repo.db.Query(`SELECT post_id, creation_time 
								FROM "issue#1".user_bookmarks 
								WHERE username = $1`, username)
	if err != nil {
		return nil, fmt.Errorf("querying for user_bookmarks failed because of: %s", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&postID, &creationTimeString)
		if err != nil {
			return nil, fmt.Errorf("scanning from rows failed because: %s", err.Error())
		}
		creationTime, err := time.Parse(time.RFC3339, creationTimeString)
		bookmarkedPosts[postID] = creationTime
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("scanning from rows faulty because: %s", err.Error())
	}
	return bookmarkedPosts, nil
}

// UpdateUser updates a user based on the passed user.User struct.
// If updating in the DB repo is successful, it updates its cache by getting
// the new user.User and converting it into a cachable format.
func (repo *UserRepository) UpdateUser(username string, u *user.User) error {
	var err error

	// if err != nil {
	// 	return fmt.Errorf("creation of update statment faild because of: %s", err.Error())
	// }
	// Checks if value is to be updated before attempting.
	// This way, there won't be columns with go's zero string value of "" instead of null
	if u.Username != "" {
		err = repo.execUpdateStatmentOnColumn("username", u.Username, username)
		if err != nil {
			return err
		}
		// change username for subsequent calls if username changed
		username = u.Username
	}
	if u.PassHash != "" {
		err = repo.execUpdateStatmentOnColumn("pass_hash", u.PassHash, username)
		if err != nil {
			return err
		}
	}
	if u.Email != "" {
		err = repo.execUpdateStatmentOnColumn("email", u.Email, username)
		if err != nil {
			return err
		}
	}
	if u.FirstName != "" {
		err = repo.execUpdateStatmentOnColumn("first_name", u.FirstName, username)
		if err != nil {
			return err
		}
	}
	if u.MiddleName != "" {
		err = repo.execUpdateStatmentOnColumn("middle_name", u.MiddleName, username)
		if err != nil {
			return err
		}
	}
	if u.LastName != "" {
		err = repo.execUpdateStatmentOnColumn("last_name", u.LastName, username)
		if err != nil {
			return err
		}
	}
	if u.Bio != "" {
		_, err = repo.db.Exec(`INSERT INTO "issue#1".users_bio(bio, username)
								VALUES ($1, $2)
								ON CONFLICT(username) DO UPDATE
								SET bio = $1`, u.Bio, username)
		if err != nil {
			return fmt.Errorf("upsertion of bio failed because of: %s", err.Error())
		}
	}
	return nil
}

// execUpdateStatmentOnColumn is just a helper function
func (repo *UserRepository) execUpdateStatmentOnColumn(column, value, username string) error {
	_, err := repo.db.Exec(fmt.Sprintf(`UPDATE "issue#1".users 
									SET %s = $1 
									WHERE username = $2`, column), value, username)
	if err != nil {
		return fmt.Errorf("updating failed of %s column with %s because of: %s", column, value, err.Error())
	}
	return nil
}

// DeleteUser deletes a user based on the passed in username.
// If deletion is successful, it also tries to delete the user from its cache.
func (repo *UserRepository) DeleteUser(username string) error {
	// TODO
	_, err := repo.db.Exec(`DELETE FROM "issue#1".users
							WHERE username = $1`, username)
	if err != nil {
		return fmt.Errorf("deletion of user failed because of: %s", err.Error())
	}
	return nil
}

// TODO sort by creation_time
// TODO sort by first_name
// TODO sort by username
// TODO sort by middle_name

// SearchUser searches for users according to the pattern.
// If no pattern is provided, it returns all users.
// It makes use of pagination.
func (repo *UserRepository) SearchUser(pattern, sortBy, sortOrder string, limit, offset int) ([]*user.User, error) {
	var users = make([]*user.User, 0)
	var err error
	var rows *sql.Rows
	if pattern == "" {
		rows, err = repo.db.Query(fmt.Sprintf(`(SELECT username,email, COALESCE(first_name, ''), COALESCE(middle_name, ''), COALESCE(last_name, ''), creation_time 
												FROM "issue#1".users) 
												ORDER BY %s %s NULLS LAST
												LIMIT $1 OFFSET $2`, sortBy, sortOrder), limit, offset)
		if err != nil {
			return nil, fmt.Errorf("querying for users failed because of: %s", err.Error())
		}
		defer rows.Close()
	} else {
		// TODO actual search queries
	}

	var creationTimeString string
	for rows.Next() {
		u := user.User{}
		err := rows.Scan(&u.Username, &u.Email, &u.FirstName, &u.MiddleName, &u.LastName, &creationTimeString)
		if err != nil {
			return nil, fmt.Errorf("scanning from rows failed because: %s", err.Error())
		}
		creationTime, err := time.Parse(time.RFC3339, creationTimeString)
		if err != nil {
			return nil, fmt.Errorf("parsing of timestamp to time.Time failed because of: %s", err.Error())
		}
		u.CreationTime = creationTime

		bio, err := repo.getBio(u.Username)
		if err != nil {
			return nil, err
		}
		u.Bio = bio

		bookmarkedPosts, err := repo.getBookmarkedPosts(u.Username)
		if err != nil {
			return nil, fmt.Errorf("unable to get bookmarked posts because of: %s", err.Error())
		}
		u.BookmarkedPosts = bookmarkedPosts
		users = append(users, &u)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("scanning from rows faulty because: %s", err.Error())
	}
	return users, nil
}

// PassHashIsCorrect checks the given pass hash agains the pass hash found in the database for the username.
func (repo *UserRepository) PassHashIsCorrect(username, passHash string) bool {
	var temp string
	err := repo.db.QueryRow(`SELECT username FROM "issue#1".users
							WHERE username = $1 AND pass_hash = $2`, username, passHash).Scan(&temp)
	if err != nil {
		return false
	}
	return false
}

// BookmarkPost bookmarks the given postID for the user of the given username.
func (repo *UserRepository) BookmarkPost(username string, postID int) error {
	_, err := repo.db.Exec(`INSERT INTO "issue#1".user_bookmarks (username, post_id)
							VALUES ($1, $2)`, username, postID)
	if err != nil {
		return fmt.Errorf("bookmarking of post failed because of: %s", err.Error())
	}
	return nil
}
