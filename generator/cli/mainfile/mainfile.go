package mainfile

import (
	_ "embed"

	"gitlab.com/mnm/bud/internal/bail"
	"gitlab.com/mnm/bud/internal/gotemplate"
	"gitlab.com/mnm/bud/internal/imports"
	"gitlab.com/mnm/bud/pkg/gen"
	"gitlab.com/mnm/bud/pkg/gomod"
)

//go:embed mainfile.gotext
var template string

var generator = gotemplate.MustParse("mainfile.gotext", template)

func New(genFS gen.FS, module *gomod.Module) *Generator {
	return &Generator{genFS, module}
}

type Generator struct {
	genFS  gen.FS
	module *gomod.Module
}

func (g *Generator) GenerateFile(f gen.F, file *gen.File) error {
	// Load command state
	state, err := g.Load()
	if err != nil {
		return err
	}
	// Generate our template
	code, err := generator.Generate(state)
	if err != nil {
		return err
	}
	file.Write(code)
	return nil
}

func (g *Generator) Load() (*State, error) {
	loader := &loader{
		Generator: g,
		imports:   imports.New(),
	}
	return loader.Load()
}

type loader struct {
	bail.Struct
	*Generator
	imports *imports.Set
}

func (l *loader) Load() (state *State, err error) {
	defer l.Recover(&err)
	l.imports.AddStd("os")
	l.imports.AddNamed("program", l.module.Import("bud/.cli/program"))
	return &State{
		Imports: l.imports.List(),
	}, nil
}