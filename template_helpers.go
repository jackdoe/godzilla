package godzilla

import (
	"fmt"
	"io/ioutil"
	"path"
	"log"
)

//{{ js "calendar_cell" "calendar_row"}} will read /static/calendar_cell.js and generate
//		<script type='text/template' id='template_calendar_cell'>
//		//actual calendar_cell.js content
//		</script>
//		<script>
//		var calendar_cell = $('#template_calendar_cell').html();
//		</script>
func Template_js(args ...string) string {
	if !EnableStaticDirectory {
		return ""
	}
	s := ""
	for _, v := range args {
		f := path.Join(static_dir, path.Clean(v)) + ".js"
		if (Debug & DebugTemplateRendering) > 0 {
			log.Printf("template_js: %s",f)
		}
		data, e := ioutil.ReadFile(f)
		if e == nil {
			s += fmt.Sprintf("<script type='text/template' id='template_%s'>\n%s\n</script>\n<script>\nvar %s = $('#template_%s').html();\n</script>\n", v, string(data), v, v)
		} 
	}
	return s
}
