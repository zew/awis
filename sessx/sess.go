package sessx

import (
	"fmt"
	"strconv"
	"time"

	"github.com/kataras/iris"
	"github.com/kataras/iris/sessions"
	"github.com/zew/awis/logx"

	_ "github.com/kataras/iris/sessions/providers/memory"
	// add a store it is auto-registers itself
)

var sess *sessions.Manager

func init() {
	// when import _ "github.com/kataras/iris/sessions/providers/memory"
	iris.Config().Sessions.Provider = "memory"
	iris.Config().Sessions.Cookie = "irissessionid"
	iris.Config().Sessions.GcDuration = time.Duration(2) * time.Hour

}

var sessDefaultStr = map[string]string{
	"year":        fmt.Sprintf("%d", time.Now().Year()),
	"country":     "DE",
	"param_group": "cit",
}

func BookMarkUrl(c *iris.Context) string {

	ret := string(c.RequestURI())
	ret = string(c.PathString()) + "?&"
	for key, _ := range sessDefaultStr {
		ret = fmt.Sprintf("%v%v=%v&", ret, key, c.Session().GetString(key))
	}
	return ret
}
func InitSess(c *iris.Context) {

	// s := Sess().Start(c)

	for key, vDef := range sessDefaultStr {
		if vReq := EffectiveParam(c, key); vReq != "" {
			c.Session().Set(key, vReq)
			continue
		}
		if vSess := c.Session().GetString(key); vSess == "" {
			c.Session().Set(key, vDef)
		}
	}

	for key, _ := range sessDefaultStr {
		logx.Printf("sess key %14v is %q", key, c.Session().GetString(key))
	}

	c.Next()
}

// EffectiveParamInt is a wrapper around EffectiveParam
// with subsequent parsing into an int
func EffectiveParamInt(c *iris.Context, key string, defaultVal ...int) int {
	s := EffectiveParam(c, key)
	if s == "" {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		} else {
			return 0
		}
	} else {
		i, _ := strconv.Atoi(s)
		return i
	}
}

// EffectiveParamFloat is a wrapper around EffectiveParam
// with subsequent parsing into float
func EffectiveParamFloat(c *iris.Context, key string, defaultVal ...float64) (float64, error) {
	s := EffectiveParam(c, key)
	if s == "" {
		if len(defaultVal) > 0 {
			return defaultVal[0], nil
		} else {
			return 0.0, nil
		}
	} else {
		fl, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0.0, err
		}
		return fl, nil
	}
}

// EffectiveParam searches for the effective value.
// First among the POST fields.
// Then among the URL "path" parameters.
// Then among the URL GET parameters.
// Then inside the session.
// It might be smarter, to condense all levels down to session level
// at the begin of each request.
// We then would only ask the session.
func EffectiveParam(c *iris.Context, key string, defaultVal ...string) string {

	p := ""

	p = c.PostFormValue(key)
	if p != "" {
		return p
	}
	if c.RequestCtx.PostArgs().Has(key) {
		return p
	}

	// Path Param
	p = c.Param(key)
	if p != "" {
		return p
	}

	// URL Get Param
	p = c.URLParam(key)
	if p != "" {
		return p
	}
	urlKeys := c.URLParams()
	if _, ok := urlKeys[key]; ok {
		return p
	}

	// Session
	p = c.Session().GetString(key)
	if p != "" {
		return p
	}
	sessKeys := c.Session().GetAll()
	if _, ok := sessKeys[key]; ok {
		return p
	}

	// default
	def := ""
	if len(defaultVal) > 0 {
		def = defaultVal[0]
	}
	return def

}
