package gocat

import (
	"fmt"
	"os"
	"path/filepath"
)

type Module struct {
	Name  string
	Path  string
	Funcs map[string]*Func
}

type Func struct {
	Type     *FuncType
	Name     string
	FuncNode *FuncNode
}

type LoadModuleError struct {
	FilePath   string
	ModulePath string
	Msg        string
}

func (lme *LoadModuleError) Error() string {
	return fmt.Sprintf("Load module error (dir: %q, file: %q): %s",
		lme.ModulePath, lme.FilePath, lme.Msg)
}

func mkFunc(fn *FuncNode) *Func {
	return &Func{
		FuncNode: fn,
		Type:     mkFuncType(fn),
		Name:     fn.Name,
	}
}

func mkFuncType(fn *FuncNode) *FuncType {
	argTypes := make([]Type, len(fn.Args))
	for i, v := range fn.Args {
		argTypes[i] = v.Type
	}
	return &FuncType{
		ArgTypes: argTypes,
		RetTypes: fn.RetTypes,
	}
}

func LoadModule(mpath string) (*Module, error) {
	mname := filepath.Base(mpath)

	// Make sure that mpath is a directory.

	f, err := os.Open(mpath)

	if err != nil {
		return nil, &LoadModuleError{
			ModulePath: mpath,
			FilePath:   "<n/a>",
			Msg:        err.Error(),
		}
	}

	fi, err := f.Stat()

	if err != nil {
		return nil, &LoadModuleError{
			ModulePath: mpath,
			FilePath:   "<n/a>",
			Msg:        err.Error(),
		}
	}

	if !fi.IsDir() {
		return nil, &LoadModuleError{
			ModulePath: mpath,
			FilePath:   "<n/a>",
			Msg:        "Not a directory.",
		}
	}

	funcs := make(map[string]*Func)

	matches, err := filepath.Glob(filepath.Join(mpath, "*.gct"))

	for _, fpath := range matches {
		f, err := os.OpenFile(fpath, os.O_RDONLY, 0)

		if err != nil {
			return nil, &LoadModuleError{
				FilePath:   fpath,
				ModulePath: mpath,
				Msg:        err.Error(),
			}
		}

		p := NewParser(NewTokenizerReader(f, fpath))

		lfuncs, err := p.Funcs()

		if err != nil {
			return nil, &LoadModuleError{
				FilePath:   fpath,
				ModulePath: mpath,
				Msg:        err.Error(),
			}
		}

		for _, lfunc := range lfuncs {
			if funcs[lfunc.Name] != nil {

				return nil, &LoadModuleError{
					ModulePath: mpath,
					FilePath:   fpath,
					Msg:        fmt.Sprintf("Duplicate function `%s`.", lfunc.Name),
				}

			} else {
				funcs[lfunc.Name] = mkFunc(lfunc)
			}
		}
	}

	return &Module{
		Name:  mname,
		Path:  mpath,
		Funcs: funcs,
	}, nil
}
