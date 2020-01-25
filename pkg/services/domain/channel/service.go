/*
Package channel contains definition and implemntation of a service that deals with User entities */
package channel

import "fmt"

type Service interface {
	AddChannel(channel *Channel) (*Channel, error)
	GetChannel(username string) (*Channel, error)
	UpdateChannel(username string, channel *Channel) (*Channel, error)
	SearchChannels(pattern string, sortBy SortBy, sortOrder SortOrder, limit, offset int) ([]*Channel, error)
	DeleteChannel(username string) error
	AddAdmin(channelUsername string, adminUsername string) error
	DeleteAdmin(channelUsername string, adminUsername string) error
	ChangeOwner(channelUsername string, ownerUsername string) error
	DeleteReleaseFromCatalog(channelUsername string, ReleaseID uint) error
	DeleteReleaseFromOfficialCatalog(channelUsername string, ReleaseID uint) error
	AddReleaseToOfficialCatalog(channelUsername string, releaseID uint, postID uint) error
	DeleteStickiedPost(channelUsername string, stickiedPostID uint) error
	StickyPost(channelUsername string, postID uint) error
	AddPicture(channelUsername string, name string) (string, error)
	RemovePicture(channelUsername string) error
}
type Repository interface {
	AddChannel(channel *Channel) (*Channel, error)
	GetChannel(username string) (*Channel, error)
	UpdateChannel(username string, channel *Channel) (*Channel, error)
	SearchChannels(pattern string, sortBy SortBy, sortOrder SortOrder, limit, offset int) ([]*Channel, error)
	DeleteChannel(username string) error
	AddAdmin(channelUsername string, adminUsername string) error
	DeleteAdmin(channelUsername string, adminUsername string) error
	ChangeOwner(channelUsername string, ownerUsername string) error
	DeleteReleaseFromCatalog(channelUsername string, ReleaseID uint) error
	DeleteReleaseFromOfficialCatalog(channelUsername string, ReleaseID uint) error
	AddReleaseToOfficialCatalog(channelUsername string, releaseID uint, postID uint) error
	DeleteStickiedPost(channelUsername string, stickiedPostID uint) error
	StickyPost(channelUsername string, postID uint) error
	AddPicture(channelUsername string, name string) (string, error)
	RemovePicture(channelUsername string) error
}
type SortOrder string
type SortBy string

const (
	SortAscending  SortOrder = "ASC"
	SortDescending SortOrder = "DESC"

	SortCreationTime SortBy = "creation_time"
	SortByUsername   SortBy = "username"
	SortByName       SortBy = "name"
)

// ErrUserNameOccupied is returned when the channel username specified is occupied
var ErrUserNameOccupied = fmt.Errorf("user name is occupied")

// ErrPostAlreadyStickied is returned when the post provided is already a sticky post
var ErrPostAlreadyStickied = fmt.Errorf("post already stickied")

// ErrAdminAlreadyExists is returned when the channel username specified already has specified user as admin
var ErrAdminAlreadyExists = fmt.Errorf("user is already an admin")

// ErrChannelNotFound is returned when the  channel username specified isn't recognized
var ErrChannelNotFound = fmt.Errorf("channel not found  error")

// ErrReleaseAlreadyExists is returned when a unique key violation happens
var ErrReleaseAlreadyExists = fmt.Errorf("release already exists")

// ErrInvalidChannelData is returned when the channel username specified isn't recognized
var ErrInvalidChannelData = fmt.Errorf("passed channel data is invalid")

// ErrAdminNotFound is returned when the channel Admin username specified isn't recognized
var ErrAdminNotFound = fmt.Errorf("admin not found")

// ErrOwnerNotFound is returned when the channel Owner username specified isn't recognized
var ErrOwnerNotFound = fmt.Errorf("owner not found")

// ErrOwnerToBeNotAdmin is returned when the channel Owner username specified isn't an admin
var ErrOwnerToBeNotAdmin = fmt.Errorf("owner to be is not an admin")

// ErrReleaseNotFound is returned when the channel Catalog Release ID specified isn't recognized
var ErrReleaseNotFound = fmt.Errorf("release not found")

// ErrStickiedPostNotFound is returned when the channel Stickied Post ID specified isn't recognized
var ErrStickiedPostNotFound = fmt.Errorf("stickied post not found")

