// micro web framework. it is not very generic, but it makes writing small apps very fast
package godzilla

import (
	"database/sql"
	"github.com/jackdoe/session"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"
)

const (
	DebugQuery             = 1
	DebugQueryResult       = 2
	DebugTemplateRendering = 4
	TypeJSON               = "application/json"
	TypeHTML               = "text/html"
	TypeText               = "text/plain"
)

var (
	Debug                 int    = 0
	EnableSessions        bool   = false
	ViewDirectory         string = "v"
	NoLayoutForXHR        bool   = true
	TemplateExt           string = ".html"
	EnableStaticDirectory bool   = true
	StaticDirectory       string = "public"

	static_dir string
	views_dir  string

	template_regexp = regexp.MustCompile(".*?(\\w+)\\.(\\w+)")
	sanitize_regexp = regexp.MustCompile("[^a-zA-Z0-9_]")
)

var routes = map[*regexp.Regexp]func(*Context){}

// example: godzilla.Route("/product/show/(\\d+)",product_show)
func Route(pattern string, handler func(*Context)) {
	routes[regexp.MustCompile(pattern)] = handler
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
func Start(addr string, db *sql.DB) {

	static_dir = static_directory()
	views_dir = views_directory()

	if (EnableStaticDirectory) {
		_log("enabled static directory: " + StaticDirectory)
		Route("^/" + StaticDirectory + "/",staticRoute)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var s *session.SessionObject
		if EnableSessions {
			s = session.New(w, r)
		}
		rpath := r.URL.Path
		for k, v := range routes {
			matched := k.FindStringSubmatch(rpath)
			if matched != nil {
				_log("%s: %s @ %%r{%s}", r.RemoteAddr, rpath, k)
				params := map[string]interface{}{}
				sparams := map[string]string{}
				r.ParseForm()
				if len(r.Form) > 0 {
					for k, v := range r.Form {
						params[k] = v[0]
						sparams[k] = v[0]
					}
				}
				ctx := &Context{w, r, s, db, map[string]interface{}{}, "layout", matched, params, sparams,k}
				ctx.ContentType(TypeHTML)
				v(ctx)
				return
			}
		}
		_log("%s - NOT FOUND", rpath)
		http.NotFound(w, r)
	})
	_log("started: http://%s/", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func sanitize(s string) string {
	return sanitize_regexp.ReplaceAllString(s, "")
}
func caller(level int) string {
	pc, _, _, ok := runtime.Caller(level)
	if !ok {
		return "unknown"
	}
	me := runtime.FuncForPC(pc)
	if me == nil {
		return "unnamed"
	}
	return me.Name()
}

func staticRoute(ctx *Context) {
	rpath := ctx.Re.ReplaceAllString(ctx.R.URL.Path,"") // so we have just the filename left
	method := ctx.R.Method
	f := path.Join(static_dir, path.Clean(rpath))
	if file_exists(f) && (method == "GET" || method == "HEAD") {
		ctx.Log("FILE: %s {URI: %s}", f, rpath)
		http.ServeFile(ctx.W, ctx.R, path.Clean(f))
		return
	}
}

// taken from https://github.com/hoisie/web
func root() string {
	arg0,_ := path.Split(path.Clean(os.Args[0]))
	wd, _ := os.Getwd()
	if starts_with_slash(arg0) { return arg0 }
	return path.Join(wd, arg0)
}
func static_directory() string {
	return absolute_or_relative(StaticDirectory)
}
func views_directory() string {
	return absolute_or_relative(ViewDirectory)
}
func template_filepath(s string) string {
	s += TemplateExt
	if starts_with_slash(s) {
		return s
	}
	return path.Join(views_dir, strings.ToLower(s))
}
func absolute_or_relative(s string) string {
	if starts_with_slash(s) {
		return s
	}
	return path.Join(root(), s)
}
func starts_with_slash(s string) bool {
	return strings.HasPrefix(s, string(os.PathSeparator))
}
func file_exists(f string) bool {
	stat, err := os.Stat(f)
	return (err == nil && (!stat.IsDir()))
}
func _log(fmt string, v ...interface{}) {
	log.Printf(fmt,v...)	
}