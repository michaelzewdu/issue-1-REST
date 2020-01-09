package postgres

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"github.com/slim-crown/issue-1-REST/pkg/domain/channel"
	"github.com/slim-crown/issue-1-REST/pkg/domain/user"

	"time"
)

func NewChannelRepository(DB *sql.DB, allRepos *map[string]interface{}) *ChannelRepository {
	return &ChannelRepository{DB, allRepos}
}

func (repo *ChannelRepository) AddChannel(c *channel.Channel) error {
	var err error

	_, err = repo.db.Exec(`INSERT INTO "issue#1".channels (username, name, description)
							VALUES ($1, $2, $3)`, c.Username, c.Name, c.Description)
	if err != nil {
		return fmt.Errorf("insertion of user failed because of: %s", err.Error())
	}

	return nil
}
func (repo *ChannelRepository) GetChannel(username string) (*channel.Channel, error) {
	var err error
	var c = new(channel.Channel)

	var creationTimeString string
	err = repo.db.QueryRow(`SELECT name, COALESCE(description, ''), creation_time
							FROM "issue#1".channels
							WHERE username = $1`, username).Scan(&c.Name, &c.Description, &creationTimeString)
	if err != nil {
		return nil, channel.ErrChannelNotFound
	}

	creationTime, err := time.Parse(time.RFC3339, creationTimeString)
	if err != nil {
		return nil, fmt.Errorf("parsing of timestamp to time.Time failed because of: %s", err.Error())
	}
	c.CreationTime = creationTime

	admins, err := repo.GetAdmins(username)
	if err != nil {
		return nil, fmt.Errorf("unable to get admins because of: %s", err.Error())
	}
	owner, err := repo.GetOwner(username)
	if err != nil {
		return nil, fmt.Errorf("unable to get owner because of: %s", err.Error())
	}
	stickiedPosts, err := repo.GetStickiedPost(username)
	if err != nil {
		return nil, fmt.Errorf("unable to get bookmarked posts because of: %s", err.Error())
	}
	posts, err := repo.GetPosts(username)
	if err != nil {
		return nil, fmt.Errorf("unable to get posts because of: %s", err.Error())
	}
	unOfficialReleases, err := repo.GetUnOfficialRelease(username)
	if err != nil {
		return nil, fmt.Errorf("unable to get UnOfficialRelease because of: %s", err.Error())
	}
	officialReleases, err := repo.GetOfficialRelease(username)
	if err != nil {
		return nil, fmt.Errorf("unable to get UnOfficialRelease because of: %s", err.Error())
	}
	c.AdminUsernames = admins
	c.OwnerUsername = owner
	c.StickiedPostIDs = stickiedPosts
	c.PostIDs = posts
	c.ReleaseIDs = unOfficialReleases
	c.OfficialReleaseIDs = officialReleases
	c.Username = username
	return c, nil
}
func (repo *ChannelRepository) UpdateChannel(username string, c *channel.Channel) error {
	var err error
	if c.Name != "" {
		err = repo.execUpdateStatementOnColumn("name", c.Name, username)
		if err != nil {
			return err
		}
	}
	if c.Username != "" {
		err = repo.execUpdateStatementOnColumn("username", c.Username, username)
		if err != nil {
			return err
		}
	}
	if c.Description != "" {
		err = repo.execUpdateStatementOnColumn("description", c.Description, username)
		if err != nil {
			return err
		}
	}
	return nil
}