// ErrPostNotFound is returned when the channel Post ID specified isn't recognized
var ErrPostNotFound = fmt.Errorf("post not found")

// ErrStickiedPostFull is returned when the channel has filled it's stickied post quota
var ErrStickiedPostFull = fmt.Errorf("two posts already stickied")

type service struct {
	allServices *map[string]interface{}
	repo        *Repository
}

// NewService returns a struct that implements the Service interface
func NewService(repo *Repository, allServices *map[string]interface{}) Service {
	s := &service{allServices: allServices, repo: repo}
	return s
}

//
func (service *service) IsPostFromChannel(channelUsername string, postID uint) bool {
	c, _ := service.GetChannel(channelUsername)
	i := 0
	for ; i < len(c.PostIDs); i++ {
		if c.PostIDs[i] == postID {
			return true
		}
	}
	return false
}

// AddChannel adds a channel
func (service *service) AddChannel(channel *Channel) (*Channel, error) {
	if channel.Name == "" && channel.ChannelUsername == "" {
		return nil, ErrInvalidChannelData
	}
	a, _ := service.GetChannel(channel.ChannelUsername)
	if a != nil {
		return nil, ErrUserNameOccupied
	}
	return (*service.repo).AddChannel(channel)
}

// GetChannel gets a channel according to the given username
func (service *service) GetChannel(username string) (*Channel, error) {
	return (*service.repo).GetChannel(username)
}

// UpdateChannel updates a channel according to the given username and channel
func (service *service) UpdateChannel(username string, channel *Channel) (*Channel, error) {
	_, err := service.GetChannel(username)
	if err != nil {
		//fmt.Errorf("channel can't be updated because %s", err.Error())
		return nil, err
	}
	a, _ := service.GetChannel(channel.ChannelUsername)
	if a != nil {
		return nil, ErrUserNameOccupied
	}
	return (*service.repo).UpdateChannel(username, channel)
}

// SearchChannels searches the channel of the given username
func (service *service) SearchChannels(pattern string, sortBy SortBy, sortOrder SortOrder, limit, offset int) ([]*Channel, error) {
	if limit < 0 || offset < 0 {
		return nil, fmt.Errorf("invalid pagination")
	}
	return (*service.repo).SearchChannels(pattern, sortBy, sortOrder, limit, offset)
}

// DeleteChannel removes the channel of the given username
func (service *service) DeleteChannel(username string) error {
	_, err := service.GetChannel(username)
	if err != nil {
		//fmt.Errorf("channel can't be deleted because %s", err.Error())
		return err
	}
	return (*service.repo).DeleteChannel(username)
}

// AddAdmin adds the given admin adminUsername from the channel of given username,ChannelUsername
func (service *service) AddAdmin(channelUsername string, adminUsername string) error {
	_, err := service.GetChannel(channelUsername)
	if err != nil {
		//fmt.Errorf("channel can'add admin because %s", err.Error())
		return err
	}
	return (*service.repo).AddAdmin(channelUsername, adminUsername)
}

// DeleteAdmin deletes the given admin adminUsername from the channel of given username,ChannelUsername
func (service *service) DeleteAdmin(channelUsername string, adminUsername string) error {
	_, err := service.GetChannel(channelUsername)
	if err != nil {
		//fmt.Errorf("channel can'delete admin because %s", err.Error())
		return err
	}
	return (*service.repo).DeleteAdmin(channelUsername, adminUsername)
}

// DeleteReleaseFromOfficialCatalog deletes the given release ReleaseID from the catalog of channel of given username,ChannelUsername
func (service *service) DeleteReleaseFromCatalog(channelUsername string, ReleaseID uint) error {
	c, err := service.GetChannel(channelUsername)
	if err != nil {
		//fmt.Errorf("channel can'delete release because %s", err.Error())
		return err
	}
	i := 0
	oia := false

	for ; i < len(c.ReleaseIDs); i++ {
		if c.ReleaseIDs[i] == ReleaseID {
			fmt.Printf("%d", c.ReleaseIDs[i])
			oia = true
			break
		} else {
			oia = false
		}
	}
	if oia == false {
		fmt.Printf("%d", ReleaseID)
		return ErrReleaseNotFound
	}
	return (*service.repo).DeleteReleaseFromCatalog(channelUsername, ReleaseID)
}

