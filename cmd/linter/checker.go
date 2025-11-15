package main

import (
	linter "github.com/noedaka/go-url-shortener/cmd/linter/analyzer"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(linter.PanicFatalExitChecker)
}
