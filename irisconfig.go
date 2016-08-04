package main

import (
	"html/template"
	"strings"
	"time"

	"github.com/iris-contrib/template/html"
	"github.com/kataras/iris"
	"github.com/kataras/iris/config"

	appcfg "github.com/zew/awis/config"
	"github.com/zew/logx"
)

var funcMapAllClassic = template.FuncMap{
	"pref":  Pref,
	"title": strings.Title,

	"toJS":    func(arg string) template.JS { return template.JS(arg) },       // JavaScript expression
	"toJSStr": func(arg string) template.JSStr { return template.JSStr(arg) }, // JavaScript string - *automatic quotation*
	"toURL":   func(arg string) template.URL { return template.URL(arg) },
}

// The url path prefix
func Pref(p ...string) string {
	s := appcfg.Config.AppName
	s = strings.ToLower(s)
	s = strings.Replace(s, " ", "_", -1)
	if len(p) > 0 {
		return "/" + s + p[0]
	}
	return "/" + s
}

func irisBaseConfig() config.Iris {

	var irisConf = config.Iris{}

	// irisConf.IsDevelopment = false

	iris.Config.Sessions.Cookie = "irissessionid"
	iris.Config.Sessions.GcDuration = time.Duration(2) * time.Hour

	// iris.Config.Sessions.Provider = "memory"
	// iris.UseSessionDB(db)

	iris.Config.Gzip = true       // compressed gzip contents to the client, the same for Response Engines also, defaults to false
	iris.Config.Charset = "UTF-8" // defaults to "UTF-8", the same for Response Engines also

	return irisConf
}

func irisInctanceConfig(i01 *iris.Framework) {

	htmlConf := html.Config{
		Layout: "layout.html",
		Funcs:  funcMapAllClassic,
	}

	var engine *html.Engine
	engine = html.New(htmlConf)
	engine.LoadDirectory("./templates", ".html")
	logx.Printf("engine loaded dir; %T", engine)
	logx.Printf("engine funcs %v", engine.Funcs())

	var tel *iris.TemplateEngineLocation
	tel = i01.UseTemplate(engine)
	tel.Directory("./templates", ".html")
	logx.Printf("loaded dir2 %T", tel)

}
