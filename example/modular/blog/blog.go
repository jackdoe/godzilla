package blog
import (
	"github.com/jackdoe/godzilla"
	"regexp"
	"time"
	"encoding/json"
	"strconv"
	"reflect"
)
func is_admin(ctx *godzilla.Context) (bool) {
	ip,_ := regexp.Match("^127\\.0\\.0\\.1:",[]byte(ctx.R.RemoteAddr))
	uri,_ := regexp.Match("^/admin/",[]byte(ctx.R.RequestURI))	
	return (ip && uri)
}

func List(ctx *godzilla.Context) {
	ctx.O["title"] = "godzilla blog!"
	ctx.O["categories"] = ctx.Query("SELECT * FROM categories ORDER BY name")
	ctx.O["selected"],_ = strconv.ParseInt(reflect.ValueOf(ctx.Params["category"]).String(),10,64)
	link := map[int64][]map[string]interface{}{}
	x := ctx.Query("SELECT b.id AS link_id, b.category_id AS cid, b.post_id AS pid,a.name as category FROM categories a,post_category b WHERE a.id=b.category_id")
	for _,v := range x {
		key := reflect.ValueOf(v["pid"]).Int()
		if link[key] == nil {
			link[key] = make([]map[string]interface{},0)
		} 
		link[key] = append(link[key],v)
	}
	ctx.O["link"] = link
	_,ok := ctx.Params["category"]; if ok {
		ctx.O["items"] = ctx.Query("SELECT a.*,(strftime('%s', 'now') - a.created_at) as ago FROM posts a, post_category b WHERE a.id = b.post_id AND b.category_id=? ORDER BY a.created_at DESC",ctx.Params["category"])
	} else {
		ctx.O["items"] = ctx.Query("SELECT *,(strftime('%s', 'now') - created_at) as ago FROM posts ORDER BY created_at DESC")
	}
	ctx.O["is_admin"] = is_admin(ctx)
	ctx.Render()
}
func Modify(ctx *godzilla.Context) {
	if ! is_admin(ctx) { ctx.Error("not allowed",404); return }

	object_id := ctx.Splat[2]
	object_type := ctx.Splat[1]
	if object_type != "posts" && object_type != "categories" && object_type != "post_category" {
		ctx.Error("bad object type",500); return
	}
	var output interface{}
	var j map[string]interface{}
	e := json.Unmarshal([]byte(ctx.Sparams["json"]),&j)
	if j == nil {
		j = map[string]interface{}{}
	}
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
func Show(ctx *godzilla.Context) {
	o := ctx.FindById("posts",ctx.Splat[1]); 
	if o == nil { ctx.Error("nothing to do here.. \\o/",404); return }
	ctx.O["title"] = o["title"]
	ctx.O["item"] = o
	ctx.Render()
}
