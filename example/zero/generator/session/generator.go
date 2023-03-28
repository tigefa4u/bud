package env

import (
	"github.com/livebud/bud/package/gomod"
	"github.com/livebud/bud/package/gotemplate"
	"github.com/livebud/bud/package/imports"
	"github.com/livebud/bud/runtime/generator"
)

func New(module *gomod.Module) *Generator {
	return &Generator{module}
}

type Generator struct {
	module *gomod.Module
}

func (g *Generator) Extend(gen generator.FileSystem) {
	gen.GenerateFile("bud/pkg/sessions/sessions.go", g.generateFile)
}

const template = `package sessions

// Code generated by bud; DO NOT EDIT.

{{- if $.Imports }}

import (
	{{- range $import := $.Imports }}
	{{$import.Name}} "{{$import.Path}}"
	{{- end }}
)
{{- end }}


type Store struct {
}

func (s *Store) Load(r *http.Request, id string) (*session.Session, error) {
	fmt.Println("loading session", id)
	cookie, err := r.Cookie(id)
	if err != nil {
		return nil, err
	}
	fmt.Println("got cookie", cookie.Name, cookie.Value)
	return &Session{
		ID: "123" + r.URL.Path,
	}, nil
}

func (s *Store) Save(w http.ResponseWriter, id string, sess *session.Session) error {
	fmt.Println("saving session", sess)
	return nil
}

func From(ctx context.Context) (*Session, error) {
	return nil, fmt.Errorf("unable to load session from context")
	// sess, ok := ctx.Value(session.ContextKey).(*Session)
	// if !ok {
	// 	return nil, fmt.Errorf("no session in context")
	// }
	// return sess, nil
}

// func Load(r *http.Request, s *session.Store) (*Session, error) {
// 	fmt.Println("loaded store", s)
// 	return &Session{
// 		ID: "123" + r.URL.Path,
// 	}, nil
// }

type Session = session.Session
`

var gen = gotemplate.MustParse("session.gotext", template)

type State struct {
	Imports []*imports.Import
}

func (g *Generator) generateFile(fsys generator.FS, file *generator.File) error {
	imset := imports.New()
	imset.AddStd("net/http", "context")
	imset.AddStd("fmt")
	imset.AddNamed("session", g.module.Import("session"))
	code, err := gen.Generate(&State{
		Imports: imset.List(),
	})
	if err != nil {
		return err
	}
	file.Data = code
	return nil
}

//////////////////////////////////
// RUNTIME
//////////////////////////////////

// func Parse(e interface{}) error {
// 	return env.Parse(e)
// }