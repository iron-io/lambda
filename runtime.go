package main

import "path/filepath"

// TODO(reed): args for :class?
func runtime(typeof, exec string) string {
	// typeof validated
	return map[string]func(string) string{
		"go":          goRuntime,
		"binary":      binRuntime,
		"monoRuntime": monoRuntime,
	}[typeof](exec)
}

func goRuntime(exec string) string {
	return `
  go run ` + filepath.Base(exec)
}

func binRuntime(exec string) string {
	return `
  chmod +x ` + filepath.Base(exec) + `

  ` + filepath.Base(exec)
}

func monoRuntime(exec string) string {
	return `
  mono ` + filepath.Base(exec)
}
