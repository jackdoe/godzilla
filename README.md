## just try it out
```
$ go get github.com/jackdoe/godzilla
$ git clone https://github.com/jackdoe/godzilla.git
$ cd godzilla/example/blog && go run main.go
```

open http://localhost:8080 and enjoy :) (/admin/ is for the admin panel)

***
check out the exampe directory for some apps - there is as simple blog, simple modular app with blog and url shortener, and one small sample app (does nothing usefull)

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
    godzilla.Route("^/url/(\\d+)",url.Redirect)
    godzilla.Route("^/url/append/(.*)",url.Append)
    godzilla.Start("localhost:8080",db)
}

```
if we access _/url/append/http://google.com_ godzilla's router will match _"^/url/append/(.*)"_
and call url.Append(ctx) with one argument of type *Context, this is hot it could look like:
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
* ctx.O - output map, used in templates
* ctx.Layout - layout template (can be empty)
* ctx.Splat - this is where regexp matched results from the route comes - **/url/(\\d+)** - /url/1 will put in Splat []{"/url/1","1"} so ctx.Splat[1] is (\\d+) 
* ctx.Params - simple mapped parameters POST+GET /?a=4 will have `map[string]interface{}{"a":"4"}`
* ctx.Sparams - same as Params but the value is of type string instead interface{} so /?a=4 will have `map[string]string{"a":"4"}`
### looks of the blog example:

![post list](http://img690.imageshack.us/img690/576/screenshot20120828at926.png)
![single post](http://img502.imageshack.us/img502/2151/screenshot20120828at927.png)
![admin-panel](http://img845.imageshack.us/img845/2151/screenshot20120828at927.png)
