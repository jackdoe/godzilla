package godzilla

import (
	"fmt"
	"strings"
)

//{{ js "calendar/cell" "calendar/row"}} will render

//	var calendar_cell; $.get('/calendar/cell.jst',function(data) { calendar_cell = data });
//	var calendar_row; $.get('/calendar/row.jst',function(data) { calendar_row = data });
func Template_js(args ...string) string {
	s := ""
	for _, v := range args {
		us := strings.Replace(v, "/", "_", -1)
		s += fmt.Sprintf("var %s; $.get('/%s.jst',function(data) { %s = data });", us, v, us)
	}
	return s
}
