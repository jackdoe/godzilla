package main
import ( 
	"database/sql" 
	"github.com/jackdoe/godzilla" 
	_ "github.com/mattn/go-sqlite3" )
func list(ctx *godzilla.Context) {
    ctx.O["posts"] = ctx.Query("SELECT * FROM posts")
    ctx.Render()
}

func main() {
    db, _ := sql.Open("sqlite3", "./foo.db")
    defer db.Close()
    db.Exec("CREATE TABLE IF NOT EXISTS posts (id INTEGER PRIMARY KEY,data TEXT NOT NULL)")
    godzilla.Route("^/$", list)
    godzilla.Start("localhost:8080", db)
}
