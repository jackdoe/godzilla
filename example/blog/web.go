package main
import (
	"github.com/jackdoe/godzilla"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"regexp"
	"time"
	"encoding/json"
	"strings"
)
func is_admin(ctx *godzilla.Context) (bool) {
	ip,_ := regexp.Match("^127\\.0\\.0\\.1:",[]byte(ctx.R.RemoteAddr))
	uri,_ := regexp.Match("^/admin/",[]byte(ctx.R.RequestURI))	
	return (ip && uri)
}
func list(ctx *godzilla.Context) {

	ctx.O["title"] = "godzilla blog!"
	ctx.O["categories"] = ctx.Query("SELECT * FROM categories ORDER BY name")
	_,ok := ctx.Params["category"]; if ok {
		ctx.O["items"] = ctx.Query("SELECT a.* FROM posts a, post_category b WHERE a.id = b.post_id AND b.category_id=? ORDER BY stamp DESC",ctx.Params["category"])
	} else {
		ctx.O["items"] = ctx.Query("SELECT * FROM posts ORDER BY stamp DESC")
	}
	ctx.O["is_admin"] = is_admin(ctx)
	ctx.Render("list")
}
func modify_category(ctx *godzilla.Context) {
	if ! is_admin(ctx) { ctx.Error("not allowed",404); return }
	ctx.ContentType(godzilla.TypeJSON)
	var output interface{}
	switch ctx.R.Method {
		case "GET":
			output = ctx.FindById("categories",ctx.Splat[1])
		case "POST","PUT":
			u := map[string]interface{}{"name":ctx.Params["title"]}
			if ctx.R.Method == "PUT" {
				u["id"] = ctx.Splat[1]
			}
			id,_ := ctx.Replace("categories",u)
			output = ctx.FindById("categories",id)
		case "DELETE":
			ctx.DB.Exec("DELETE FROM categories WHERE id=?",ctx.Splat[1])
			output = "deleted " + ctx.Splat[1]
		case "OPTIONS":
			output = ctx.Query("SELECT * FROM categories a ORDER BY name")
	}
	b,e := json.Marshal(output)
	if e != nil { b = []byte(e.Error()) }
	ctx.W.Write(b)
}
func show(ctx *godzilla.Context) {
	err := func() { ctx.Error("nothing to do here.. \\o/",404) }
	switch ctx.Splat[1] {
		case "edit","create":
			if ! is_admin(ctx) { err(); return }
			u := map[string]interface{}{"title":ctx.Params["title"],"long":ctx.Params["long"],"stamp":time.Now().Unix()}
			ctx.O["title"] = ctx.Splat[1]
			o := ctx.FindById("posts",ctx.Splat[2]); 
			if (ctx.R.Method == "GET") {
				if o != nil { ctx.O["item"] = o }
				ctx.Render("form")
			} else {
				if ctx.Splat[1] != "create" {
					u["id"] = ctx.Splat[2]
				}	
				_,e := ctx.Replace("posts",u)
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
	db.Exec("CREATE TABLE IF NOT EXISTS post_category (id INTEGER PRIMARY KEY,post_id BIGINT NOT NULL,category_id TEXT NOT NULL)")
	db.Exec("CREATE TABLE IF NOT EXISTS posts (id INTEGER PRIMARY KEY,title TEXT NOT NULL,long TEXT NOT NULL,stamp INTEGER)")
	db.Exec("CREATE TABLE IF NOT EXISTS categories (id INTEGER PRIMARY KEY,name TEXT NOT NULL UNIQUE)")

	godzilla.EnableSessions = false
	godzilla.Debug = (godzilla.DebugQuery)
	godzilla.Route("^/$",list)
	godzilla.Route("^/show/(\\d+)$",show)
	godzilla.Route("^/admin/category/(\\d+)$",modify_category)
	godzilla.Route("^/admin/$",list)
	godzilla.Route("^/admin/show/(edit|delete|create)/(\\d+)$",show)
	godzilla.Start("localhost:8080",db)
}
