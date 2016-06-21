package util

import "strconv"
import "github.com/kataras/iris"

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
