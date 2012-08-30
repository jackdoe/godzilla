package godzilla

import (
	"fmt"
	"io/ioutil"
	"path"
)

//{{ js "calendar_cell" "calendar_row"}} will read and render
// reads /static/calendar_cell.js
func Template_js(args ...string) string {
	if !EnableStaticDirectory {
		return ""
	}
	s := ""
	for _, v := range args {
		data, e := ioutil.ReadFile(path.Join(static_dir, path.Clean(v)) + ".js")
		if e != nil {
			s += fmt.Sprintf("<script type='text/template' id='template_%s'>%s</script><script>var %s = $('#template_%s').html();</script>", v, string(data), v, v)
		}
	}
	return s
}
