package main

import (
	"database/sql"
	"fmt"

	"github.com/slim-crown/Issue-1/pkg/domain/channel"
	"github.com/slim-crown/Issue-1/pkg/domain/user"

	"github.com/slim-crown/Issue-1/pkg/storage/memory"

	"github.com/slim-crown/Issue-1/pkg/http/rest"
	"github.com/slim-crown/Issue-1/pkg/storage/postgres"

	_ "github.com/lib/pq"
)

/*
type postgresHandler struct{
	dbConnection *sql.DB
}

func (dbHandler *postgresHandler) Query(query string) {

}

func NewPostgresHandler() postgres.DBHandler {
	db, err := sql.Open(
		"postgres",
		"user='issue #1 dev' " +
		"password='password1234!@#$' " +
		"dbname='issue #1' " +
		"sslmode=disable"
	)
	if err != nil {
		panic(err)
	}
	defer db.Close()
}
*/

type logger struct{}

// Log ...
func (logger *logger) Log(message string) {
	fmt.Printf(message)
}

func main() {

	dbConnection, err := sql.Open(
		"postgres",
		"user='issue #1 dev' "+
			"password='password1234!@#$' "+
			"dbname='issue #1' "+
			"sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer dbConnection.Close()

	services := make(map[string]interface{})
	cacheRepos := make(map[string]interface{})
	dbRepos := make(map[string]interface{})
	var logger rest.Logger = &logger{}
	{
		var usrDBRepo user.Repository = postgres.NewUserRepository(dbConnection, &dbRepos)
		dbRepos["User"] = &usrDBRepo
		var usrCacheRepo user.Repository = memory.NewUserRepository(&usrDBRepo, &cacheRepos)
		cacheRepos["User"] = &usrCacheRepo
		var usrService user.Service = user.NewService(&usrCacheRepo, &services)
		services["User"] = &usrService
	}
	{
		var chnlDBRepo channel.Repository = postgres.NewChannelRepository(dbConnection, &dbRepos)
		dbRepos["Channel"] = &chnlDBRepo
		var chnlCacheRepo channel.Repository = memory.NewChannelRepository(&chnlDBRepo, &cacheRepos)
		cacheRepos["Channel"] = &chnlCacheRepo
		var chnlService channel.Service = channel.NewService(&chnlCacheRepo, &services)
		services["Channel"] = &chnlService
	}
	// TODO other packages
	server := rest.NewServer(&logger, &services)
	server.ListenAndServe(":8080")
}
