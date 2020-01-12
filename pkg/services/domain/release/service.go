/*
Package release contains definition and implementation of a service that deals with Release entities */
package release

import (
	"fmt"
)

// Service specifies a method to service Release entities.
type Service interface {
	GetRelease(id int) (*Release, error)
	SearchRelease(pattern string, by SortBy, order SortOrder, limit int, offset int) ([]*Release, error)
	DeleteRelease(id int) error
	AddRelease(r *Release) (*Release, error)
	UpdateRelease(rel *Release) (*Release, error)
}

// Repository specifies a repo interface to serve the release Service interface
type Repository interface {
	GetRelease(id int) (*Release, error)
	SearchRelease(pattern string, by SortBy, order SortOrder, limit int, offset int) ([]*Release, error)
	DeleteRelease(id int) error
	AddRelease(r *Release) (*Release, error)
	UpdateRelease(rel *Release) (*Release, error)
}

// SortOrder holds enums used by SearchRelease methods the order of Users are sorted with
type SortOrder string

// SortBy  holds enums used by SearchRelease methods the attribute of Users are sorted with
type SortBy string

// Sorting constants used by SearchRelease methods
const (
	SortAscending  SortOrder = "ASC"
	SortDescending SortOrder = "DESC"

	SortCreationTime SortBy = "creation_time"
	SortByOwner      SortBy = "owner_channel"
	SortByType       SortBy = "type"
)

// ErrReleaseNotFound is returned when the requested release is not found
var ErrReleaseNotFound = fmt.Errorf("release not found")

// ErrInvalidReleaseData is returned when the requested passed release has invalid dat
var ErrInvalidReleaseData = fmt.Errorf("release data invalid")

// ErrSomeReleaseDataNotPersisted is returned when the requested passed release has invalid dat
var ErrSomeReleaseDataNotPersisted = fmt.Errorf("was unable to persist some release data")

// ErrAttemptToChangeReleaseType is returned when the requested passed release has invalid dat
var ErrAttemptToChangeReleaseType = fmt.Errorf("attempt to change release type")

type service struct {
	repo *Repository
}

// NewService returns a struct that implements the release.Service interface
func NewService(repo *Repository) Service {
	return &service{repo: repo}
}

// AddRelease adds an new release based on the passed in struct
func (s service) AddRelease(r *Release) (*Release, error) {
	if r.Content == "" || r.OwnerChannel == "" {
		return nil, ErrInvalidReleaseData
	}
	return (*s.repo).AddRelease(r)
}

// GetRelease gets the release stored under the given id.
func (s service) GetRelease(id int) (*Release, error) {
	return (*s.repo).GetRelease(id)
}

// SearchRelease returns a list of official releases that match against the pattern.
// Note: this won't return releases that aren't in a channel's official catalog.
// If pattern is empty, it returns all releases.
// Sorting and pagination can be specified.
func (s service) SearchRelease(pattern string, by SortBy, order SortOrder, limit int, offset int) ([]*Release, error) {
	if limit < 0 || offset < 0 {
		return nil, fmt.Errorf("invalid pagination")
	}
	return (*s.repo).SearchRelease(pattern, by, order, limit, offset)
}

// DeleteRelease removes the release stored under the given id.
func (s service) DeleteRelease(id int) error {
	if _, err := s.GetRelease(id); err != nil {
		return err
	}
	return (*s.repo).DeleteRelease(id)
}

// UpdateRelease updates the release stored under the given id
// based on the passed in struct.
func (s service) UpdateRelease(r *Release) (*Release, error) {
	if rel, err := s.GetRelease(r.ID); err != nil {
		return s.AddRelease(r)
	} else {
		if r.Type != "" && r.Type != rel.Type {
			return nil, ErrAttemptToChangeReleaseType
		}
		r.Authors = mergeStringSlicesRemovingDuplicates(r.Authors, rel.Authors)
		r.Genres = mergeStringSlicesRemovingDuplicates(r.Genres, rel.Genres)
		if r.OwnerChannel == rel.OwnerChannel {
			r.OwnerChannel = ""
		}
	}
	return (*s.repo).UpdateRelease(r)
}

// removeDuplicates is a helper function
func mergeStringSlicesRemovingDuplicates(slice1, slice2 []string) []string {
	set := make(map[string]struct{})
	for _, value := range slice1 {
		set[value] = struct{}{}
	}
	for _, value := range slice2 {
		set[value] = struct{}{}
	}
	slice3 := make([]string, 0)
	for value := range set {
		slice3 = append(slice3, value)
	}
	return slice3
}
