package main

import (
	"database/sql"
	"fmt"
	"github.com/microcosm-cc/bluemonday"
	"github.com/slim-crown/issue-1-REST/pkg/domain/channel"
	"github.com/slim-crown/issue-1-REST/pkg/domain/comment"
	"github.com/slim-crown/issue-1-REST/pkg/domain/feed"
	"github.com/slim-crown/issue-1-REST/pkg/domain/post"
	"github.com/slim-crown/issue-1-REST/pkg/domain/release"
	"github.com/slim-crown/issue-1-REST/pkg/domain/user"
	"github.com/slim-crown/issue-1-REST/pkg/http/rest"
	"log"
	"net/http"
	"os"
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
/*
type logger struct{}

// Log ...
func (logger *logger) Log(format string, a ...interface{}) {
	if a != nil {
		fmt.Printf(fmt.Sprintf("[%s] %s\n", time.Now().Format(time.StampMilli), format), a)
	} else {
		fmt.Printf("[%s] %s\n", time.Now().Format(time.StampMilli), format)
	}
}
*/

func main() {
	setup := rest.Setup{}
	setup.Logger = log.New(os.Stdout, "", log.Lmicroseconds|log.Lshortfile)
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
			setup.Logger.Printf("database connection failed because: %s", err.Error())
			return
		}
		defer db.Close()

		if err = db.Ping(); err != nil {
			setup.Logger.Printf("database ping failed because: %s", err.Error())
			return
		}
	}

	services := make(map[string]interface{})
	cacheRepos := make(map[string]interface{})
	dbRepos := make(map[string]interface{})
	{
		var channelDBRepo = postgres.NewChannelRepository(db, &dbRepos)
		dbRepos["Channel"] = &channelDBRepo
		var channelCacheRepo = memory.NewChannelRepository(&channelDBRepo, &cacheRepos)
		cacheRepos["Channel"] = &channelCacheRepo
		setup.ChannelService = channel.NewService(&channelCacheRepo, &services)
		services["Channel"] = &setup.ChannelService
	}
	{
		var usrDBRepo = postgres.NewUserRepository(db, &dbRepos)
		dbRepos["User"] = &usrDBRepo
		var usrCacheRepo = memory.NewUserRepository(&usrDBRepo, &cacheRepos)
		cacheRepos["User"] = &usrCacheRepo
		setup.UserService = user.NewService(&usrCacheRepo, &services)
		services["User"] = &setup.UserService
	}
	{
		var feedDBRepo = postgres.NewFeedRepository(db, &dbRepos)
		dbRepos["Feed"] = &feedDBRepo
		var feedCacheRepo = memory.NewFeedRepository(&feedDBRepo, &cacheRepos)
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
	{
		var postDBRepo = postgres.NewPostRepository(db, &dbRepos)
		dbRepos["Post"] = &postDBRepo
		var postCacheRepo = memory.NewPostRepository(&postDBRepo)
		cacheRepos["Post"] = &postCacheRepo
		setup.PostService = post.NewService(&postCacheRepo)
		services["Post"] = &setup.PostService
	}
	{
		var commentDBRepo = postgres.NewRepository(db, &dbRepos)
		dbRepos["Comment"] = &commentDBRepo
		var commentCacheRepo = memory.NewRepository(&commentDBRepo)
		cacheRepos["Comment"] = &commentCacheRepo
		setup.CommentService = comment.NewService(&commentCacheRepo)
		services["Comment"] = &setup.CommentService
	}

	setup.ImageServingRoute = "/images/"
	setup.ImageStoragePath = "data/images/"
	setup.HostAddress = "http://localhost"
	setup.Port = "8080"

	setup.HostAddress += ":" + setup.Port

	setup.StrictSanitizer = bluemonday.StrictPolicy()
	setup.MarkupSanitizer = bluemonday.UGCPolicy()
	setup.MarkupSanitizer.AllowAttrs("class").Matching(regexp.MustCompile("^language-[a-zA-Z0-9]+$")).OnElements("code")

	setup.TokenSigningSecret = []byte("secret")
	setup.TokenAccessLifetime = 15 * time.Minute
	setup.TokenRefreshLifetime = 7 * 24 * time.Hour

	mux := rest.NewMux(&setup)

	setup.Logger.Printf("server running...")

	//log.Fatal(http.ListenAndServeTLS(":"+setup.Port, "cmd/server/cert.pem", "cmd/server/key.pem",mux))

	log.Fatal(http.ListenAndServe(":"+setup.Port, mux))
}
