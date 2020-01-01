/*
Package release contains definition and implementation of a service that deals with Release entities */
package release

import (
	"fmt"
)

// Service specifies a method to service User entities.
type Service interface {
	GetRelease(id int) (*Release, error)
	SearchRelease(pattern string, by SortBy, order SortOrder, limit int, offset int) (*[]Release, error)
	DeleteRelease(id int) error
	AddRelease(r *Release) (*Release, error)
	UpdateRelease(rel *Release) (*Release, error)
}
type Repository interface {
	GetRelease(id int) (*Release, error)
	SearchRelease(pattern string, by SortBy, order SortOrder, limit int, offset int) (*[]Release, error)
	DeleteRelease(id int) error
	AddRelease(r *Release) (*Release, error)
	UpdateRelease(rel *Release) (*Release, error)
}

// SortOrder holds enums used by SearchUser methods the order of Users are sorted with
type SortOrder string

// SortBy  holds enums used by SearchUser methods the attribute of Users are sorted with
type SortBy string

// Sorting constants used by SearchUser methods
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

type service struct {
	repo *Repository
}

func NewService(repo *Repository) Service {
	return &service{repo: repo}
}

func (s service) AddRelease(r *Release) (*Release, error) {
	if r.Content == "" || r.OwnerChannel == "" {
		return nil, ErrInvalidReleaseData
	}
	return (*s.repo).AddRelease(r)
}

func (s service) GetRelease(id int) (*Release, error) {
	return (*s.repo).GetRelease(id)
}

func (s service) SearchRelease(pattern string, by SortBy, order SortOrder, limit int, offset int) (*[]Release, error) {
	if limit < 0 || offset < 0 {
		return nil, fmt.Errorf("invalid pagination")
	}
	return (*s.repo).SearchRelease(pattern, by, order, limit, offset)
}

func (s service) DeleteRelease(id int) error {
	if _, err := s.GetRelease(id); err != nil {
		return err
	}
	return (*s.repo).DeleteRelease(id)
}

func (s service) UpdateRelease(r *Release) (*Release, error) {
	if _, err := s.GetRelease(r.ID); err != nil {
		return s.AddRelease(r)
	}
	return (*s.repo).UpdateRelease(r)
}
