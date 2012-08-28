package main
import (
	"github.com/jackdoe/godzilla"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"regexp"
	"time"
	"encoding/json"
	"strconv"
	"reflect"
	"log"
)
func is_admin(ctx *godzilla.Context) (bool) {
	ip,_ := regexp.Match("^127\\.0\\.0\\.1:",[]byte(ctx.R.RemoteAddr))
	uri,_ := regexp.Match("^/admin/",[]byte(ctx.R.RequestURI))	
	return (ip && uri)
}
func list(ctx *godzilla.Context) {

	ctx.O["title"] = "godzilla blog!"
	ctx.O["categories"] = ctx.Query("SELECT * FROM categories ORDER BY name")
	ctx.O["selected"],_ = strconv.ParseInt(reflect.ValueOf(ctx.Params["category"]).String(),10,64)
	_,ok := ctx.Params["category"]; if ok {
		ctx.O["items"] = ctx.Query("SELECT a.* FROM posts a, post_category b WHERE a.id = b.post_id AND b.category_id=? ORDER BY created_at DESC",ctx.Params["category"])
	} else {
		ctx.O["items"] = ctx.Query("SELECT * FROM posts ORDER BY created_at DESC")
	}
	ctx.O["is_admin"] = is_admin(ctx)
	ctx.Render("list")
}
func modify(ctx *godzilla.Context) {
	if ! is_admin(ctx) { ctx.Error("not allowed",404); return }

	object_id := ctx.Splat[2]
	object_type := ctx.Splat[1]
	if object_type != "posts" && object_type != "categories" {
		ctx.Error("bad object type",500); return
	}
	var output interface{}
	var j map[string]interface{}
	e := json.Unmarshal([]byte(ctx.Sparams["json"]),&j)
	if j == nil {
		j = map[string]interface{}{}
	}
	if e != nil {
		log.Printf("%s",e.Error())
	}
	log.Printf("%#v",j)
	switch ctx.R.Method {
		case "PATCH":
			output = ctx.Query("SELECT a.* FROM categories a, post_category b WHERE a.id = b.category_id AND b.post_id=?",object_id)
		case "GET":
			output = ctx.FindById(object_type,object_id)
		case "POST":
			if j["id"] == nil {
				j["created_at"] = time.Now().Unix()
			}
			j["updated_at"] = time.Now().Unix()
			id,_ := ctx.Replace(object_type,j)
			output = ctx.FindById(object_type,id)
		case "DELETE":
			ctx.DeleteId(object_type,object_id)
			output = "deleted " + object_id + "@" + object_type
		case "OPTIONS":
			output = ctx.Query("SELECT * FROM `" + object_type+ "`")//; ORDER BY created_at,updated_at")
	}
	b,e := json.Marshal(output)
	if e != nil { b = []byte(e.Error()) }
	ctx.ContentType(godzilla.TypeJSON)
	ctx.W.Write(b)
}
func show(ctx *godzilla.Context) {
	o := ctx.FindById("posts",ctx.Splat[1]); 
	if o == nil { ctx.Error("nothing to do here.. \\o/",404); return }
	ctx.O["title"] = o["title"]
	ctx.O["item"] = o
	ctx.Render("show")
}
func main() {
	db, _ := sql.Open("sqlite3", "./high-preformance-database.db")
	defer db.Close()
	db.Exec("CREATE TABLE IF NOT EXISTS post_category (post_id BIGINT NOT NULL,category_id TEXT NOT NULL,CONSTRAINT uc_post_category UNIQUE (post_id,category_id))")
	db.Exec("CREATE TABLE IF NOT EXISTS posts (id INTEGER PRIMARY KEY,title TEXT NOT NULL,long TEXT NOT NULL,created_at INTEGER,updated_at INTEGER)")
	db.Exec("CREATE TABLE IF NOT EXISTS categories (id INTEGER PRIMARY KEY,name TEXT NOT NULL UNIQUE,created_at INTEGER,updated_at INTEGER)")

	godzilla.EnableSessions = false
	godzilla.Debug = (godzilla.DebugQuery)
	godzilla.Route("^/$",list)
	godzilla.Route("^/show/(\\d+)$",show)
	godzilla.Route("^/admin/modify/(posts|categories)/(\\d+)$",modify)
	godzilla.Route("^/admin/$",list)
	godzilla.Route("^/admin/show/(edit|delete|create)/(\\d+)$",show)
	godzilla.Start("localhost:8080",db)
}
