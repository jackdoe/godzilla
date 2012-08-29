package main
import (
	"github.com/jackdoe/godzilla"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"./url"
	"./blog"
)
func main() {
	db, _ := sql.Open("sqlite3", "./high-preformance-database.db")
	defer db.Close()
	db.Exec("CREATE TABLE IF NOT EXISTS post_category (id INTEGER PRIMARY KEY,post_id BIGINT NOT NULL,category_id BIGINT NOT NULL,created_at INTEGER,updated_at INTEGER,CONSTRAINT uc_post_category UNIQUE (post_id,category_id))")
	db.Exec("CREATE TABLE IF NOT EXISTS posts (id INTEGER PRIMARY KEY,title TEXT NOT NULL,long TEXT NOT NULL,created_at INTEGER,updated_at INTEGER)")
	db.Exec("CREATE TABLE IF NOT EXISTS categories (id INTEGER PRIMARY KEY,name TEXT NOT NULL UNIQUE,created_at INTEGER,updated_at INTEGER)")
	db.Exec("CREATE TABLE IF NOT EXISTS url (id INTEGER PRIMARY KEY,url TEXT NOT NULL UNIQUE)")

	godzilla.EnableSessions = false
	godzilla.Debug = (godzilla.DebugQuery | godzilla.DebugTemplateRendering)



	godzilla.Route("^/$",blog.List)
	godzilla.Route("^/admin/$",blog.List)
	godzilla.Route("^/admin/modify/(posts|categories|post_category)/(\\d+)$",blog.Modify)
	godzilla.Route("^/show/(\\d+)$",blog.Show)


	godzilla.Route("^/url/(\\d+)",url.Redirect)
	godzilla.Route("^/url/append/(.*)",url.Append)

	godzilla.Start("localhost:8080",db)
}