// DeleteReleaseFromOfficialCatalog deletes the given release ReleaseID from the official catalog of channel of given username,ChannelUsername
func (service *service) DeleteReleaseFromOfficialCatalog(channelUsername string, ReleaseID uint) error {

	c, err := service.GetChannel(channelUsername)
	if err != nil {
		//fmt.Errorf("channel can'delete release because %s", err.Error())
		return err
	}
	i := 0
	oia := false

	for ; i < len(c.OfficialReleaseIDs); i++ {
		if c.OfficialReleaseIDs[i] == ReleaseID {
			oia = true
			break
		} else {
			oia = false
		}
	}
	if oia == false {
		return ErrReleaseNotFound
	}
	return (*service.repo).DeleteReleaseFromOfficialCatalog(channelUsername, ReleaseID)
}

// AddReleaseToOfficialCatalog adds the given release ReleaseID from the official catalog of channel of given username,ChannelUsername
func (service *service) AddReleaseToOfficialCatalog(channelUsername string, releaseID uint, postID uint) error {
	_, err := service.GetChannel(channelUsername)
	if err != nil {
		//fmt.Errorf("channel add  release because %s", err.Error())
		return err
	}
	if !service.IsPostFromChannel(channelUsername, postID) {
		return ErrPostNotFound
	}
	return (*service.repo).AddReleaseToOfficialCatalog(channelUsername, releaseID, postID)
}

// DeleteStickiedPost deletes the given post id from stickied post for the channel of given username,channelUsername.
func (service *service) DeleteStickiedPost(channelUsername string, stickiedPostID uint) error {
	c, err := service.GetChannel(channelUsername)
	if err != nil {
		//fmt.Errorf("channel can'delete stickied post because %s", err.Error())
		return err
	}
	i := 0
	oia := false

	for ; i < len(c.StickiedPostIDs); i++ {
		if c.StickiedPostIDs[i] == stickiedPostID {
			oia = true
			break
		} else {
			oia = false
		}
	}
	if oia == false {
		return ErrStickiedPostNotFound
	}
	return (*service.repo).DeleteStickiedPost(channelUsername, stickiedPostID)
}

// ChangeOwner changes the given ownerUsername as the owner for the given channel of username channelUsername.
func (service *service) ChangeOwner(channelUsername string, ownerUsername string) error {
	c, err := service.GetChannel(channelUsername)
	if err != nil {
		//fmt.Errorf("channel can't change owner because %s", err.Error())
		return err
	}

	i := 0
	oia := false

	for ; i < len(c.AdminUsernames); i++ {
		if c.AdminUsernames[i] == ownerUsername {
			oia = true
			break
		} else {
			oia = false
		}
	}
	if oia == false {
		return ErrOwnerToBeNotAdmin
	}
	return (*service.repo).ChangeOwner(channelUsername, ownerUsername)
}

// StickyPost sticks the given postID for the channel of the given username on top of the post view of channel.
func (service *service) StickyPost(channelUsername string, postID uint) error {
	c, err := service.GetChannel(channelUsername)
	if err != nil {
		//fmt.Errorf("channel can't sticky post because %s", err.Error())
		return err
	}
	i := 0
	oia := false

	for ; i < len(c.StickiedPostIDs); i++ {
		if c.StickiedPostIDs[i] == postID {
			oia = true
			break
		} else {
			oia = false
		}
	}
	if oia == true {
		return ErrPostAlreadyStickied
	}
	return (*service.repo).StickyPost(channelUsername, postID)
}

// AddPicture adds the given image name as the picture for the given username.
func (service *service) AddPicture(channelUsername string, name string) (string, error) {
	_, err := service.GetChannel(channelUsername)
	if err != nil {
		//fmt.Errorf("channel can't add picture because %s", err.Error())
		return "", err
	}
	return (*service.repo).AddPicture(channelUsername, name)
}

// RemovePicture removes the picture for the given username.
func (service *service) RemovePicture(channelUsername string) error {
	_, err := service.GetChannel(channelUsername)
	if err != nil {
		//fmt.Errorf("channel can't remove picture because %s", err.Error())
		return err
	}
	return (*service.repo).RemovePicture(channelUsername)
}
