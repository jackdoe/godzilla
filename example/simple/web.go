package main
import (
	"github.com/jackdoe/session"
	"github.com/jackdoe/godzilla"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func h(ctx *godzilla.Context) {
	ctx.Output["list"] = []interface{}{"a","b"}
	ctx.Output["SessionList"] = ctx.Query("SELECT * FROM session")
	ctx.Render("sample")
}


func main() {
	db, _ := sql.Open("sqlite3", "./foo.db")
	defer db.Close()
	session.Init(db,"session")
	session.CookieKey = "go.is.awesome"
	session.CookieDomain = "localhost"
	godzilla.Route("^/(sample)/(.*)$",h)
	godzilla.Route("^/event/(create|edit|delete)/(.*)?$",h)
	godzilla.Start("localhost:8080",db)
}
