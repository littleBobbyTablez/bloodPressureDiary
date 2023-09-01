package main

import (
	"fmt"
	"html/template"
	"io"
	"os"

	chartrender "github.com/go-echarts/go-echarts/v2/render"
)

type snippetRenderer struct {
	c      interface{}
	before []func()
}

var baseTpl string = fileToString("statics/chart.html")

func fileToString(path string) string {
	bytes, _ := os.ReadFile(path)
	return string(bytes)
}

func newSnippetRenderer(c interface{}, before ...func()) chartrender.Renderer {
	return &snippetRenderer{c: c, before: before}
}

func (r *snippetRenderer) Render(w io.Writer) error {
	const tplName = "chart"
	for _, fn := range r.before {
		fn()
	}

	tpl := template.
		Must(template.New(tplName).
			Funcs(template.FuncMap{
				"safeJS": func(s interface{}) template.JS {
					return template.JS(fmt.Sprint(s))
				},
			}).Parse(baseTpl),
		)

	err := tpl.ExecuteTemplate(w, tplName, r.c)
	return err
}
