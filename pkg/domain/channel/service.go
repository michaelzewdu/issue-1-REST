/*
Package channel contains definition and implemntation of a service that deals with User entities */
package channel

import "fmt"

type Service interface {
	AddChannel(channel *Channel) error
	GetChannel(username string) (*Channel, error)
	UpdateChannel(username string, channel *Channel) error
	SearchChannels(pattern string, sortBy SortBy, sortOrder SortOrder, limit, offset int) ([]*Channel, error)
	DeleteChannel(username string) error
	AddAdmin(channelUsername string, adminUsername string) error
	DeleteAdmin(channelUsername string, adminUsername string) error
	ChangeOwner(channelUsername string, ownerUsername string) error
	DeleteReleaseFromCatalog(username string, ReleaseID int) error
	AddReleaseToOfficialCatalog(username string, releaseID int) error
	DeleteStickiedPost(username string, stickiedPostID int) error
	StickyPost(username string, postID int) error
}
type Repository interface {
	AddChannel(channel *Channel) error
	GetChannel(username string) (*Channel, error)
	UpdateChannel(username string, channel *Channel) error
	SearchChannels(pattern string, sortBy SortBy, sortOrder SortOrder, limit, offset int) error
	DeleteChannel(username string) error
	AddAdmin(channelUsername string, adminUsername string) error
	DeleteAdmin(channelUsername string, adminUsername string) error
	DeleteReleaseFromCatalog(username string, ReleaseID int) error
	AddReleaseToOfficialCatalog(username string, releaseID int) error
	DeleteStickiedPost(username string, stickiedPostID int) error
	StickyPost(username string, postID int) error
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

var ErrUserNameOccupied = fmt.Errorf("user name is occupied")
var ErrChannelNotFound = fmt.Errorf("channel not found")
var ErrInvalidChannelData = fmt.Errorf("passed channel data is invalid")
var ErrAdminNotFound = fmt.Errorf("admin not found")
var ErrOwnerNotFound = fmt.Errorf("owner not found")
var ErrReleaseNotFound = fmt.Errorf("release not found")
var ErrStickiedPostNotFound = fmt.Errorf("release not found")
var ErrPostNotFound = fmt.Errorf("release not found")
var ErrStickiedPostFull = fmt.Errorf("two posts already stickied")

type service struct {
	allServices *map[string]interface{}
	repo        *Repository
}

func NewService(repo *Repository, allServices *map[string]interface{}) *service {
	s := &service{allServices: allServices, repo: repo}
	return s
}

func (service *service) AddChannel(channel *Channel) error {
	a, _ := service.GetChannel(channel.Username)
	if a != nil {
		fmt.Errorf("there is a channel %s with that name", (*channel).Username)
	}
	return (*service.repo).AddChannel(channel)
}
func (service *service) GetChannel(username string) (*Channel, error) {
	a, _ := service.GetChannel(username)
	if a == nil {
		fmt.Errorf("there is no channel %s with that name", username)
	}
	return (*service.repo).GetChannel(username)
}
func (service *service) UpdateChannel(username string, channel *Channel) error {
	_, err := service.GetChannel(username)
	if err != nil {
		fmt.Errorf("channel can't be updated because %s", err.Error())
	}
	a, _ := service.GetChannel(channel.Username)
	if a != nil {
		fmt.Errorf("there is a channel %s with that name", (*channel).Username)
	}
	return (*service.repo).UpdateChannel(username, channel)
}
func (service *service) SearchChannels(channel *Channel) error {
	return nil
}
func (service *service) DeleteChannel(username string) error {
	_, err := service.GetChannel(username)
	if err != nil {
		fmt.Errorf("channel can't be deleted because %s", err.Error())
	}
	return (*service.repo).DeleteChannel(username)
}
func (service *service) AddAdmin(channel *Channel) error {

	return nil
}
func (service *service) DeleteAdmin(channel *Channel) error {

	return nil
}
func (service *service) DeleteReleaseFromCatalog(channel *Channel) error {

	return nil
}
func (service *service) AddReleaseToOfficialCatalog(channel *Channel) error {

	return nil
}
func (service *service) DeleteStickiedPost(channel *Channel) error {

	return nil
}
func (service *service) ChangeOwner(channel *Channel) error {

	return nil
}
func (service *service) StickyPost(channel *Channel) error {

	return nil
}
