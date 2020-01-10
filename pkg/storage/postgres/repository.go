/*
Package postgres contains implementations of business level objects
all contained in a single package to help with circular dependencies */
package postgres

import (
	"database/sql"
)

/*
//DBHandler ...
type DBHandler interface {
	Execute(statement string) error
	Query(query string) Row
}

// Row ...
type Row interface {
	Scan(dest ...interface{})
	Next() bool
}
*/

type repository struct {
	db       *sql.DB
	allRepos *map[string]interface{}
}
