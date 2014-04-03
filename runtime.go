package main

import (
	"errors"
	"path/filepath"

	"github.com/iron-io/iron_go/worker"
)

// TODO(reed): args for :class?
func runtime(cs worker.CodeSource, typeof, exec string) (string, error) {
	// typeof validated
	rts := map[string]func(worker.CodeSource, string) string{
		"go":     goRuntime,
		"binary": binRuntime,
		"mono":   monoRuntime,
		"ruby":   rbRuntime,
		"python": pyRuntime,
		"perl":   perlRuntime,
		"java":   javaRuntime,
	}
	if rt, ok := rts[typeof]; ok {
		return rt(cs, exec), nil
	}
	return "", errors.New(typeof + " not a valid runtime")
}

func goRuntime(cs worker.CodeSource, exec string) string {
	return `
  go run ` + filepath.Base(exec)
}

func binRuntime(cs worker.CodeSource, exec string) string {
	return `
  chmod +x ` + filepath.Base(exec) + `

  ` + filepath.Base(exec)
}

func monoRuntime(cs worker.CodeSource, exec string) string {
	return `
  mono ` + filepath.Base(exec)
}

// BREAKING: no nice @params from inside ruby script, takes args like everybody else
// TODO(reed): gems
func rbRuntime(cs worker.CodeSource, exec string) string {
	return `
  GEM_PATH="__gems__" 
  GEM_HOME="__gems__"

  ruby  ` + filepath.Base(exec)
}

func pyRuntime(cs worker.CodeSource, exec string) string {
	return ` ` +
		"PATH=`pwd`/__pips__/__bin__:$PATH " +
		"PYTHONPATH=`pwd`/__pips__ " +

		"python -u " + filepath.Base(exec)
}

// assumes .jar with manifest configured
func javaRuntime(cs worker.CodeSource, exec string) string {
	return `
  java -jar ` + filepath.Base(exec)
}

func perlRuntime(cs worker.CodeSource, exec string) string {
	return `
  perl ` + filepath.Base(exec)
}
