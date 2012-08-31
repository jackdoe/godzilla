package godzilla
import (
	"fmt"
	"strings"
)

// WARNING: POC, bad performance, do not use in production.
// 
// Returns slice of map[query_result_fields]query_result_values,
// so for example table with fields id,data,stamp will return
// [{id: xx,data: xx, stamp: xx},{id: xx,data: xx,stamp: xx}]
// example:
// 		ctx.O["SessionList"] = ctx.Query("SELECT * FROM session")
// and then in the template:
// 	{{range .SessionList}}
//		id: {{.id}}<br>
//		data: {{.data}}<br>
//		stamp: {{.stamp}}
//	{{end}}
func (this *Context) Query(query string, args ...interface{}) []map[string]interface{} {
	var err error
	r := make([]map[string]interface{}, 0)
	rows, err := this.DB.Query(query, args...)
	if err != nil {
		_log("%s - %s", query, err)
		return r
	}
	columns, err := rows.Columns()
	if err != nil {
		_log("%s - %s", query, err)
		return r
	}
	for rows.Next() {
		row := map[string]*interface{}{}
		fields := []interface{}{}
		for _, v := range columns {
			t := new(interface{})
			row[v] = t
			fields = append(fields, t)
		}
		err = rows.Scan(fields...)
		if err != nil {
			_log("%s", err)
		} else {
			x := map[string]interface{}{}
			for k, v := range row {
				x[k] = *v
			}
			r = append(r, x)
		}
	}
	if (Debug & DebugQuery) > 0 {
		_log("extracted %d rows @ %s", len(r), query)
	}
	if (Debug & DebugQueryResult) > 0 {
		_log("%s: %#v", query, r)
	}
	return r
}

// POC: DO NOT USE!
func (this *Context) QueryRecursive(query string, args ...interface{}) []map[string]interface{} {
	r := this.Query(query,args...)
	for _,row := range r {
		for k,v := range row {
			if strings.HasSuffix(k,"_id") {
				rec := "_" + strings.TrimRight(k,"_id")
				row[rec] = this.FindAllBy(rec,"id",v)
			}
		}
	}
	return r
}

// WARNING: POC bad performance
// updates database fields based on map's keys - every key that begins with _ is skipped, returns (last insert id, error)
func (this *Context) Replace(table string, input map[string]interface{}) (int64, error) {
	table = sanitize(table)
	keys := []interface{}{}
	values := []interface{}{}
	skeys := []string{}
	for k, v := range input {
		if len(k) > 0 && k[0] != '_' {
			keys = append(keys, k)
			skeys = append(skeys, "`"+k+"`")
			values = append(values, v)
		}
	}

	questionmarks := strings.TrimRight(strings.Repeat("?,", len(skeys)), ",")
	q := fmt.Sprintf("REPLACE INTO `%s` (%s) VALUES(%s)", table, strings.Join(skeys, ","), questionmarks)
	if (Debug & DebugQuery) > 0 {
		_log("%s", q)
	}
	if (Debug & DebugQueryResult) > 0 {
		_log("%s: %#v", q, input)
	}
	res, e := this.DB.Exec(q, values...)
	if e != nil && (Debug&DebugQuery) > 0 {
		_log("%s: %s", q, e.Error())
	}
	last_id := int64(0)
	if res != nil {
		last_id, _ = res.LastInsertId()
	}
	return last_id, e
}

func (this *Context) FindAllBy(table string, field string, v interface{}) []map[string]interface{} {
	table = sanitize(table)
	field = sanitize(field)
	return this.Query("SELECT * FROM `"+table+"` WHERE `"+field+"`=?", v)
}

func (this *Context) FindBy(table string, field string, v interface{}) map[string]interface{} {
	o := this.FindAllBy(table,field,v)
	if len(o) > 0 {
		return o[0]
	}
	return nil
}
func (this *Context) FindById(table string, id interface{}) map[string]interface{} {
	return this.FindBy(table, "id", id)
}

func (this *Context) DeleteBy(table string, field string, v interface{}) {
	table = sanitize(table)
	field = sanitize(field)
	q := "DELETE FROM `" + table + "` WHERE `" + field + "`=?"
	if (Debug & DebugQuery) > 0 {
		_log("%s", q)
	}
	this.DB.Exec(q, v)
}
func (this *Context) DeleteById(table string, id interface{}) {
	this.DeleteBy(table, "id", id)
}