func (repo *ChannelRepository) execUpdateStatementOnColumn(column, value, username string) error {
	_, err := repo.db.Exec(fmt.Sprintf(`UPDATE "issue#1".channels 
									SET %s = $1 
									WHERE username = $2`, column), value, username)
	if err != nil {
		return fmt.Errorf("updating failed of %s column with %s because of: %s", column, value, err.Error())
	}
	return nil
}
func (repo *ChannelRepository) DeleteChannel(username string) error {
	_, err := repo.db.Exec(`DELETE FROM "issue#1".channels
							WHERE username = $1`, username)
	if err != nil {
		return fmt.Errorf("deletion of tuple from channels because of: %s", err.Error())
	}
	return nil
}
func (repo *ChannelRepository) GetUnOfficialRelease(username string) ([]int, error) {
	var UnOfficialList []int
	var Release int
	var official bool
	rows, err := repo.db.Query(`SELECT release_id,is_official
                FROM "issue#1".channel_catalog
                WHERE channel_username = $1`, username)
	if err != nil {
		return nil, fmt.Errorf("querying for unofficial catalog failed because of: %v", err)
	}
	defer rows.Close()
	i := 0
	for rows.Next() {
		err := rows.Scan(&Release, &official)
		if err != nil {
			return nil, fmt.Errorf("scanning from rows failed because: %v", err)
		}
		if official == false {

			fmt.Errorf("is it here")
			UnOfficialList = append(UnOfficialList, Release)
		}
		i++
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("scanning from rows faulty because: %v", err)
	}
	return UnOfficialList, nil
}
func (repo *ChannelRepository) GetOfficialRelease(username string) ([]int, error) {
	var OfficialList []int
	var Release int
	var official bool
	rows, err := repo.db.Query(`SELECT release_id,is_official
                FROM "issue#1".channel_catalog
                WHERE channel_username = $1`, username)
	if err != nil {
		return nil, fmt.Errorf("querying for unofficial catalog failed because of: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&Release, &official)
		if err != nil {
			return nil, fmt.Errorf("scanning from rows failed because: %v", err)
		}
		if official == true {
			OfficialList = append(OfficialList, Release)
		}

	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("scanning from rows faulty because: %v", err)
	}
	return OfficialList, nil
}
func (repo *ChannelRepository) GetPosts(username string) ([]int, error) {
	var PostList []int
	var Post int

	rows, err := repo.db.Query(`SELECT id
                FROM "issue#1".posts
                WHERE channel_from = $1`, username)
	if err != nil {
		return nil, fmt.Errorf("querying for posts failed because of: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&Post)
		if err != nil {
			return nil, fmt.Errorf("scanning from rows failed because: %v", err)
		}

		PostList = append(PostList, Post)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("scanning from rows faulty because: %v", err)
	}
	return PostList, nil
}
func (repo *ChannelRepository) GetStickiedPost(username string) ([]int, error) {
	var stickied []int

	var Post int

	rows, err := repo.db.Query(`SELECT
  post_id
FROM
  "issue#1".channel_stickies
INNER JOIN "issue#1".posts ON channel_stickies.post_id =posts.id WHERE posts.channel_from=$1
 `, username)
	if err != nil {
		return nil, fmt.Errorf("querying for posts failed because of: %v", err)
	}
	defer rows.Close()
	i := 0
	for rows.Next() {
		err := rows.Scan(&Post)
		if err != nil {
			return nil, fmt.Errorf("scanning from rows failed because: %v", err)
		}
		stickied = append(stickied, Post)
		i++
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("scanning from rows faulty because: %v", err)
	}

	return stickied, nil
}
func (repo *ChannelRepository) GetOwner(username string) (string, error) {
	var owner string
	var Admin string
	var isOwner bool
	rows, err := repo.db.Query(`SELECT "user",is_owner
                FROM "issue#1".channel_admins
                WHERE channel_username = $1`, username)
	if err != nil {
		return "", fmt.Errorf("querying for admins failed because of: %v", err)
	}
	defer rows.Close()
	i := 0
	for rows.Next() {
		err := rows.Scan(&Admin, &isOwner)
		if err != nil {
			return "", fmt.Errorf("scanning from rows failed because: %v", err)
		}
		if isOwner {
			owner = Admin
		}
		i++
	}
	err = rows.Err()
	if err != nil {
		return "", fmt.Errorf("scanning from rows faulty because: %v", err)
	}

	return owner, nil
}
func (repo *ChannelRepository) SearchChannels(pattern string, sortBy channel.SortBy, sortOrder channel.SortOrder, limit, offset int) ([]*channel.Channel, error) {
	var channels = make([]*channel.Channel, 0)
	var err error
	var rows *sql.Rows
	if pattern == "" {
		rows, err = repo.db.Query(fmt.Sprintf(`(SELECT username,name, COALESCE(description, ''),creation_time 
												FROM "issue#1".channels) 
												ORDER BY %s %s NULLS LAST
												LIMIT $1 OFFSET $2`, sortBy, sortOrder), limit, offset)
		if err != nil {
			return nil, fmt.Errorf("querying for channels failed because of: %s", err.Error())
		}
		defer rows.Close()
	} else {
		// TODO actual search queries
	}

	var creationTimeString string
	for rows.Next() {
		c := channel.Channel{}
		err := rows.Scan(&c.Username, &c.Name, &c.Description, &creationTimeString)
		if err != nil {
			return nil, fmt.Errorf("scanning from rows failed because: %s", err.Error())
		}
		creationTime, err := time.Parse(time.RFC3339, creationTimeString)
		if err != nil {
			return nil, fmt.Errorf("parsing of timestamp to time.Time failed because of: %s", err.Error())
		}
		creationTime, errC := time.Parse(time.RFC3339, creationTimeString)
		if errC != nil {
			return nil, fmt.Errorf("parsing of timestamp to time.Time failed because of: %s", err.Error())
		}
		c.CreationTime = creationTime

		admins, err := repo.GetAdmins(c.Username)
		if err != nil {
			return nil, fmt.Errorf("unable to get bookmarked posts because of: %s", err.Error())
		}
		owner, err := repo.GetOwner(c.Username)
		if err != nil {
			return nil, fmt.Errorf("unable to get bookmarked posts because of: %s", err.Error())
		}
		stickiedPosts, err := repo.GetStickiedPost(c.Username)
		if err != nil {
			return nil, fmt.Errorf("unable to get bookmarked posts because of: %s", err.Error())
		}
		posts, err := repo.GetPosts(c.Username)
		if err != nil {
			return nil, fmt.Errorf("unable to get bookmarked posts because of: %s", err.Error())
		}
		unOfficialReleases, err := repo.GetUnOfficialRelease(c.Username)
		if err != nil {
			return nil, fmt.Errorf("unable to get bookmarked posts because of: %s", err.Error())
		}
		officialReleases, err := repo.GetUnOfficialRelease(c.Username)
		if err != nil {
			return nil, fmt.Errorf("unable to get bookmarked posts because of: %s", err.Error())
		}
		c.AdminUsernames = admins
		c.OwnerUsername = owner
		c.StickiedPostIDs = stickiedPosts
		c.PostIDs = posts
		c.ReleaseIDs = unOfficialReleases
		c.OfficialReleaseIDs = officialReleases

		channels = append(channels, &c)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("scanning from rows faulty because: %s", err.Error())
	}
	return channels, nil
}

func (repo *ChannelRepository) AddAdmin(channelUsername string, adminUsername string) error {
	var err error
	owner := false
	_, err = repo.db.Exec(`INSERT INTO "issue#1".channel_admins (channel_username,"user",is_owner)
							VALUES ($1, $2,$3)`, channelUsername, adminUsername, owner)
	if err != nil {
		return fmt.Errorf("insertion of user failed because of: %s", err.Error())
	}
	return nil
}
func (repo *ChannelRepository) DeleteAdmin(channelUsername string, adminUsername string) error {
	_, err := repo.db.Exec(`DELETE FROM "issue#1".channel_admins
							WHERE channel_username = $1 AND "user"= $2`, channelUsername, adminUsername)
	if err != nil {
		return fmt.Errorf("deletion of tuple from channel_admins because of: %s", err.Error())
	}
	return nil
}
func (repo *ChannelRepository) GetAdmins(username string) ([]string, error) {
	var AdminList []string
	var Admin string
	rows, err := repo.db.Query(`SELECT "user"
                FROM "issue#1".channel_admins
                WHERE channel_username = $1`, username)
	if err != nil {
		return nil, fmt.Errorf("querying for admins failed because of: %v", err)
	}
	defer rows.Close()
	i := 0
	for rows.Next() {
		err := rows.Scan(&Admin)
		if err != nil {
			return nil, fmt.Errorf("scanning from rows failed because: %v", err)
		}
		AdminList = append(AdminList, Admin)
		i++
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("scanning from rows faulty because: %v", err)
	}
	return AdminList, nil
}

func (repo *ChannelRepository) ChangeOwner(channelUsername string, ownerUsername string) error {
	var err error
	var owner bool = true
	_, err = repo.db.Exec(`UPDATE "issue#1".channel_admins
								  SET is_owner = $3 WHERE channel_username =$1 AND "user"=$2`, channelUsername, ownerUsername, owner)
	if err != nil {
		return fmt.Errorf("changing of owner failed because of: %s", err.Error())
	}
	return nil
}
func (repo *ChannelRepository) AddReleaseToOfficialCatalog(channelUsername string, releaseID int) error {
	is_official := true
	_, err := repo.db.Exec(`UPDATE "issue#1".channel_catalog
							SET is_official = $3
							WHERE channel_username = $1 AND release_id = $2`, channelUsername, releaseID, is_official)
	if err != nil {
		return fmt.Errorf("deletion of tuple from channel_catalogs because of: %s", err.Error())
	}
	return nil
}
func (repo *ChannelRepository) DeleteReleaseFromCatalog(channelUsername string, ReleaseID int) error {
	_, err := repo.db.Exec(`DELETE FROM "issue#1".channel_catalog
							WHERE channel_username = $1 AND release_id = $2`, channelUsername, ReleaseID)
	if err != nil {
		return fmt.Errorf("deletion of tuple from channel_catalogs because of: %s", err.Error())
	}
	return nil
}

func (repo *ChannelRepository) StickyPost(channelUsername string, postID int) error {
	a, err := repo.GetStickiedPost(channelUsername)
	if err != nil {
		return fmt.Errorf("getting stickied posts failed because of: %s", err.Error())
	}
	if len(a) == 2 {
		return channel.ErrStickiedPostFull
	} else {
		_, err := repo.db.Exec(`INSERT INTO channel_stickies (post_id)
							VALUES ($1)
							ON CONFLICT DO NOTHING`, postID)
		const foreignKeyViolationErrorCode = pq.ErrorCode("23503")
		if err != nil {
			if pgErr, isPGErr := err.(pq.Error); !isPGErr {
				if pgErr.Code != foreignKeyViolationErrorCode {
					return user.ErrPostNotFound
				}
				return fmt.Errorf("inserting into channel_stickies failed because of: %s", err.Error())
			}
		}
	}
	return nil
}

func (repo *ChannelRepository) DeleteStickiedPost(channelUsername string, stickiedPostID int) error {
	//TODO
	_, err := repo.db.Exec(`DELETE FROM "issue#1".channel_stickies
							WHERE post_id = $1`, stickiedPostID)
	if err != nil {
		return fmt.Errorf("deletion of tuple from channel_stickie because of: %s", err.Error())
	}
	return nil
}
