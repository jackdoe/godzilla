package godzilla

import (
	"database/sql"
	"github.com/jackdoe/session"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"reflect"
	"regexp"
	"testing"
	"time"
)

type myjar struct {
	jar map[string][]*http.Cookie
}

func (p *myjar) SetCookies(u *url.URL, cookies []*http.Cookie) {
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
	ctx.Error("errorize", 500)
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
func sample_no_template_name(ctx *Context) {
	ctx.Render()
}

func sample_force_no_layout(ctx *Context) {
	ctx.Layout = ""
	ctx.Render("sample")
}
func set(ctx *Context) {
	ctx.S.Set("key", "value")
	x, ok := ctx.S.Get("key")
	if ok {
		ctx.O["key"] = x
	}
	ctx.Render("session")
}
func get(ctx *Context) {
	x, ok := ctx.S.Get("key")
	if ok {
		ctx.O["key"] = x
	}
	ctx.Render("session")
}

func clear(ctx *Context) {
	ctx.S.Set("key", nil)
	x, ok := ctx.S.Get("key")
	if ok {
		ctx.O["key"] = x
	}
	ctx.Render("session")
}

func start_server(db *sql.DB) {
	go func() {
		Start(addr, db)
	}()
	time.Sleep(1000000000) // XXX: pff..
}
func stop_server() {
	http.Get(URL + "exit")
}

var client *http.Client

func expect(t *testing.T, url string, code int, pattern interface{}, ajax bool) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("%s - %s", url, err)
	}
	if ajax {
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("%s - %s", url, err)
	}

	if resp.StatusCode != code {
		t.Fatalf("%s - expected %d got: %d (%s)", url, code, resp.StatusCode, resp.Status)
	}
	if pattern != nil {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("%s - %s", url, err)
		}
		r := reflect.ValueOf(pattern)
		matched, err := regexp.Match(r.String(), body)
		if err != nil {
			t.Fatalf("%s - %s", url, err)
		}
		if !matched {
			s := string(body)
			t.Fatalf("%s - EXPECT: %s GOT: %#v", url, r.String(), s)
		}
	}
}
func TestStart(t *testing.T) {
	cleanup()
	var err error
	db, _ := sql.Open("sqlite3", "./foo.db")
	client = &http.Client{}
	jar := &myjar{}
	jar.jar = make(map[string][]*http.Cookie)
	client.Jar = jar
	EnableSessions = true
	session.Init(db, "session")
	session.CookieKey = "go.is.awesome"
	session.CookieDomain = "localhost"
	Route("^/exit", exit)
	Route("^/blabla$", blabla)
	Route("^/sample$", sample)
	Route("^/sample_no_template_name$", sample_no_template_name)
	Route("^/sample_force_no_layout$", sample_force_no_layout)
	Route("^/set$", set)
	Route("^/clear$", set)
	Route("^/get$", get)
	Route("^/redir_to_blabla$", redir_to_blabla)
	Route("^/errorize$", errorize)
	Route("^/splat/(\\d+)_(\\d+)$", splat)
	ViewDirectory = path.Join(os.TempDir(), "godzilla-test-view-directory")
	start_server(db)
	expect(t, URL, 404, nil, false)
	expect(t, URL+"blabla", 200, "blabla", false)
	os.Mkdir(ViewDirectory, 0700)
	/* create 2 simple templates */
	err = ioutil.WriteFile(template_filepath("layout"), []byte(`{{define "layout"}}<body>{{template "yield" .}}</body>{{end}}`), 0644)
	if err != nil {
		t.Fatalf("%s", err)
	}
	err = ioutil.WriteFile(template_filepath("sample"), []byte(`{{define "yield"}}sample{{end}}`), 0644)
	if err != nil {
		t.Fatalf("%s", err)
	}
	// err = ioutil.WriteFile(Views + "sample_no_template_name" + TemplateExt, []byte(`{{define "yield"}}sample_no_template_name{{end}}`), 0644)
	// if err != nil { t.Fatalf("%s",err)}
	err = ioutil.WriteFile(template_filepath("session"), []byte(`{{define "yield"}}{{.key}}{{end}}`), 0644)
	if err != nil {
		t.Fatalf("%s", err)
	}
	expect(t, URL+"sample", 200, "^<body>sample</body>$", false)
	NoLayoutForXHR = true
	expect(t, URL+"sample", 200, "^sample$", true) // should have no layout with ajax
	// expect(t,URL + "sample_no_template_name",200,"^sample_no_template_name$",true) 
	expect(t, URL+"sample_force_no_layout", 200, "^sample$", false)
	expect(t, URL+"get", 200, "", true) // nothins is set yet
	expect(t, URL+"set", 200, "^value$", true)
	expect(t, URL+"get", 200, "^value$", true)

	expect(t, URL+"clear", 200, "", true)
	expect(t, URL+"get", 200, "", true) // nothins is set yet
	expect(t, URL+"set", 200, "^value$", true)
	expect(t, URL+"get", 200, "^value$", true)

	expect(t, URL+"errorize", 500, "^errorize", true)
	expect(t, URL+"redir_to_blabla", 200, "^blabla$", true) // redirects to /blabla
	expect(t, URL+"splat/1234_2345", 200, "^1234-2345$", true)

	db.Exec("CREATE TABLE IF NOT EXISTS x (id INTEGER PRIMARY KEY,title TEXT NOT NULL,long TEXT NOT NULL,stamp INTEGER)")
	ctx := &Context{}
	ctx.DB = db
	u := map[string]interface{}{
		"title": "zzz",
		"long":  "adasdasd",
		"stamp": 0}
	last_id, err := ctx.Replace("x", u)
	if err != nil {
		t.Fatalf("%s", err)
	}
	o := ctx.Query("SELECT * FROM x")

	if len(o) != 1 {
		t.Fatalf("expecting 1 row in the table got: %d - %#v", len(o), o)
	}
	id := o[0]["id"]
	if last_id != id {
		t.Fatalf("last_insert_id(%d) != id(%d)", last_id, id)
	}
	found := ctx.FindById("x", id)
	if found == nil {
		t.Fatalf("couldnt find %s", id)
	}
	if found["title"] != u["title"] {
		t.Fatalf("title field mismatch: %s - %s", found["title"], u["title"])
	}
	found["title"] = "yyy"

	_, err = ctx.Replace("x", found)
	if err != nil {
		t.Fatalf("%s", err)
	}
	found_again := ctx.FindById("x", found["id"])
	if found_again == nil {
		t.Fatalf("couldnt find %s", found["id"])
	}
	if found_again["title"] != found["title"] {
		t.Fatalf("tite field mismatch: %s - %s", found_again["title"], found["title"])
	}
	last_id, err = ctx.Replace("x", u)
	if last_id == 0 {
		t.Fatalf("expecting last_insert_id != 0")
	}
	// sqlite does not care for type swap
	// found["stamp"] = "ssss"
	// err = ctx.Replace("x",found)
	// if err == nil { t.Fatalf("expected error when updating int with string %#v",ctx.FindById("x",found["id"]))}

	// test string sanitize
	bad := []string{"^^^x", "&&&x&&&", "^%x", "ƒƒåß∂®xƒ∆å∆∆ß∂"}
	for _, v := range bad {
		sanitized := sanitize(v)
		if sanitized != "x" {
			t.Fatalf("expecting x, got: %s", sanitized)
		}
	}

	cleanup()
	stop_server()
}
func cleanup() {
	f := []string{template_filepath("layout"), template_filepath("session"), template_filepath("sample"), views_dir, "./foo.db"}
	for _, file := range f {
		os.Remove(file)
	}
}
