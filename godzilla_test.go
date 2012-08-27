package godzilla
import (
	"github.com/jackdoe/session"
	"database/sql"
	"net/http"
	"os"
	"testing"
	_ "github.com/mattn/go-sqlite3"
	"time"
	"io/ioutil"
	"regexp"
	"reflect"
	"net/url"
)


type myjar struct { 
	jar map[string] []*http.Cookie 
}

func (p* myjar) SetCookies(u *url.URL, cookies []*http.Cookie) { 
	p.jar[u.Host] = cookies 
}

func (p *myjar) Cookies(u *url.URL) []*http.Cookie { 
	return p.jar[u.Host] 
}

func exit(ctx *Context) {
	os.Exit(0)
}

var addr string = "localhost:65444"
var URL string = "http://" + addr + "/"

func redir_to_blabla(ctx *Context) {
	ctx.Redirect("/blabla")
}
func errorize(ctx *Context) {
	ctx.Error("errorize",500)
}

func blabla(ctx *Context) {
	ctx.Write("blabla")
}
func splat(ctx *Context) {
	ctx.Write(ctx.Splat[1] + "-" + ctx.Splat[2])
}
func sample(ctx *Context) {
	ctx.Render("sample")
}
func sample_force_no_layout(ctx *Context) {
	ctx.Layout = ""
	ctx.Render("sample")
}
func set(ctx *Context) {
	ctx.S.Set("key","value")
	x,ok := ctx.S.Get("key"); if ok {
		ctx.O["key"] = x
	}
	ctx.Render("session")
}
func get(ctx *Context) {
	x,ok := ctx.S.Get("key"); if ok {
		ctx.O["key"] = x
	}
	ctx.Render("session")
}

func clear(ctx *Context) {
	ctx.S.Set("key",nil)
	x,ok := ctx.S.Get("key"); if ok {
		ctx.O["key"] = x
	}
	ctx.Render("session")
}

func start_server(db *sql.DB) {
	go func() {
		Start(addr,db)
	}()
	time.Sleep(1) // XXX: pff..
}
func stop_server() {
	http.Get(URL + "exit")
}
var client *http.Client

func expect(t *testing.T,url string, code int,pattern interface{},ajax bool) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("%s - %s",url,err)
	}
	if ajax {
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
	}
	resp, err := client.Do(req)
	if err != nil { t.Fatalf("%s - %s",url,err)	}

	if resp.StatusCode != code {
		t.Fatalf("%s - expected %d got: %d (%s)",url,code,resp.StatusCode,resp.Status)
	}
	if pattern != nil {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil { t.Fatalf("%s - %s",url,err)	}
		r := reflect.ValueOf(pattern)
		matched,err := regexp.Match(r.String(),body)
		if err != nil { t.Fatalf("%s - %s",url,err) }
		if !matched {
			s := string(body)
			t.Fatalf("%s - EXPECT: %s GOT: %#v",url,r.String(),s)
		}
	}
}
func TestStart(t *testing.T)  {
	var err error
	db, _ := sql.Open("sqlite3", "./foo.db")
	Views = "./v/"
	os.Mkdir(Views,0770)
	client = &http.Client{}
	jar := &myjar{} 
	jar.jar = make(map[string] []*http.Cookie) 
	client.Jar = jar

	session.Init(db,"session")
	session.CookieKey = "go.is.awesome"
	session.CookieDomain = "localhost"
	Route("^/exit",exit)
	Route("^/blabla$",blabla)
	Route("^/sample$",sample)
	Route("^/sample_force_no_layout$",sample_force_no_layout)
	Route("^/set$",set)
	Route("^/clear$",set)
	Route("^/get$",get)	
	Route("^/redir_to_blabla$",redir_to_blabla)
	Route("^/errorize$",errorize)
	Route("^/splat/(\\d+)_(\\d+)$",splat)

	start_server(db)
	expect(t,URL,404,nil,false)
	expect(t,URL + "blabla",200,"blabla",false)

	/* create 2 simple templates */
	err = ioutil.WriteFile(Views + "layout" + TemplateExt, []byte(`<body>{{template "yield" .}}</body>`), 0644)
	if err != nil { t.Fatalf("%s",err)}
	err = ioutil.WriteFile(Views + "sample" + TemplateExt, []byte(`{{define "yield"}}sample{{end}}`), 0644)
	if err != nil { t.Fatalf("%s",err)}
	err = ioutil.WriteFile(Views + "session" + TemplateExt, []byte(`{{define "yield"}}{{.key}}{{end}}`), 0644)
	if err != nil { t.Fatalf("%s",err)}
	expect(t,URL + "sample",200,"^<body>sample</body>$",false)
	NoLayoutForXHR = true
	expect(t,URL + "sample",200,"^sample$",true) // should have no layout with ajax
	expect(t,URL + "sample_force_no_layout",200,"^sample$",false)
	expect(t,URL + "get",200,"",true) // nothins is set yet
	expect(t,URL + "set",200,"^value$",true) 
	expect(t,URL + "get",200,"^value$",true) 

	expect(t,URL + "clear",200,"",true)
	expect(t,URL + "get",200,"",true) // nothins is set yet
	expect(t,URL + "set",200,"^value$",true) 
	expect(t,URL + "get",200,"^value$",true) 

	expect(t,URL + "errorize",500,"^errorize",true) 
	expect(t,URL + "redir_to_blabla",200,"^blabla$",true) // redirects to /blabla
	expect(t,URL + "splat/1234_2345",200,"^1234-2345$",true)

	db.Exec("CREATE TABLE IF NOT EXISTS x (id INTEGER PRIMARY KEY,title TEXT NOT NULL,long TEXT NOT NULL,stamp INTEGER)")
	ctx := &Context{nil,nil,nil,db,make(map[string]interface{}),"layout",[]string{}}
	u := map[string]interface{}{
		"title": "zzz",
		"long": "adasdasd",
		"stamp": 0}
	err = ctx.Replace("x",u)
	if err != nil { t.Fatalf("%s",err)}
	o := ctx.Query("SELECT * FROM x")

	if len(o) != 1 { t.Fatalf("expecting 1 row in the table got: %d - %#v",len(o),o)}
	id := o[0]["id"]
	found := ctx.FindById("x",id)
	if found == nil { t.Fatalf("couldnt find %s",id)}
	if found["title"] != u["title"] { t.Fatalf("title field mismatch: %s - %s",found["title"],u["title"])}
	found["title"] = "yyy"

	err = ctx.Replace("x",found)
	if err != nil { t.Fatalf("%s",err)}
	found_again := ctx.FindById("x",found["id"])
	if found_again == nil { t.Fatalf("couldnt find %s",found["id"])}
	if found_again["title"] != found["title"] { t.Fatalf("tite field mismatch: %s - %s",found_again["title"],found["title"])}

	gen := func(s string) string {
		return Views + s + TemplateExt
	}
	f := []string{gen("layout"),gen("session"),gen("sample"), Views,"foo.db"}
	for _, file := range f {
		err = os.Remove(file)
		if err != nil {
			t.Fatalf("%s",err)
		}
	}


	stop_server()
}
