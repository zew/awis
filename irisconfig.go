package main

import (
	"html/template"
	"strconv"
	"strings"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/sessions"

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

func irisBaseConfig(i01 *iris.Application) {
	i01.Use(iris.Gzip)

	iris.WithCharset("UTF-8")
	iris.WithoutServerError(iris.ErrServerClosed)
}

func irisInctanceConfig(i01 *iris.Application) {
	engine := iris.HTML("./templates", ".html").Layout("layout.html")
	for k, v := range funcMapAllClassic {
		engine.AddFunc(k, v)
	}
	i01.RegisterView(engine)
	logx.Printf("engine loaded dir; %T", engine)
	logx.Printf("engine funcs %v", funcMapAllClassic)
}

func irisSessionsConfig(i01 *iris.Application) {
	var keysToPersist = map[string]string{
		"country": "DE",
	}

	/*
		var sessDefaultStr = map[string]string{
			"year":        fmt.Sprintf("%d", time.Now().Year()),
			"country":     "DE",
			"param_group": "cit",
		}
	*/

	sessManager := sessions.New(sessions.Config{
		Cookie:       "irissessionid",
		AllowReclaim: true,
		Expires:      2 * time.Hour,
	})
	i01.Use(sessManager.Handler())
	i01.Use(func(c iris.Context) {
		sess := sessions.Get(c)

		for key, vDef := range keysToPersist {
			if vReq := EffectiveParam(c, key); vReq != "" {
				sess.Set(key, vReq)
				continue
			}
			if vSess := sess.GetString(key); vSess == "" {
				sess.Set(key, vDef)
			}
		}

		for key := range keysToPersist {
			logx.Printf("sess key %14v is %q", key, sess.GetString(key))
		}
	})
}

// EffectiveParam searches for the effective value.
// First among the POST fields.
// Then among the URL "path" parameters.
// Then among the URL GET parameters.
// Then inside the session.
// It might be smarter, to condense all levels down to session level
// at the begin of each request.
// We then would only ask the session and flash messages.
func EffectiveParam(ctx iris.Context, key string, defaultVal ...string) string {
	// Form data and url query parameters for POST or PUT HTTP methods.
	if v := ctx.FormValue(key); v != "" {
		return v
	}

	// Path Param.
	if v := ctx.Params().Get(key); v != "" {
		return v
	}

	// URL Get Param.
	if v := ctx.URLParam(key); v != "" {
		return v
	}

	// Session.
	sess := sessions.Get(ctx)
	if sess != nil {
		if v := sess.GetString(key); v != "" {
			return v
		}

		if v := sess.GetFlashString(key); v != "" {
			return v
		}
	}

	def := ""
	if len(defaultVal) > 0 {
		def = defaultVal[0]
	}

	return def
}

// EffectiveParamInt is a wrapper around EffectiveParam
// with subsequent parsing into an int
func EffectiveParamInt(c iris.Context, key string, defaultVal ...int) int {
	s := EffectiveParam(c, key)
	if s == "" {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0

	}
	i, _ := strconv.Atoi(s)
	return i
}

// EffectiveParamFloat is a wrapper around EffectiveParam
// with subsequent parsing into float
func EffectiveParamFloat(c iris.Context, key string, defaultVal ...float64) (float64, error) {
	s := EffectiveParam(c, key)
	if s == "" {
		if len(defaultVal) > 0 {
			return defaultVal[0], nil
		}
		return 0.0, nil

	}

	fl, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0, err
	}
	return fl, nil

}
