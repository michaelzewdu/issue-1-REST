package memory

import (
	. "github.com/slim-crown/issue-1-REST/pkg/domain/channel"
)

type ChannelRepository struct {
	cache         map[string]Channel
	secondaryRepo *Repository
	allRepos      *map[string]interface{}
}

func NewChannelRepository(secondaryRepo *Repository, allRepos *map[string]interface{}) *ChannelRepository {
	return &ChannelRepository{cache: make(map[string]Channel), secondaryRepo: secondaryRepo, allRepos: allRepos}
}
func (repo *ChannelRepository) cacheChannel(username string) error {
	c, err := (*repo.secondaryRepo).GetChannel(username)
	if err != nil {
		return err
	}

	repo.cache[username] = *c

	return err
}
func (repo *ChannelRepository) AddChannel(channel *Channel) error {
	return (*repo.secondaryRepo).AddChannel(channel)
}
func (repo *ChannelRepository) GetChannel(username string) (*Channel, error) {
	if _, ok := repo.cache[username]; ok == false {

		err := repo.cacheChannel(username)
		if err != nil {
			return nil, err
		}
	}

	c := repo.cache[username]
	return &c, nil
}
func (repo *ChannelRepository) UpdateChannel(username string, c *Channel) error {
	err := (*repo.secondaryRepo).UpdateChannel(username, c)
	if err == nil {
		if c.Username != "" {
			err = repo.cacheChannel(c.Username)
			if err != nil {
				return err
			}
		} else {
			delete(repo.cache, c.Username)
			err = repo.cacheChannel(username)
			if err != nil {
				return err
			}
		}
	}
	return err
}
func (repo *ChannelRepository) DeleteChannel(username string) error {
	err := (*repo.secondaryRepo).DeleteChannel(username)
	if err == nil {
		delete(repo.cache, username)
	}
	return err
}

func (repo *ChannelRepository) SearchChannels(pattern string, sortBy SortBy, sortOrder SortOrder, limit, offset int) ([]*Channel, error) {
	result, err := (*repo.secondaryRepo).SearchChannels(pattern, sortBy, sortOrder, limit, offset)
	if err == nil {
		for _, c := range result {
			cTemp := *c
			repo.cache[c.Username] = cTemp
			c.Name = ""
			// TODO

		}
	}
	return result, err
}

func (repo *ChannelRepository) AddAdmin(channelUsername string, adminUsername string) error {
	err := (*repo.secondaryRepo).AddAdmin(channelUsername, adminUsername)
	if err == nil {
		err = repo.cacheChannel(channelUsername)
		if err != nil {
			return err
		}
	}
	return err
}
func (repo *ChannelRepository) DeleteAdmin(channelUsername string, adminUsername string) error {
	err := (*repo.secondaryRepo).DeleteAdmin(channelUsername, adminUsername)
	if err == nil {
		err = repo.cacheChannel(channelUsername)
		if err != nil {
			return err
		}
	}
	return err
}
func (repo *ChannelRepository) ChangeOwner(channelUsername string, ownerUsername string) error {
	err := (*repo.secondaryRepo).ChangeOwner(channelUsername, ownerUsername)
	if err == nil {
		err = repo.cacheChannel(channelUsername)
		if err != nil {
			return err
		}
	}
	return err
}
func (repo *ChannelRepository) AddReleaseToOfficialCatalog(channelUsername string, releaseID int) error {
	err := (*repo.secondaryRepo).AddReleaseToOfficialCatalog(channelUsername, releaseID)
	if err == nil {
		err = repo.cacheChannel(channelUsername)
		if err != nil {
			return err
		}
	}
	return err
}
func (repo *ChannelRepository) DeleteReleaseFromCatalog(channelUsername string, ReleaseID int) error {
	err := (*repo.secondaryRepo).DeleteReleaseFromCatalog(channelUsername, ReleaseID)
	if err == nil {
		err = repo.cacheChannel(channelUsername)
		if err != nil {
			return err
		}
	}
	return err
}

func (repo *ChannelRepository) StickyPost(channelUsername string, postID int) error {
	err := (*repo.secondaryRepo).StickyPost(channelUsername, postID)
	if err == nil {
		err = repo.cacheChannel(channelUsername)
		if err != nil {
			return err
		}
	}
	return err
}

func (repo *ChannelRepository) DeleteStickiedPost(channelUsername string, stickiedPostID int) error {
	err := (*repo.secondaryRepo).DeleteStickiedPost(channelUsername, stickiedPostID)
	if err == nil {
		err = repo.cacheChannel(channelUsername)
		if err != nil {
			return err
		}
	}
	return err
}
