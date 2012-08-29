## just try it out
```
$ go get github.com/jackdoe/godzilla
$ git clone https://github.com/jackdoe/godzilla.git
$ cd godzilla/example/blog && go run main.go
```
***

open http://localhost:8080 and enjoy :)

> http://localhost:8080/admin/ is for the admin panel 

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

### looks of the blog example:

![post list](http://img690.imageshack.us/img690/576/screenshot20120828at926.png)
![single post](http://img502.imageshack.us/img502/2151/screenshot20120828at927.png)
![admin-panel](http://img845.imageshack.us/img845/2151/screenshot20120828at927.png)
