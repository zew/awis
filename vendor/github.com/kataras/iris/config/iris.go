package config

import "github.com/imdario/mergo"

const (
	DefaultProfilePath = "/debug/pprof"
)

type (
	// Iris configs for the station
	// All fields can be changed before server's listen except the PathCorrection field
	//
	// MaxRequestBodySize is the only options that can be changed after server listen -
	// using Config().MaxRequestBodySize = ...
	// Render's rest config can be changed after declaration but before server's listen -
	// using Config().Render.Rest...
	// Render's Template config can be changed after declaration but before server's listen -
	// using Config().Render.Template...
	// Sessions config can be changed after declaration but before server's listen -
	// using Config().Sessions...
	// and so on...
	Iris struct {
		// MaxRequestBodySize Maximum request body size.
		//
		// The server rejects requests with bodies exceeding this limit.
		//
		// By default request body size is unlimited.
		MaxRequestBodySize int64
		// PathCorrection corrects and redirects the requested path to the registed path
		// for example, if /home/ path is requested but no handler for this Route found,
		// then the Router checks if /home handler exists, if yes,
		// (permant)redirects the client to the correct path /home
		//
		// Default is true
		PathCorrection bool

		// Log turn it to false if you want to disable logger,
		// Iris prints/logs ONLY errors, so be careful when you disable it
		Log bool

		// Profile set to true to enable web pprof (debug profiling)
		// Default is false, enabling makes available these 7 routes:
		// /debug/pprof/cmdline
		// /debug/pprof/profile
		// /debug/pprof/symbol
		// /debug/pprof/goroutine
		// /debug/pprof/heap
		// /debug/pprof/threadcreate
		// /debug/pprof/pprof/block
		Profile bool

		// ProfilePath change it if you want other url path than the default
		// Default is /debug/pprof , which means yourhost.com/debug/pprof
		ProfilePath string

		// Sessions the config for sessions
		// contains 3(three) properties
		// Provider: (look /sessions/providers)
		// Secret: cookie's name (string)
		// Life: cookie life (time.Duration)
		Sessions Sessions

		// Render contains the configs for template and rest configuration
		Render Render
	}

	// Render struct keeps organise all configuration about rendering, templates and rest currently.
	Render struct {
		// Template the configs for template
		Template Template
		// Rest configs for rendering.
		//
		// these options inside this config don't have any relation with the TemplateEngine
		// from github.com/kataras/iris/rest
		Rest Rest
	}
)

// DefaultRender returns default configuration for templates and rest rendering
func DefaultRender() Render {
	return Render{
		// set the default template config both not nil and default Engine to Standar
		Template: DefaultTemplate(),
		// set the default configs for rest
		Rest: DefaultRest(),
	}
}

// Default returns the default configuration for the Iris staton
func Default() Iris {
	return Iris{
		PathCorrection:     true,
		MaxRequestBodySize: -1,
		Log:                true,
		Profile:            false,
		ProfilePath:        DefaultProfilePath,
		Sessions:           DefaultSessions(),
		Render:             DefaultRender(),
	}
}

// Merge merges the default with the given config and returns the result
// receives an array because the func caller is variadic
func (c Iris) Merge(cfg []Iris) (config Iris) {
	// I tried to make it more generic with interfaces for all configs, inside config.go but it fails,
	// so do it foreach configuration np they aint so much...

	if cfg != nil && len(cfg) > 0 {
		config = cfg[0]
		mergo.Merge(&config, c)
	} else {
		_default := c
		config = _default
	}

	return
}

// Merge MergeSingle the default with the given config and returns the result
func (c Iris) MergeSingle(cfg Iris) (config Iris) {

	config = cfg
	mergo.Merge(&config, c)

	return
}
