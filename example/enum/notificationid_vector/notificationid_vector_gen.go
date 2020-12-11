// Code generated by "genenum.exe -typesize=uint16 -typename=NotificationID -packagename=notificationid -basedir=enum -vectortype=int"

package notificationid_vector

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"

	"github.com/kasworld/mininet/example/enum/notificationid"
)

type NotificationIDVector [notificationid.NotificationID_Count]int

func (es NotificationIDVector) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "NotificationIDVector[")
	for i, v := range es {
		fmt.Fprintf(&buf,
			"%v:%v ",
			notificationid.NotificationID(i), v)
	}
	buf.WriteString("]")
	return buf.String()
}
func (es *NotificationIDVector) Dec(e notificationid.NotificationID) {
	es[e] -= 1
}
func (es *NotificationIDVector) Inc(e notificationid.NotificationID) {
	es[e] += 1
}
func (es *NotificationIDVector) Add(e notificationid.NotificationID, v int) {
	es[e] += v
}
func (es *NotificationIDVector) SetIfGt(e notificationid.NotificationID, v int) {
	if es[e] < v {
		es[e] = v
	}
}
func (es NotificationIDVector) Get(e notificationid.NotificationID) int {
	return es[e]
}

// Iter return true if iter stop, return false if iter all
// fn return true to stop iter
func (es NotificationIDVector) Iter(fn func(i notificationid.NotificationID, v int) bool) bool {
	for i, v := range es {
		if fn(notificationid.NotificationID(i), v) {
			return true
		}
	}
	return false
}

// VectorAdd add element to element
func (es NotificationIDVector) VectorAdd(arg NotificationIDVector) NotificationIDVector {
	var rtn NotificationIDVector
	for i, v := range es {
		rtn[i] = v + arg[i]
	}
	return rtn
}

// VectorSub sub element to element
func (es NotificationIDVector) VectorSub(arg NotificationIDVector) NotificationIDVector {
	var rtn NotificationIDVector
	for i, v := range es {
		rtn[i] = v - arg[i]
	}
	return rtn
}

func (es *NotificationIDVector) ToWeb(w http.ResponseWriter, r *http.Request) error {
	tplIndex, err := template.New("index").Funcs(IndexFn).Parse(`
		<html>
		<head>
		<title>NotificationID statistics</title>
		</head>
		<body>
		<table border=1 style="border-collapse:collapse;">` +
		HTML_tableheader +
		`{{range $i, $v := .}}` +
		HTML_row +
		`{{end}}` +
		HTML_tableheader +
		`</table>
	
		<br/>
		</body>
		</html>
		`)
	if err != nil {
		return err
	}
	if err := tplIndex.Execute(w, es); err != nil {
		return err
	}
	return nil
}

func Index(i int) string {
	return notificationid.NotificationID(i).String()
}

var IndexFn = template.FuncMap{
	"NotificationIDIndex": Index,
}

const (
	HTML_tableheader = `<tr>
		<th>Name</th>
		<th>Value</th>
		</tr>`
	HTML_row = `<tr>
		<td>{{NotificationIDIndex $i}}</td>
		<td>{{$v}}</td>
		</tr>
		`
)