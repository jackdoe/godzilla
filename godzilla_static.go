package godzilla
import (
	"net/http"
	"path"
)
func define_static_route() {
	if (EnableStaticDirectory) {
		_log("enabled static directory: " + StaticDirectory)
		Route("^/" + StaticDirectory + "/",staticRoute)
	}
}
func staticRoute(ctx *Context) {
	rpath := ctx.Re.ReplaceAllString(ctx.R.URL.Path,"") // so we have just the filename left
	f := path.Join(static_dir, path.Clean(rpath))
	if file_exists(f) && (ctx.R.Method == "GET" || ctx.R.Method == "HEAD") {
		ctx.Log("FILE: %s {URI: %s}", f, rpath)
		http.ServeFile(ctx.W, ctx.R, path.Clean(f))
		return
	}
}