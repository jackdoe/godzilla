package main
import (
	"github.com/jackdoe/godzilla"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"regexp"
	"time"
)
func is_admin(ctx *godzilla.Context) (bool) {
	ip,_ := regexp.Match("^127\\.0\\.0\\.1:",[]byte(ctx.R.RemoteAddr))
	uri,_ := regexp.Match("^/admin/",[]byte(ctx.R.RequestURI))	
	return (ip && uri)
}
func list(ctx *godzilla.Context) {
	ctx.O["title"] = "godzilla blog!"
	ctx.O["items"] = ctx.Query("SELECT * FROM posts ORDER BY stamp DESC")
	ctx.O["is_admin"] = is_admin(ctx)
	ctx.Render("list")
}
func show(ctx *godzilla.Context) {
	err := func() { ctx.Error("nothing to do here.. \\o/",404) }
	switch ctx.Splat[1] {
		case "edit","create":
			if ! is_admin(ctx) { err(); return }
			ctx.O["title"] = ctx.Splat[1]
			o := ctx.FindById("posts",ctx.Splat[2]); 
			if (ctx.R.Method == "GET") {
				if o != nil { ctx.O["item"] = o }
				ctx.Render("form")
			} else {
				u := map[string]interface{}{
						"title":ctx.R.FormValue("title"),
						"long":ctx.R.FormValue("long"),
						"stamp":time.Now().Unix()}
				if ctx.Splat[1] != "create" {
					u["id"] = ctx.Splat[2]
				}	
				e := ctx.Replace("posts",u)
				if e != nil {
					ctx.Write(e.Error())
				} else {
					ctx.Redirect("/admin/")
				}
			}
		case "delete":
			if ! is_admin(ctx) { err() ; return }
			ctx.DB.Exec("DELETE FROM posts WHERE id=?",ctx.Splat[2])
			ctx.Redirect("/admin/")
		default:
			o := ctx.FindById("posts",ctx.Splat[1]); 
			if o == nil { err(); return }
			ctx.O["title"] = o["title"]
			ctx.O["item"] = o
			ctx.Render("show")
	}
}
func main() {
	db, _ := sql.Open("sqlite3", "./high-preformance-database.db")
	defer db.Close()
	db.Exec("CREATE TABLE IF NOT EXISTS posts (id INTEGER PRIMARY KEY,title TEXT NOT NULL,long TEXT NOT NULL,stamp INTEGER)")
	godzilla.EnableSessions = false
	godzilla.Debug = (godzilla.DebugQuery)
	godzilla.Route("^/$",list)
	godzilla.Route("^/show/(\\d+)$",show)
	godzilla.Route("^/admin/$",list)
	godzilla.Route("^/admin/show/(edit|delete|create)/(\\d+)$",show)
	godzilla.Start("localhost:8080",db)
}
