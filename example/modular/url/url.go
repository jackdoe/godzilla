package url
import (
	"github.com/jackdoe/godzilla"
	"reflect"
	"fmt"
)
func Append(ctx *godzilla.Context) {
	id,_ := ctx.Replace("url",map[string]interface{}{"url":ctx.Splat[1]})
	ctx.Write(fmt.Sprintf("/url/%d",id))
}

func Redirect(ctx *godzilla.Context) {
	u := ctx.FindById("url",ctx.Splat[1])
	if u == nil {
		ctx.Error("not found",404)
	} else {
		ctx.Redirect(reflect.ValueOf(u["url"]).String())
	}
}
