package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/slim-crown/issue-1-REST/pkg/domain/feed"
)


// FeedRepository ...
type FeedRepository repository

// NewFeedRepository returns a new in PostgreSQL implementation of user.Repository.
// the databse connection must be passed as the first argument
// since for the repo to work.
func NewFeedRepository(db *sql.DB, allRepos *map[string]interface{}) *FeedRepository {
	return &FeedRepository{db: db, allRepos: allRepos}
}

// AddFeed persists a feed entity to the DB according to the feed.Feed struct passed in.
func (repo *FeedRepository) AddFeed(f *feed.Feed) error {
	var err error
	var sorting string
	switch f.Sorting {
	case feed.SortHot:
		sorting = "hot"
	case feed.SortNew:
		sorting = "new"
	default:
		sorting = "top"
	}
	result, err := repo.db.Exec(`INSERT INTO feeds (owner_username,sorting)
										VALUES ($1, $2)`, f.OwnerUsername, sorting)
	if err != nil {
		return fmt.Errorf("insertion of user failed because of: %s", err.Error())
	}

	id64, err := result.LastInsertId()
	if err == nil {
		return fmt.Errorf("unable to get LastInsertId")
	}
	username := f.OwnerUsername
	f.OwnerUsername = ""
	id := uint(id64)
	err = repo.UpdateFeed(id, f)
	if err != nil {
		return fmt.Errorf("the feed was created for user %s but some data wasn't persisted because of: %s", username, err.Error())
	}
	return nil
}

// GetFeed retrieve the feed entity in the database belonging to the user of the passed
// in username.
func (repo *FeedRepository) GetFeed(username string) (*feed.Feed, error) {
	var err error
	f := feed.Feed{OwnerUsername: username}
	var sorting string
	err = repo.db.QueryRow(`SELECT id, sorting
	 								FROM feeds
	 								WHERE owner_username = $1`, username).Scan(&f.ID, &sorting)
	if err != nil {
		return nil, feed.ErrFeedNotFound //fmt.Errorf("scanning Row from users failed because of: %s", err.Error())
	}
	switch sorting {
	case "hot":
		f.Sorting = feed.SortHot
	case "new":
		f.Sorting = feed.SortNew
	default:
		f.Sorting = feed.SortTop
	}
	return &f, nil
}

// GetChannels retrieves the all the channels the given feed has subscribed to.
func (repo *FeedRepository) GetChannels(f *feed.Feed, sortBy string, sortOrder string) ([]*feed.Channel, error) {
	// TODO test this method
	channelSubscriptions := make([]*feed.Channel, 0)
	var (
		channelname, name, subTimeString string
	)
	rows, err := repo.db.Query(fmt.Sprintf(`
		SELECT username, name, subscription_time
		FROM (
			(
				SELECT channel_username, subscription_time
				FROM feed_subscriptions
				WHERE feed_id = $1
			) AS S (username, subscription_time)
			NATURAL JOIN
			channels
		)
		ORDER BY %s %s NULLS LAST`, sortBy, sortOrder), f.ID)
	if err != nil {
		return nil, fmt.Errorf("querying for feed_subscriptions failed because of: %s", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&channelname, &name, &subTimeString)
		if err != nil {
			return nil, fmt.Errorf("scanning from rows failed because: %s", err.Error())
		}
		subTime, err := time.Parse(time.RFC3339, subTimeString)
		channelSubscriptions = append(channelSubscriptions, &feed.Channel{
			Username:         channelname,
			Name:             name,
			SubscriptionTime: subTime,
		})
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("scanning from rows faulty because: %s", err.Error())
	}
	return channelSubscriptions, nil
}

