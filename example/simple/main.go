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
    // db.Exec("CREATE TABLE IF NOT EXISTS posts (id INTEGER PRIMARY KEY,data TEXT NOT NULL)")
    // db.Exec("INSERT INTO posts(data) VALUES('godzilla was here')")
    // db.Exec("INSERT INTO posts(data) VALUES('godzilla left')")
    godzilla.Route("^/$", list)
    godzilla.Start("localhost:8080", db)
}
