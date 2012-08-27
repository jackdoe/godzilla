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
func blabla(ctx *Context) {
	ctx.Write("blabla")
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
		ctx.Output["key"] = x
	}
	ctx.Render("session")
}
func get(ctx *Context) {
	x,ok := ctx.S.Get("key"); if ok {
		ctx.Output["key"] = x
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
	defer os.RemoveAll(Views)
	defer os.Remove("./foo.db")
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
	Route("^/get$",get)	
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
	stop_server()
}
