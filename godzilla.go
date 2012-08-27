package godzilla
import (
	"net/http"
	"github.com/jackdoe/session"
	"database/sql"
	"log"
	"strings"
	"regexp"
	"text/template"
)
type Context struct {
	W http.ResponseWriter
	R *http.Request
	S *session.SessionObject
	DB *sql.DB
	Output map[string]interface{}
	Layout string
	Splat []string
}
var (
	Views string = "./v/"
	NoLayoutForXHR bool = true
	TemplateExt string = ".html"
)
func (this *Context) IsXHR() bool {
	v,ok := this.R.Header["X-Requested-With"]; 
	if ok {
		for _,val := range v {
			if strings.ToLower(val) == strings.ToLower("XMLHttpRequest") {
				return true
			}
		}
	}
	return false
}

func (this *Context) Render(name string) {
	var ts *template.Template
	var err error
	gen := func(s string) string {
		return Views + s + TemplateExt
	}
	name = gen(name)
	if (NoLayoutForXHR && this.IsXHR()) || len(this.Layout) == 0 {
		ts,err = template.ParseFiles(name)
		ts.Parse(`{{template "yield" .}}`)
	} else {
		ts, err = template.ParseFiles(gen(this.Layout),name)
	}
	if err != nil {
		log.Printf("error rendering: %s - %s",name,err.Error())
	}
	ts.Execute(this.W, this.Output)
}

var routes = map[*regexp.Regexp]func(*Context)(){}

func Route(pattern string, handler func(*Context)()) {
	routes[regexp.MustCompile(pattern)]=handler
}

func Start(host string, port string,db *sql.DB) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		s := session.New(w,r)
		path := r.URL.Path
		w.Header().Set("Content-Type", "text/html")
		for k, v := range routes {
			matched := k.FindStringSubmatch(path)
			if matched != nil {
				log.Printf("%s @ %%r{%s}",path,k)
				v(&Context{w,r,s,db,make(map[string]interface{}),"layout",matched})
				return
			}
		}
		log.Printf("%s - NOT FOUND",path)
		http.NotFound(w,r)
	})
	log.Printf("started: http://%s:%s/",host,port)
	log.Fatal(http.ListenAndServe(host + ":" + port, nil))
}
