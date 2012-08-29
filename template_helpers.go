package godzilla
import (
	"strings"
	"fmt"
)

//{{ js "calendar/cell" "calendar/row"}} will render
//		<script type='text/template' id='calendar/cell' src='/calendar/cell.js'></script><script>var calendar_cell = $('#calendar/cell').html();</script>
//		<script type='text/template' id='calendar/row' src='/calendar/row.js'></script><script>var calendar_row = $('#calendar/row').html();</script>
func Template_js(args ...string) string {
	s := ""
	for _,v := range args {
		us := strings.Replace(v,"/","_",-1)
		// s += fmt.Sprintf("<script type='text/template' id='%s' src='/%s.js'></script><script>var %s = $('#%s').html();</script>",us,v,us,us)
		s += fmt.Sprintf("<script>var %s; $.get('/%s.js',function(_data) { %s = _data; });</script>",us,v,us)
	}
	return s
}

