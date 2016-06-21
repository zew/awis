package main

import (
	"html/template"
	"strings"

	"github.com/kataras/iris"
	"github.com/kataras/iris/config"

	appcfg "github.com/zew/awis/config"
	"github.com/zew/awis/gorpx"
	"github.com/zew/awis/logx"
)

var funcMap1 = template.FuncMap{
	"pref":  Pref,
	"title": strings.Title,
	"toJS":  func(arg string) template.JS { return template.JS(arg) },
}

var funcMapAll = []template.FuncMap{
	funcMap1,
}

var funcMapAll2 = map[string]interface{}{
	"fmap1": funcMap1,
}

var irisConfig = config.Iris{}

// The url path prefix
func Pref(p ...string) string {
	s := appcfg.Config.AppName
	s = strings.ToLower(s)
	if len(p) > 0 {
		return "/" + s + p[0]
	}
	return "/" + s
}

// The name of the application
func AppName(p ...string) string {
	s := appcfg.Config.AppName
	if len(p) > 0 {
		return s + p[0]
	}
	return s
}

func main() {

	// iris.Templates("./*.html")

	var renderOptions = config.Template{
		Directory:  "templates",
		Extensions: []string{".tmpl", ".html"},
		// RequirePartials: true,
		HTMLTemplate: config.HTMLTemplate{
			Funcs: funcMapAll,
		},
	}

	irisConfig.Render.Template = renderOptions
	irisConfig.Render.Template.Layout = "layout.html"

	i01 := iris.New(irisConfig)
	// i01 := iris.Custom(iris.StationOptions{})

	i01.Static(Pref("/js"), "./static/js/", 2)
	// i01.Static("/js", "./static/js/", 1)
	i01.Static(Pref("/img"), "./static/img/", 2)
	i01.Static(Pref("/css"), "./static/css/", 2)

	i01.Get("/", index)
	i01.Get(Pref(""), index)
	i01.Get(Pref("/"), index)

	i01.Get(Pref("/queryawis"), queryawis)

	logx.Printf("setting up mysql server...")
	gorpx.DBMap()
	defer gorpx.DB().Close()

	logx.Printf("starting http server...")
	logx.Fatal(i01.ListenWithErr(":8081"))

}
