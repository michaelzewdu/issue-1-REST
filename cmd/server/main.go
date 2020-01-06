package main

import (
	"database/sql"
	"fmt"
	"github.com/microcosm-cc/bluemonday"
	"github.com/slim-crown/issue-1-REST/pkg/domain/feed"
	"github.com/slim-crown/issue-1-REST/pkg/domain/release"
	"github.com/slim-crown/issue-1-REST/pkg/domain/user"
	"github.com/slim-crown/issue-1-REST/pkg/http/rest"
	"log"
	"net/http"
	"regexp"
	"time"

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
	if a != nil {
		fmt.Printf(fmt.Sprintf("[%s] %s\n", time.Now().Format(time.StampMilli), format), a)
	} else {
		fmt.Printf("[%s] %s\n", time.Now().Format(time.StampMilli), format)
	}
}

func main() {
	setup := rest.Setup{}
	setup.Logger = &logger{}

	//logger := log.New(os.Stdout, "",log.Ltime)
	var db *sql.DB
	{
		var err error
		const (
			host     = "localhost"
			port     = "5432"
			dbname   = "issue#1_db"
			role     = "issue#1_dev"
			password = "password1234!@#$"
		)
		dataSourceName := fmt.Sprintf(
			`host=%s port=%s dbname='%s' user='%s' password='%s' sslmode=disable`,
			host, port, dbname, role, password)
		db, err = sql.Open(
			"postgres", dataSourceName)
		if err != nil {
			setup.Logger.Log("database connection failed because: %s", err.Error())
			return
		}
		defer db.Close()

		if err = db.Ping(); err != nil {
			setup.Logger.Log("database ping failed because: %s", err.Error())
			return
		}
	}

	services := make(map[string]interface{})
	cacheRepos := make(map[string]interface{})
	dbRepos := make(map[string]interface{})

	{
		var usrDBRepo = postgres.NewUserRepository(db, &dbRepos)
		dbRepos["User"] = &usrDBRepo
		var usrCacheRepo = memory.NewUserRepository(&usrDBRepo, &cacheRepos)
		cacheRepos["User"] = &usrCacheRepo
		setup.UserService = user.NewService(&usrCacheRepo, &services)
		services["User"] = &setup.UserService
	}
	{
		var feedDBRepo feed.Repository = postgres.NewFeedRepository(db, &dbRepos)
		dbRepos["Feed"] = &feedDBRepo
		var feedCacheRepo feed.Repository = memory.NewFeedRepository(&feedDBRepo, &cacheRepos)
		cacheRepos["Feed"] = &feedCacheRepo
		setup.FeedService = feed.NewService(&feedCacheRepo, &services)
		services["Feed"] = &setup.FeedService
	}
	{
		var releaseDBRepo = postgres.NewReleaseRepository(db, &dbRepos)
		dbRepos["Release"] = &releaseDBRepo
		var releaseCacheRepo = memory.NewReleaseRepository(&releaseDBRepo)
		cacheRepos["Release"] = &releaseCacheRepo
		setup.ReleaseService = release.NewService(&releaseCacheRepo)
		services["Release"] = &setup.ReleaseService
	}

	setup.ImageServingRoute = "/images/"
	setup.ImageStoragePath = "data/images/"
	setup.Port = "8080"
	setup.HostAddress = "http://localhost"

	setup.HostAddress += ":" + setup.Port

	setup.StrictSanitizer = bluemonday.StrictPolicy()
	setup.MarkupSanitizer = bluemonday.UGCPolicy()
	setup.MarkupSanitizer.AllowAttrs("class").Matching(regexp.MustCompile("^language-[a-zA-Z0-9]+$")).OnElements("code")

	mux := rest.NewMux(&setup)
	setup.Logger.Log("starting up server...")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
