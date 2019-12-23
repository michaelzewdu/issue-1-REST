/*
Package memory contains cache based implementations of the different
Repository interfaces defined in the domain packages */
package memory

import (
	"github.com/slim-crown/issue-1-REST/pkg/domain/user"
)

// UserRepository ...
type UserRepository struct {
	cache         map[string]user.User
	secondaryRepo *user.Repository
	allRepos      *map[string]interface{}
}
