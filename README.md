## just try it out
```
$ go get github.com/jackdoe/godzilla
$ git clone https://github.com/jackdoe/godzilla.git
$ cd godzilla/example/blog && go build main.go && ./main
```

open http://localhost:8080 and enjoy :) (/admin/ is for the admin panel)

***
check out the exampe directory for some apps - there is as simple blog, simple modular app with blog and url shortener, and one small sample app (does nothing usefull)

## soo, this is how it looks

first we will create sqlite3 database

```
sqlite3 ./foo.db
CREATE TABLE IF NOT EXISTS posts (id INTEGER PRIMARY KEY,data TEXT NOT NULL);
INSERT INTO posts(data) VALUES('godzilla was here');
INSERT INTO posts(data) VALUES('godzilla left');
```

then we will write the main package 

./main.go:
```
package main
import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "github.com/jackdoe/godzilla"
)
func list(ctx *godzilla.Context) {
    ctx.O["posts"] = ctx.Query("SELECT * FROM posts")
    ctx.Render()
}

func main() {
    db, _ := sql.Open("sqlite3", "./foo.db")
    defer db.Close()
    godzilla.Route("^/$", list)
    godzilla.Start("localhost:8080", db)
}
```

and the list view

./v/main/list.html:
```
{{define "yield"}}
    <ul>
        {{range .posts}}
            <li>
                {{.data}}
            </li>
        {{end}}
    </ul>
{{end}}
```

