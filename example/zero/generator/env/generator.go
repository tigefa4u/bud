package env

import (
	"github.com/caarlos0/env/v7"
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
	gen.GenerateFile("bud/internal/env/env.go", g.generateFile)
}

const template = `package env

// Code generated by bud; DO NOT EDIT.

{{- if $.Imports }}

import (
	{{- range $import := $.Imports }}
	{{$import.Name}} "{{$import.Path}}"
	{{- end }}
)
{{- end }}

func Load() (*Env, error) {
	var e Env
	// TODO: do this statically instead of using reflect
	if err := runenv.Parse(&e); err != nil {
		return nil, err
	}
	return &e, nil
}

type Env = env.Env
`

var gen = gotemplate.MustParse("env.gotext", template)

type State struct {
	Imports []*imports.Import
}

func (g *Generator) generateFile(fsys generator.FS, file *generator.File) error {
	imset := imports.New()
	imset.AddNamed("env", g.module.Import("env"))
	imset.AddNamed("runenv", g.module.Import("generator/env"))
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

func Parse(e interface{}) error {
	return env.Parse(e)
}
