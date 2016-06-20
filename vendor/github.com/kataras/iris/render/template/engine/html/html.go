package html

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/kataras/iris/config"
	"github.com/kataras/iris/context"
	"github.com/kataras/iris/utils"
)

var (
	buffer *utils.BufferPool
)

type (
	Engine struct {
		Config    *config.Template
		Templates *template.Template
	}
)

var emptyFuncs = template.FuncMap{
	"yield": func() (string, error) {
		return "", fmt.Errorf("yield was called, yet no layout defined")
	},
	"partial": func() (string, error) {
		return "", fmt.Errorf("block was called, yet no layout defined")
	},
	"current": func() (string, error) {
		return "", nil
	}, "render": func() (string, error) {
		return "", nil
	},
}

// New creates and returns a HTMLTemplate  engine
func New(cfg ...config.Template) *Engine {
	if buffer == nil {
		buffer = utils.NewBufferPool(64)
	}

	c := config.DefaultTemplate().Merge(cfg)

	return &Engine{Config: &c}
}

func (s *Engine) GetConfig() *config.Template {
	return s.Config
}

func (s *Engine) BuildTemplates() error {

	if s.Config.Asset == nil || s.Config.AssetNames == nil {
		return s.buildFromDir()

	}
	return s.buildFromAsset()

}

func (s *Engine) buildFromDir() error {
	if s.Config.Directory == "" {
		return nil //we don't return fill error here(yet)
	}

	var templateErr error
	dir := s.Config.Directory
	s.Templates = template.New(dir)
	s.Templates.Delims(s.Config.HTMLTemplate.Left, s.Config.HTMLTemplate.Right)

	// Walk the supplied directory and compile any files that match our extension list.
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		// Fix same-extension-dirs bug: some dir might be named to: "users.tmpl", "local.html".
		// These dirs should be excluded as they are not valid golang templates, but files under
		// them should be treat as normal.
		// If is a dir, return immediately (dir is not a valid golang template).
		if info == nil || info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		ext := ""
		if strings.Index(rel, ".") != -1 {
			ext = filepath.Ext(rel)
		}

		for _, extension := range s.Config.Extensions {
			if ext == extension {
				buf, err := ioutil.ReadFile(path)
				if err != nil {
					templateErr = err
					break
				}
				name := filepath.ToSlash(rel)
				tmpl := s.Templates.New(name)

				// Add our funcmaps.
				for _, funcs := range s.Config.HTMLTemplate.Funcs {
					tmpl.Funcs(funcs)
				}

				tmpl.Funcs(emptyFuncs).Parse(string(buf))
				break
			}
		}
		return nil
	})

	return templateErr
}

func (s *Engine) buildFromAsset() error {
	var templateErr error
	dir := s.Config.Directory
	s.Templates = template.New(dir)
	s.Templates.Delims(s.Config.HTMLTemplate.Left, s.Config.HTMLTemplate.Right)

	for _, path := range s.Config.AssetNames() {
		if !strings.HasPrefix(path, dir) {
			continue
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			panic(err)
		}

		ext := ""
		if strings.Index(rel, ".") != -1 {
			ext = "." + strings.Join(strings.Split(rel, ".")[1:], ".")
		}

		for _, extension := range s.Config.Extensions {
			if ext == extension {

				buf, err := s.Config.Asset(path)
				if err != nil {
					panic(err)
				}

				name := filepath.ToSlash(rel)
				tmpl := s.Templates.New(name)

				// Add our funcmaps.
				for _, funcs := range s.Config.HTMLTemplate.Funcs {
					tmpl.Funcs(funcs)
				}

				tmpl.Funcs(emptyFuncs).Parse(string(buf))
				break
			}
		}
	}
	return templateErr
}

func (s *Engine) executeTemplateBuf(name string, binding interface{}) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	err := s.Templates.ExecuteTemplate(buf, name, binding)
	return buf, err
}

func (s *Engine) layoutFuncsFor(name string, binding interface{}) {
	funcs := template.FuncMap{
		"yield": func() (template.HTML, error) {
			buf, err := s.executeTemplateBuf(name, binding)
			// Return safe HTML here since we are rendering our own template.
			return template.HTML(buf.String()), err
		},
		"current": func() (string, error) {
			return name, nil
		},
		"partial": func(partialName string) (template.HTML, error) {
			fullPartialName := fmt.Sprintf("%s-%s", partialName, name)
			if s.Config.HTMLTemplate.RequirePartials || s.Templates.Lookup(fullPartialName) != nil {
				buf, err := s.executeTemplateBuf(fullPartialName, binding)
				// Return safe HTML here since we are rendering our own template.
				return template.HTML(buf.String()), err
			}
			return "", nil
		},
		"render": func(fullPartialName string) (template.HTML, error) {
			buf, err := s.executeTemplateBuf(fullPartialName, binding)
			// Return safe HTML here since we are rendering our own template.
			return template.HTML(buf.String()), err

		},
	}
	if tpl := s.Templates.Lookup(name); tpl != nil {
		tpl.Funcs(funcs)
	}
}

func (s *Engine) executeTemplate(out io.Writer, name string, binding interface{}, layout string) error {

	if layout != "" {
		s.layoutFuncsFor(name, binding)
		name = layout
	}

	return s.Templates.ExecuteTemplate(out, name, binding)
}

func (s *Engine) Execute(ctx context.IContext, name string, binding interface{}, layout string) error {
	// Retrieve a buffer from the pool to write to.
	out := buffer.Get()
	if err := s.executeTemplate(out, name, binding, layout); err != nil {
		buffer.Put(out)
		return err
	}
	w := ctx.GetRequestCtx().Response.BodyWriter()
	out.WriteTo(w)

	// Return the buffer to the pool.
	buffer.Put(out)
	return nil
}

func (s *Engine) ExecuteGzip(ctx context.IContext, name string, binding interface{}, layout string) error {
	// Retrieve a buffer from the pool to write to.
	out := gzip.NewWriter(ctx.GetRequestCtx().Response.BodyWriter())

	if err := s.executeTemplate(out, name, binding, layout); err != nil {
		return err
	}
	//out.Flush()
	out.Close()
	ctx.GetRequestCtx().Response.Header.Add("Content-Encoding", "gzip")
	return nil
}