and you are set to go :) compile `go build main.go && ./main` and open http://localhost:8080 to see the product of your code
this is actualy in the `example/simple` directory [https://github.com/jackdoe/godzilla/tree/master/example/simple](https://github.com/jackdoe/godzilla/tree/master/example/simple)

### random blabering

i think that using the modular approach is very nice,here is a sample directory structure:
```
app/
app/main.go - import ("./blog" "./shortener" "./gallery")
app/blog/blog.go - has functions List() and Show()
app/shortener/shortener.go - has functions Redirect() and Append()
app/gallery/gallery.go - has functions Albums(), Album(), and Picture()
app/v/ (views)
app/v/blog/show.html
app/v/blog/list.html
app/v/gallery/albums.html
app/v/gallery/album.html
app/v/gallery/picture.html
```

now lets create some routes:
```
func main() {
    db, _ := sql.Open("sqlite3", "./lite.db")
    defer db.Close()
    godzilla.EnableSessions = false
    godzilla.Route("^/$",blog.List)
    godzilla.Route("^/show/(\\d+)$",blog.Show)
    godzilla.Route("^/gallery/albums$",gallery.Albums)
    godzilla.Route("^/gallery/album/(\\d+)$",gallery.Album)
    godzilla.Route("^/gallery/picture/(\\d+)$",gallery.Picture)
    godzilla.Route("^/url/(\\d+)",shortener.Redirect)
    godzilla.Route("^/url/append/(.*)",shortener.Append)
    godzilla.Start("localhost:8080",db)
}

```
if we access _/url/append/http://google.com_ godzilla's router will match _"^/url/append/(.*)"_
and call url.Append(ctx) with one argument of type *Context, this is how it could look like:
```
func Append(ctx *godzilla.Context) {
    id,_ := ctx.Replace("url",map[string]interface{}{"url":ctx.Splat[1]})
    ctx.Write(fmt.Sprintf("/url/%d",id))
}
```
so after accessing _/url/append/http://google.com_ you will see "/url/%d" (for example /url/44)
and if you access _/url/44_ the router will call url.Rediret() with *Context
```
func Redirect(ctx *godzilla.Context) {
    u := ctx.FindById("url",ctx.Splat[1])
    if u == nil {
        ctx.Error("not found",404)
    } else {
        ctx.Redirect(reflect.ValueOf(u["url"]).String())
    }
}
```

## Context
the whole concept is very simple, every function is given context as an argument
in this context you have link to the SQL database, basic database functions, http request and writer, and more
```
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
```
lets go over the fields one by one and imagine we are accessing the object in a function called **Show(ctx *godzilla.Context)** in module **blog**

* ctx.W - is of type http.ResponseWriter we can use it to write all kids of stuff to it(headers, body, etc..) [http://golang.org/pkg/net/http/#ResponseWriter](http://golang.org/pkg/net/http/#ResponseWriter)
* ctx.R - is the http.Request [http://golang.org/pkg/net/http/#Request](http://golang.org/pkg/net/http/#Request)
* ctx.S - the SessionObject - ctx.S.Get('is_user_logged_in') [http://github.com/jackdoe/session](http://github.com/jackdoe/session)
* ctx.DB - database/sql [http://golang.org/pkg/database/sql/](http://golang.org/pkg/database/sql/) [sql driver list](http://code.google.com/p/go-wiki/source/browse/SQLDrivers.wiki?repo=wiki)
* ctx.O - output map, used as templates argument when called ctx.Render()
* ctx.Layout - layout template (can be empty)
* ctx.Splat - this is where regexp matched results from the route comes - `/url/(\\d+)` - `/url/1` will put in Splat `[]{"/url/1","1"}` so `ctx.Splat[1]` is `(\\d+)` 
* ctx.Params - simple mapped parameters POST+GET /?a=4 will have `map[string]interface{}{"a":"4"}`
* ctx.Sparams - same as Params but the value is of type string instead interface{} so /?a=4 will have `map[string]string{"a":"4"}`

## ctx.Render()
    
### javascript templates
`{{ js "calendar_cell"}}` will generate the following code:

```
<script type='text/template' id='template_calendar_cell'>
//actual calendar_cell.js content
<ul>
    <% _.each(events,function(e) { %>
        <li>
            <%= e.title%>
        </li>
    <% }) %>
</ul>
</script>
<script>
var calendar_cell = $('#template_calendar_cell').html();
</script>
```

and later we can just use

`$('#cell_5').html(_.template(calendar_cell,{events: [{title: 'start learning Go'},{title: 'more Go'}]}))`


### SQL
currently we have 3-4 simple SQL helper functions that return `map[string]interface{}`
`ctx.Query("SELECT id,title,stamp FROM events")` for example will return `[]map[string]interface{}` that looks like this:

```
[
    {
        "id": 5
        "title": "very nice title"
        "stamp": 48374729
    },
    {
        "id": 6
        "title": "true story"
        "stamp": 48374729
    }
]
```

this type of hash can be used nicely in the templates for example:
```
in main.go:

func list_events(ctx *Context) {
    ctx.O["events"] = ctx.Query("SELECT id,title,stamp FROM events")
    ctx.Render() // if render is without arguments it will render views/module_name/function_name.html
}

in ./v/main/list_events.html

{{define "yield"}}
    <ul>
    {{range .events}}
        <li>
            <a href="/show/{{.id}}">
                {{.title}} - {{.stamp}}
            </a>
        </li>
    {{end}}
    </ul>
{{end}}

```

### godoc

```
PACKAGE

package godzilla
    import "github.com/jackdoe/godzilla"

    micro web framework. it is not very generic, but it makes writing small
    apps very fast

CONSTANTS

const (
    DebugQuery             = 1
    DebugQueryResult       = 2
    DebugTemplateRendering = 4
    TypeJSON               = "application/json"
    TypeHTML               = "text/html"
    TypeText               = "text/plain"
)


VARIABLES

var (
    Debug                 int    = 0
    ViewDirectory         string = "v"
    NoLayoutForXHR        bool   = true
    TemplateExt           string = ".html"
    EnableSessions        bool   = true
    EnableStaticDirectory bool   = true
    StaticDirectory       string = "public"
)


FUNCTIONS

func Route(pattern string, handler func(*Context))
    example: godzilla.Route("/product/show/(\\d+)",product_show)

func Start(addr string, db *sql.DB)
    starts the http server example:

    db, _ := sql.Open("sqlite3", "./foo.db")
    defer db.Close()
    session.Init(db,"session")
    session.CookieKey = "go.is.awesome"
    session.CookieDomain = "localhost"
    godzilla.Route("/product/show/(\\d+)",product_show)
    godzilla.Start("localhost:8080",db)

func Template_js(args ...string) string
    {{ js "calendar_cell" "calendar_row"}} will read
    /static/calendar_cell.js and generate

    <script type='text/template' id='template_calendar_cell'>
    //actual calendar_cell.js content
    </script>
    <script>
    var calendar_cell = $('#template_calendar_cell').html();
    </script>


TYPES

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
}

func (this *Context) ContentType(s string)

func (this *Context) DeleteBy(table string, field string, v interface{})

func (this *Context) DeleteId(table string, id interface{})

func (this *Context) Error(message string, code int)
    example:

    ctx.Error("something very very bad just happened",500)
    // or
    ctx.Error("something very very bad just happened",http.StatusInternalServerError)

func (this *Context) FindBy(table string, field string, v interface{}) map[string]interface{}

func (this *Context) FindById(table string, id interface{}) map[string]interface{}

func (this *Context) IsXHR() bool
    returns true/false if the request is XHR example:

    if ctx.IsXHR() {
        ctx.Layout = "special-ajax-lajout"
        // or
        ctx.Render("ajax")
    }

func (this *Context) Log(format string, v ...interface{})

func (this *Context) Query(query string, args ...interface{}) []map[string]interface{}
    WARNING: POC, bad performance, do not use in production.

    Returns slice of map[query_result_fields]query_result_values, so for
    example table with fields id,data,stamp will return [{id: xx,data: xx,
    stamp: xx},{id: xx,data: xx,stamp: xx}] example:

    ctx.O["SessionList"] = ctx.Query("SELECT * FROM session")

    and then in the template:

    {{range .SessionList}}
        id: {{.id}}<br>
        data: {{.data}}<br>
        stamp: {{.stamp}}
    {{end}}

func (this *Context) Redirect(url string)
    example:

    ctx.Redirect("http://golang.org")

func (this *Context) Render(extra ...string)
    renders a template, if the template name starts with os.PathSeparator it
    is rendered with absolute path otherwise it is appended to Views
    WARNING: all template names are converted to lower case

    ctx.Render("show") // -> ./v/show.html (Views + "show" + ".html")
    ctx.Render("/tmp/show") // -> /tmp/show.html ("/tmp/show" + ".html")

    if left without arguments (ctx.Render()) - it takes the
    package_name.function_name and renders

    v/package_name/function.templateExt

    so for example if we have package gallery with function Album() and we
    have ctx.Render() inside it it will render Views + /gallery/ + album +
    TemplateExt (default: ./v/gallery/album.html)

func (this *Context) RenderJSON(j interface{}, error_code int) error

func (this *Context) Replace(table string, input map[string]interface{}) (int64, error)
    POC: bad performance updates database fields based on map's keys - every
    key that begins with _ is skipped

func (this *Context) Sanitize(s string) string

func (this *Context) Write(s string)
    shorthand for writing strings into the http writer example:

    ctx.Write("luke, i am your father")


SUBDIRECTORIES

    example
```


### looks of the blog example:

![post list](http://img690.imageshack.us/img690/576/screenshot20120828at926.png)
![single post](http://img502.imageshack.us/img502/2151/screenshot20120828at927.png)
![admin-panel](http://img845.imageshack.us/img845/2151/screenshot20120828at927.png)
