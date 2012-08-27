// micro web framework. it is not very generic, but it makes writing small apps very fast
package godzilla
import (
	"net/http"
	"github.com/jackdoe/session"
	"database/sql"
	"log"
	"strings"
	"regexp"
	"text/template"
	"fmt"
)
type Context struct {
	W http.ResponseWriter
	R *http.Request
	S *session.SessionObject
	DB *sql.DB
	O map[string]interface{}
	Layout string
	Splat []string
}
const (
	DebugQuery = 1
	DebugQueryResult = 2
)
var (
	Debug int = 0
	Views string = "./v/"
	NoLayoutForXHR bool = true
	TemplateExt string = ".html"
	EnableSessions bool = true
)

var routes = map[*regexp.Regexp]func(*Context)(){}

// example: godzilla.Route("/product/show/(\\d+)",product_show)
func Route(pattern string, handler func(*Context)()) {
	routes[regexp.MustCompile(pattern)]=handler
}

// starts the http server
// example: 
// 		db, _ := sql.Open("sqlite3", "./foo.db")
// 		defer db.Close()
// 		session.Init(db,"session")
// 		session.CookieKey = "go.is.awesome"
// 		session.CookieDomain = "localhost"
// 		godzilla.Route("/product/show/(\\d+)",product_show)
// 		godzilla.Start("localhost:8080",db)
func Start(addr string,db *sql.DB) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var s *session.SessionObject
		if EnableSessions {
			s = session.New(w,r)
		}
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

// returns true/false if the request is XHR
// example:
//		if ctx.IsXHR() {
//			ctx.Layout = "special-ajax-lajout" 
//			// or
//			ctx.Render("ajax")
//		}
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
	ts.Execute(this.W, this.O)
}

// shorthand for writing strings into the http writer
// example:
//		ctx.Write("luke, i am your father")
func (this *Context) Write(s string) {
	fmt.Fprintf(this.W,"%s",s)
}

// example: 
//	ctx.Redirect("http://golang.org")
func (this *Context) Redirect(url string) {
	http.Redirect(this.W,this.R,url,302)
}

// example:
//		ctx.Error("something very very bad just happened",500)
//		// or
//		ctx.Error("something very very bad just happened",http.StatusInternalServerError)
func (this *Context) Error(message string, code int) {
	http.Error(this.W,message,code)
}

// WARNING: POC, bad performance, do not use in production.
// 
// Returns slice of map[query_result_fields]query_result_values,
// so for example table with fields id,data,stamp will return
// [{id: xx,data: xx, stamp: xx},{id: xx,data: xx,stamp: xx}]
// example:
// 		ctx.O["SessionList"] = ctx.Query("SELECT * FROM session")
// and then in the template:
// 	{{range .SessionList}}
//		id: {{.id}}<br>
//		data: {{.data}}<br>
//		stamp: {{.stamp}}
//	{{end}}
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
	if (Debug & DebugQuery) > 0 { log.Printf("extracted %d rows @ %s",len(r),query) }
	if (Debug & DebugQueryResult) > 0 { log.Printf("%s: %#v",query,r) }
	return r
}

func (this *Context) FindById(table string, id interface{}) (map[string]interface{}) {
	o := this.Query("SELECT * FROM `"+table+"` WHERE id=?",id)
	if len(o) > 0 {
		return o[0]
	}
	return nil
} 

// POC: bad performance
func (this *Context) Replace(table string,input map[string]interface{}) (error) {
	keys := []interface{}{}
	values := []interface{}{}
	skeys := []string{}
	for k,v := range input {
		keys = append(keys,k)
		skeys = append(skeys,"`" + k + "`")
		values = append(values,v)
	}

	questionmarks := strings.TrimRight(strings.Repeat("?,",len(skeys)),",")
	q := fmt.Sprintf("REPLACE INTO `%s` (%s) VALUES(%s)",table,strings.Join(skeys,","),questionmarks)
	if (Debug & DebugQuery) > 0 { log.Printf("%s",q) }
	if (Debug & DebugQueryResult) > 0 { log.Printf("%s: %#v",q,input) }
	_,e := this.DB.Exec(q,values...)
	return e
}