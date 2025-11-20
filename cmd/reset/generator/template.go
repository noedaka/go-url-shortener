package generator

import (
	"bytes"
	"text/template"
)

type templateImpl struct {
	tmpl *template.Template
}

func newTemplate() (*templateImpl, error) {
	tmpl, err := template.New("func").Parse(funcTemplStr)
	if err != nil {
		return nil, err
	}
	return &templateImpl{tmpl: tmpl}, nil
}

func (t *templateImpl) execute(buf *bytes.Buffer, data TemplateEnum) error {
	return t.tmpl.Execute(buf, data)
}

const funcTemplStr = `
func (rs *{{.StructType}}) Reset() {
	if rs == nil {
		return
	}
	{{range .Entries}}
	{{if .IsPrimitive}}
	{{if .IsPointer}}
	if rs.{{.Name}} != nil {
		*rs.{{.Name}} = {{.ZeroValue}}
	}
	{{else}}
	rs.{{.Name}} = {{.ZeroValue}}
	{{end}}
	{{else if .IsSlice}}
	rs.{{.Name}} = rs.{{.Name}}[:0]
	{{else if .IsMap}}
	clear(rs.{{.Name}})
	{{else if .IsStruct}}
	{{if .IsPointer}}
	if rs.{{.Name}} != nil {
		if resetter, ok := interface{}(rs.{{.Name}}).(interface{ Reset() }); ok {
			resetter.Reset()
		}
	}
	{{else}}
	if resetter, ok := interface{}(&rs.{{.Name}}).(interface{ Reset() }); ok {
		resetter.Reset()
	}
	{{end}}
	{{end}}
	{{end}}
}
`
