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
	Views                 string = "./v/"
	NoLayoutForXHR        bool   = true
	TemplateExt           string = ".html"
	EnableSessions        bool   = true
	EnableStaticDirectory bool   = true
	StaticDirectory       string = "./public"
	template_regexp              = regexp.MustCompile(".*?(\\w+)\\.(\\w+)")
	sanitize_regexp              = regexp.MustCompile("[^a-zA-Z0-9_]")
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
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var s *session.SessionObject
		if EnableSessions {
			s = session.New(w, r)
		}
		rpath := r.URL.Path
		if EnableStaticDirectory {
			f := path.Join(StaticDirectory, path.Clean(rpath))
			stat, err := os.Stat(f)
			if err == nil && (!stat.IsDir()) && (r.Method == "GET" || r.Method == "HEAD") {
				log.Printf("FILE: %s {URI: %s}", f, rpath)
				http.ServeFile(w, r, path.Clean(f))
				return
			}
		}
		for k, v := range routes {
			matched := k.FindStringSubmatch(rpath)
			if matched != nil {
				log.Printf("%s @ %%r{%s}", rpath, k)
				params := map[string]interface{}{}
				sparams := map[string]string{}
				r.ParseForm()
				if len(r.Form) > 0 {
					for k, v := range r.Form {
						params[k] = v[0]
						sparams[k] = v[0]
					}
				}
				ctx := &Context{w, r, s, db, map[string]interface{}{}, "layout", matched, params, sparams}
				ctx.ContentType(TypeHTML)
				v(ctx)
				return
			}
		}
		log.Printf("%s - NOT FOUND", rpath)
		http.NotFound(w, r)
	})
	log.Printf("started: http://%s/", addr)
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
