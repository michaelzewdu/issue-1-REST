package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/slim-crown/Issue-1/pkg/domain/user"

	"github.com/slim-crown/Issue-1/pkg/storage/memory"

	"github.com/slim-crown/Issue-1/pkg/http/rest"
	"github.com/slim-crown/Issue-1/pkg/storage/postgres"

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

	db, err := sql.Open(
		"postgres",
		"user='issue#1_dev' "+
			"password='password1234!@#$' "+
			"dbname='issue#1' "+
			"sslmode=disable")
	if err != nil {
		panic(fmt.Errorf("database connection failed because: %s", err.Error()))
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		panic(fmt.Errorf("database ping failed because: %s", err.Error()))
	}

	services := make(map[string]interface{})
	cacheRepos := make(map[string]interface{})
	dbRepos := make(map[string]interface{})
	var logger rest.Logger = &logger{}

	{
		var usrDBRepo user.Repository = postgres.NewUserRepository(db, &dbRepos)
		dbRepos["User"] = &usrDBRepo
		var usrCacheRepo user.Repository = memory.NewUserRepository(&usrDBRepo, &cacheRepos)
		cacheRepos["User"] = &usrCacheRepo
		var usrService user.Service = user.NewService(&usrCacheRepo, &services)
		services["User"] = &usrService
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
	http.ListenAndServe(":8080", mux)
}
