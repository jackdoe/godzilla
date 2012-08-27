package main
import (
	"github.com/jackdoe/godzilla"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"regexp"
	"time"
)
func is_admin(ctx *godzilla.Context) (bool) {
	ip,_ := regexp.Match("127.0.0.1:",[]byte(ctx.R.RemoteAddr))
	uri,_ := regexp.Match("^/admin/",[]byte(ctx.R.RequestURI))	
	return (ip && uri)
}
func list(ctx *godzilla.Context) {
	ctx.Output["title"] = "godzilla blog!"
	ctx.Output["items"] = ctx.Query("SELECT * FROM posts ORDER BY stamp DESC")
	ctx.Output["is_admin"] = is_admin(ctx)
	ctx.Render("list")
}
func find_by_id(ctx *godzilla.Context,id interface{}) ([]map[string]interface{},int) {
	o := ctx.Query("SELECT * FROM posts WHERE id=? ORDER BY stamp DESC",id)
	return o,len(o)
}
func show(ctx *godzilla.Context) {
	err := func() {
		ctx.Error("nothing to do here.. \\o/",404)
	}
	switch ctx.Splat[1] {
		case "edit","create":
			if ! is_admin(ctx) { err() ; return }
			ctx.Output["title"] = ctx.Splat[1]
			o,l := find_by_id(ctx,ctx.Splat[2]);
			if (ctx.R.Method == "GET") {
				if (l == 1) { ctx.Output["item"] = o[0] }
				ctx.Render("form")
			} else {
				values := []interface{}{ctx.R.FormValue("title"),ctx.R.FormValue("long"),time.Now().Unix()}
				var e error
				if ctx.Splat[1] == "create" {
					_,e = ctx.DB.Exec("INSERT INTO posts (title,long,stamp) VALUES(?,?,?)",values...)
				} else {
					values = append(values,ctx.Splat[2])
					_,e = ctx.DB.Exec("REPLACE INTO posts (title,long,stamp,id) VALUES(?,?,?,?)",values...)
				}
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
			o,l := find_by_id(ctx,ctx.Splat[1]); 
			if l != 1 { err() ; return }
			ctx.Output["title"] = o[0]["title"]
			ctx.Output["item"] = o[0]
			ctx.Render("show")
	}
}
func main() {
	db, _ := sql.Open("sqlite3", "./high-preformance-database.db")
	defer db.Close()
	db.Exec("CREATE TABLE IF NOT EXISTS posts (id INTEGER PRIMARY KEY,title TEXT NOT NULL,long TEXT NOT NULL,stamp INTEGER)")
	godzilla.EnableSessions = false
	godzilla.Route("^/$",list)
	godzilla.Route("^/show/(\\d+)$",show)
	godzilla.Route("^/admin/$",list)
	godzilla.Route("^/admin/show/(edit|delete|create)/(\\d+)$",show)
	godzilla.Start("localhost:8080",db)
}
