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
	Debug bool = false
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

func Start(addr string,db *sql.DB) {
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
	log.Printf("started: http://%s/",addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// POC, bad performance, do not use in production
func (this *Context) Query(query string, args ...interface{}) []map[string]interface{} {
	var err error
	r := make([]map[string]interface{},0)
	rows,err := this.DB.Query(query,args...); 
	if err != nil {
		log.Printf("%s - %s",query,err)
		return r
	}
	columns, err := rows.Columns()
	if err != nil {
		log.Printf("%s - %s",query, err)
		return r
	}
	for rows.Next() {
		row := map[string]*interface{}{}
		fields := []interface{}{}
		for _,v := range columns {
			t := new(interface{})
			row[v] = t
			fields = append(fields,t)
		}
		err = rows.Scan(fields...)
		if err != nil {
			log.Printf("%s",err)
		} else {
			x := map[string]interface{}{}
			for k,v := range row {
				x[k] = *v
			}
			r = append(r,x)
		}
	}
	if (Debug) {
		log.Printf("extracted %d rows @ %s",len(r),query)
	}
	return r
}
