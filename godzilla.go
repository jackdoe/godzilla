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
	"reflect"
	"runtime"
	"path"
	"os"
	"io/ioutil"
)
type Context struct {
	W http.ResponseWriter
	R *http.Request
	S *session.SessionObject
	DB *sql.DB
	O map[string]interface{}
	Layout string
	Splat []string
	Params map[string]interface{}
	Sparams map[string]string
}
const (
	DebugQuery = 1
	DebugQueryResult = 2
	DebugTemplateRendering = 4
	TypeJSON = "application/json" //http://stackoverflow.com/questions/477816/what-format-should-i-use-for-the-right-json-content-type
	TypeHTML = "text/html"
	TypeText = "text/plain"
)
var (
	Debug int = 0
	Views string = "./v/"
	NoLayoutForXHR bool = true
	TemplateExt string = ".html"
	EnableSessions bool = true
	EnableStaticDirectory bool = true
	StaticDirectory string = "./public"
	template_regexp = regexp.MustCompile(".*?(\\w+)\\.(\\w+)")
	sanitize_regexp = regexp.MustCompile("[^a-zA-Z0-9_]")
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
		rpath := r.URL.Path
		if EnableStaticDirectory {
			f := path.Join(StaticDirectory, rpath)
			stat,err := os.Stat(f);
			log.Printf("%s",f) 
			if err == nil && (!stat.IsDir()) && (r.Method == "GET" || r.Method == "HEAD") {
				log.Printf("serving file: %s %s",f,rpath)
				b, err := ioutil.ReadFile(path.Clean(f)); if err == nil {
					w.Write(b)
				}
			}
		}
		for k, v := range routes {
			matched := k.FindStringSubmatch(rpath)
			if matched != nil {
				log.Printf("%s @ %%r{%s}",rpath,k)
				params := map[string]interface{}{}
				sparams := map[string]string{}
				r.ParseForm()
				if len(r.Form) > 0 {
					for k, v := range r.Form {
						params[k] = v[0]
						sparams[k] = v[0]
					}
				}
				ctx := &Context{w,r,s,db,map[string]interface{}{},"layout",matched,params,sparams}
				ctx.ContentType(TypeHTML)
				v(ctx)
				return
			}
		}
		log.Printf("%s - NOT FOUND",rpath)
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


// renders a template, if the template name starts with os.PathSeparator it is rendered with absolute path
// otherwise it is appended to Views 
// WARNING: all template names are converted to lower case
// example
//	ctx.Render("show") // -> ./v/show.html (Views + "show" + ".html")
//	ctx.Render("/tmp/show") // -> /tmp/show.html ("/tmp/show" + ".html")
//
// if left without arguments (ctx.Render())
// it takes the package_name.function_name
// and renders v/package_name/function.templateExt
// so for example if we have package gallery with function Album() and we have ctx.Render() inside it
// it will render Views + /gallery/ + album + TemplateExt (default: ./v/gallery/album.html)
func (this *Context) Render(extra ...string) {
	var ROOT string
	templates := []string{}	

	gen := func(s string) string {
		s += TemplateExt

		if len(s) > 0 && s[0] == os.PathSeparator { return s }
		return Views + strings.ToLower(s)
	}

	if len(extra) == 0 {
		c := caller(2)
		c = template_regexp.ReplaceAllString(c,"$1"+string(os.PathSeparator)+"$2")
		extra = append(extra,c)
	}
	if (NoLayoutForXHR && this.IsXHR()) || len(this.Layout) == 0 {
		ROOT = "yield"
	} else {
		ROOT = "layout"
		templates = append(templates,gen(this.Layout))
	}
	for _,v := range extra {
		templates = append(templates,gen(v))
	}
	if (Debug & DebugTemplateRendering) > 0 {
		log.Printf("loading: %#v",templates)
	}
	ts := template.New("ROOT")
	ts.Funcs(template.FuncMap{"eq": reflect.DeepEqual,"js": javascript_template})
	ts = template.Must(ts.ParseFiles(templates...))
	ts.ExecuteTemplate(this.W, ROOT,this.O)
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

func (this *Context) ContentType(s string) {
	this.W.Header().Set("Content-Type", s)
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


func (this *Context) FindBy(table string, field string,v interface{}) (map[string]interface{}) {
	table = sanitize(table)
	field = sanitize(field)
	o := this.Query("SELECT * FROM `"+table+"` WHERE `"+field+"`=?",v)
	if len(o) > 0 {
		return o[0]
	}
	return nil
}
func (this *Context) FindById(table string, id interface{}) (map[string]interface{}) {
	return this.FindBy(table,"id",id)
} 

func (this *Context) DeleteBy(table string, field string, v interface{}) {
	table = sanitize(table)
	field = sanitize(field)
	q := "DELETE FROM `"+table+"` WHERE `"+field+"`=?"
	if (Debug & DebugQuery) > 0 { log.Printf("%s",q) }
	this.DB.Exec(q,v)
}
func (this *Context) DeleteId(table string, id interface{}) {
	this.DeleteBy(table,"id",id)
}

// POC: bad performance
// updates database fields based on map's keys - every key that begins with _ is skipped
func (this *Context) Replace(table string,input map[string]interface{}) (int64,error) {
	table = sanitize(table)
	keys := []interface{}{}
	values := []interface{}{}
	skeys := []string{}
	for k,v := range input {
		if len(k) > 0 && k[0] != '_' {
			keys = append(keys,k)
			skeys = append(skeys,"`" + k + "`")
			values = append(values,v)
		}
	}

	questionmarks := strings.TrimRight(strings.Repeat("?,",len(skeys)),",")
	q := fmt.Sprintf("REPLACE INTO `%s` (%s) VALUES(%s)",table,strings.Join(skeys,","),questionmarks)
	if (Debug & DebugQuery) > 0 { log.Printf("%s",q) }
	if (Debug & DebugQueryResult) > 0 { log.Printf("%s: %#v",q,input) }
	res,e := this.DB.Exec(q,values...)
	if e != nil && (Debug & DebugQuery) > 0 { log.Printf("%s: %s",q,e.Error()) }
	last_id := int64(0)
	if res != nil {
		last_id,_ = res.LastInsertId()
	}
	return last_id,e
}
func (this *Context) Log(format string, v ...interface{}) {
	log.Printf(format,v...)
}
func (this *Context) Sanitize(s string) string {
	return sanitize(s)
}

func sanitize(s string) string {
	return sanitize_regexp.ReplaceAllString(s,"")
}
func caller(level int) string {
	pc, _, _, ok := runtime.Caller(level)
	if !ok { return "unknown" } 
	me := runtime.FuncForPC(pc)
	if me == nil { return "unnamed" } 
	return me.Name()
}

func javascript_template(args ...string) string {
	s := ""
	for _,v := range args {
		s += fmt.Sprintf("<script type='text/template' id='%s' src='/public/%s.js'></script><script>var %s = $('#%s').html();</script>",v,v,v,v)
	}
	return s
}
