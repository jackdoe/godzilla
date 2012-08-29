package godzilla
import (
	"strings"
	"fmt"
)

//{{ js "calendar/cell" "calendar/row"}}
//	<script type='text/template' id='calendar/cell' src='/calendar/cell.js'></script><script>var calendar_cell = $('#calendar/cell').html();</script>
//	<script type='text/template' id='calendar/row' src='/calendar/row.js'></script><script>var calendar_row = $('#calendar/row').html();</script>
func Template_js(args ...string) string {
	s := ""
	for _,v := range args {
		s += fmt.Sprintf("<script type='text/template' id='%s' src='/%s.js'></script><script>var %s = $('#%s').html();</script>",v,v,strings.Replace(v,"/","_",-1),v)
	}
	return s
}