// GetPosts returns a list of posts collected from the channels
// the given feed has subscribed to sorted according to the given
// method.
func (repo *FeedRepository) GetPosts(f *feed.Feed, sort feed.Sorting, limit, offset int) ([]*feed.Post, error) {
	var err error

	var rows *sql.Rows
	switch sort {
	// TODO test queries with actual posts
	case feed.SortNew:
		rows, err = repo.db.Query(`
		SELECT *
		FROM	(SELECT id, creation_time
				FROM posts
				NATURAL JOIN
				(SELECT username
				FROM (
						(
						SELECT channel_username
						FROM feed_subscriptions
						WHERE feed_id = $1
						) AS S (username)
						NATURAL JOIN
						channels
					)
				) AS C (channel_username)
				) AS P
		ORDER BY creation_time DESC NULLS LAST LIMIT $2 OFFSET $3`, f.ID, limit, offset)
	case feed.SortHot:
		rows, err = repo.db.Query(`
		SELECT post_id
		FROM(SELECT *
			FROM(SELECT *
				FROM (SELECT id, creation_time
						FROM posts
						NATURAL JOIN
							(SELECT username
							FROM (
								(
								SELECT channel_username
								FROM feed_subscriptions
								WHERE feed_id = $1
								) AS S (username)
								NATURAL JOIN
								channels
								)
							) AS C (channel_username)
						) AS P
				ORDER BY creation_time DESC NULLS LAST
				) AS LP (post_id)
				NATURAL JOIN
				(SELECT post_id,COUNT(*)
				 FROM comments
				 GROUP BY post_id
				) AS PS (post_id, comment_count)
			ORDER BY creation_time DESC, comment_count DESC
		) AS F
		LIMIT $2 OFFSET $3`, f.ID, limit, offset)
	case feed.NotSet:
		fallthrough
	case feed.SortTop:
		fallthrough
	default:
		rows, err = repo.db.Query(`
		SELECT post_id
		FROM(SELECT *
			FROM(SELECT *
				FROM (SELECT id, creation_time
						FROM posts
						NATURAL JOIN
							(SELECT username
							FROM (
								(
								SELECT channel_username
								FROM feed_subscriptions
								WHERE feed_id = 1
								) AS S (username)
								NATURAL JOIN
								channels
								)
							) AS C (channel_username)
						) AS P
				ORDER BY creation_time DESC NULLS LAST
				) AS LP (post_id)
				NATURAL JOIN
				(SELECT post_id,SUM(star_count)
				 FROM post_stars
				 GROUP BY post_id
				) AS PS (post_id, total_star_count)
			ORDER BY creation_time DESC, total_star_count DESC
		) AS F
		LIMIT 2 OFFSET 3`, f.ID, limit, offset)
	}
	if err != nil {
		return nil, fmt.Errorf("querying for feed_subscriptions failed because of: %s", err.Error())
	}
	defer rows.Close()

	var id int
	posts := make([]*feed.Post, 0)

	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("scanning from rows failed because: %s", err.Error())
		}
		posts = append(posts, &feed.Post{ID: id})
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("scanning from rows faulty because: %s", err.Error())
	}

	return posts, nil
}

// UpdateFeed updates the feed entity under the given id based on the passed in feed.Feed struct.
func (repo *FeedRepository) UpdateFeed(id uint, f *feed.Feed) error {
	var err error
	var sorting string
	switch f.Sorting {
	case feed.SortHot:
		sorting = "hot"
	case feed.SortNew:
		sorting = "new"
	default:
		sorting = "top"
	}
	if f.OwnerUsername != "" {
		err = repo.execUpdateStatementOnColumn("owner_username", f.OwnerUsername, id)
		if err != nil {
			return err
		}
		// change username for subsequent calls if username changed
	}
	err = repo.execUpdateStatementOnColumn("sorting", sorting, id)
	return nil
}

// execUpdateStatementOnColumn is just a helper function
func (repo *FeedRepository) execUpdateStatementOnColumn(column, value string, id uint) error {
	_, err := repo.db.Exec(fmt.Sprintf(`
			UPDATE feeds 
			SET %s = $1 
			WHERE id = $2`, column), value, id)
	if err != nil {
		return fmt.Errorf("updating failed of %s column with %s because of: %s", column, value, err.Error())
	}
	return nil
}

// Subscribe adds the given channel to the list of channels that the feed collects posts from.
func (repo *FeedRepository) Subscribe(f *feed.Feed, channelname string) error {
	_, err := repo.db.Exec(`
		INSERT INTO feed_subscriptions (feed_id,channel_username)
		VALUES ($1, $2)`, f.ID, channelname)
	const foreignKeyViolationErrorCode = pq.ErrorCode("23503")
	if err != nil {
		if pgErr, isPGErr := err.(pq.Error); !isPGErr {
			if pgErr.Code != foreignKeyViolationErrorCode {
				return feed.ErrChannelDoesNotExist
			}
			return fmt.Errorf("insertion of user failed because of: %s", err.Error())
		}
	}
	return nil
}

// Unsubscribe removes the channel to the list of channels that the feed collects posts from.
func (repo *FeedRepository) Unsubscribe(f *feed.Feed, channelname string) error {
	_, err := repo.db.Exec(`
		DELETE FROM feed_subscriptions
		WHERE feed_id = $1 AND channel_username = $2`, f.ID, channelname)
	if err != nil {
		return fmt.Errorf("deletion of user failed because of: %s", err.Error())
	}
	return nil
}
