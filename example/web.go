package main
import (
	"github.com/jackdoe/session"
	"github.com/jackdoe/godzilla"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

func h(ctx *godzilla.Context) {
	ctx.Query(ctx,"SELECT * FROM session")
	ctx.Output["list"] = []interface{}{"a","b"}
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
	godzilla.Start("localhost","8080",db)
}
