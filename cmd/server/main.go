package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/slim-crown/issue-1-REST/pkg/domain/feed"
	"github.com/slim-crown/issue-1-REST/pkg/domain/user"

	"github.com/slim-crown/issue-1-REST/pkg/http/rest"

	"github.com/slim-crown/issue-1-REST/pkg/storage/memory"
	"github.com/slim-crown/issue-1-REST/pkg/storage/postgres"

	_ "github.com/lib/pq"
)

/*
type postgresHandler struct{
	db *sql.DB
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
func (logger *logger) Log(format string, a ...interface{}) {
	fmt.Printf("\n[%s] "+format, time.Now(), a)
}

func main() {
	var logger rest.Logger = &logger{}

	const(
		host = "localhost"
		port = "5432"
		dbname = "issue#1_db"
		role = "issue#1_dev"
		password = "password1234!@#$"
	)
	dataSourceName := fmt.Sprintf(
		`host=%s port=%s dbname='%s' user='%s' password='%s' sslmode=disable`,
		host, port, dbname, role, password)
	db, err := sql.Open(
		"postgres", dataSourceName)
	if err != nil {
		logger.Log("database connection failed because: %s", err.Error())
		return
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		logger.Log("database ping failed because: %s", err.Error())
		return
	}

	services := make(map[string]interface{})
	cacheRepos := make(map[string]interface{})
	dbRepos := make(map[string]interface{})

	{
		var usrDBRepo user.Repository = postgres.NewUserRepository(db, &dbRepos)
		dbRepos["User"] = &usrDBRepo
		var usrCacheRepo user.Repository = memory.NewUserRepository(&usrDBRepo, &cacheRepos)
		cacheRepos["User"] = &usrCacheRepo
		var usrService user.Service = user.NewService(&usrCacheRepo, &services)
		services["User"] = &usrService
	}
	{
		var feedDBRepo feed.Repository = postgres.NewFeedRepository(db, &dbRepos)
		dbRepos["Feed"] = &feedDBRepo
		var feedCacheRepo feed.Repository = memory.NewFeedRepository(&feedDBRepo, &cacheRepos)
		cacheRepos["Feed"] = &feedCacheRepo
		feedService := feed.NewService(&feedCacheRepo, &services)
		services["Feed"] = &feedService
	}
	/*
		{
			var chnlDBRepo channel.Repository = postgres.NewChannelRepository(db, &dbRepos)
			dbRepos["Channel"] = &chnlDBRepo
			var chnlCacheRepo channel.Repository = memory.NewChannelRepository(&chnlDBRepo, &cacheRepos)
			cacheRepos["Channel"] = &chnlCacheRepo
			var chnlService channel.Service = channel.NewService(&chnlCacheRepo, &services)
			services["Channel"] = &chnlService
		}
	*/
	mux := rest.NewMux(&logger, &services)
	logger.Log("starting up server...")
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		logger.Log("http server failed to start")
		return
	}
}
