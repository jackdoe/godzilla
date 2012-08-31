package godzilla

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/jackdoe/session"
	"net/http"
	"os"
	"reflect"
	"strings"
	"text/template"
	"regexp"
)

type Context struct {
	W       http.ResponseWriter
	R       *http.Request
	S       *session.SessionObject
	DB      *sql.DB
	O       map[string]interface{}
	Layout  string
	Splat   []string
	Params  map[string]interface{}
	Sparams map[string]string
	Re 		*regexp.Regexp
}

// renders a template, if the template name starts with os.PathSeparator it is rendered with absolute path
// otherwise it is appended to Views 
// WARNING: all template names are converted to lower case
// 
//		ctx.Render("show") // -> ./v/show.html (Views + "show" + ".html")
//		ctx.Render("/tmp/show") // -> /tmp/show.html ("/tmp/show" + ".html")
//
// if left without arguments (ctx.Render()) - it takes the package_name.function_name and renders 
//	v/package_name/function.templateExt
// so for example if we have package gallery with function Album() and we have ctx.Render() inside it
// it will render Views + /gallery/ + album + TemplateExt (default: ./v/gallery/album.html)
func (this *Context) Render(extra ...string) {
	var ROOT string
	templates := []string{}

	if len(extra) == 0 {
		c := caller(2)
		c = template_regexp.ReplaceAllString(c, "$1"+string(os.PathSeparator)+"$2")
		extra = append(extra, c)
	}

	if (NoLayoutForXHR && this.IsXHR()) || (!file_exists(template_filepath(this.Layout))) || len(this.Layout) == 0 {
		ROOT = "yield"
	} else {
		ROOT = "layout"
		templates = append(templates, template_filepath(this.Layout))
	}
	for _, v := range extra {
		templates = append(templates, template_filepath(v))
	}
	if (Debug & DebugTemplateRendering) > 0 {
		_log("loading: %#v", templates)
	}
	ts := template.New("ROOT")
	ts.Funcs(template.FuncMap{"eq": reflect.DeepEqual, "js": Template_js})
	ts = template.Must(ts.ParseFiles(templates...))
	ts.ExecuteTemplate(this.W, ROOT, this.O)
}

// returns true/false if the request is XHR
// example:
//		if ctx.IsXHR() {
//			ctx.Layout = "special-ajax-lajout" 
//			// or
//			ctx.Render("ajax")
//		}
func (this *Context) IsXHR() bool {
	v, ok := this.R.Header["X-Requested-With"]
	if ok {
		for _, val := range v {
			if strings.ToLower(val) == strings.ToLower("XMLHttpRequest") {
				return true
			}
		}
	}
	return false
}

func (this *Context) Log(format string, v ...interface{}) {
	_log(format, v...)
}
func (this *Context) Sanitize(s string) string {
	return sanitize(s)
}

func (this *Context) RenderJSON(j interface{}, error_code int) error {
	b, e := json.Marshal(j)
	if e != nil && error_code > 0 {
		this.Error(e.Error(), error_code)
		return e
	}
	this.ContentType(TypeJSON)
	this.W.Write(b)
	return e
}

// shorthand for writing strings into the http writer
// example:
//		ctx.Write("luke, i am your father")
func (this *Context) Write(s string) {
	fmt.Fprintf(this.W, "%s", s)
}

// example: 
//	ctx.Redirect("http://golang.org")
func (this *Context) Redirect(url string) {
	http.Redirect(this.W, this.R, url, 302)
}

func (this *Context) ContentType(s string) {
	this.W.Header().Set("Content-Type", s)
}

// example:
//		ctx.Error("something very very bad just happened",500)
//		// or
//		ctx.Error("something very very bad just happened",http.StatusInternalServerError)
func (this *Context) Error(message string, code int) {
	http.Error(this.W, message, code)
}