package linter

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var PanicFatalExitChecker = &analysis.Analyzer{
	Name: "panicfatalexitchecker",
	Doc:  "checks for panic, log.Fatal and os.Exit calls",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			if callExpr, ok := n.(*ast.CallExpr); ok {
				if ident, ok := callExpr.Fun.(*ast.Ident); ok {
					if ident.Name == "panic" {
						pass.Reportf(ident.Pos(), "panic call")
					}
				}
				
				if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
					if ident, ok := selExpr.X.(*ast.Ident); ok {
						if pass.Pkg.Name() != "main" {
							if ident.Name == "log" && selExpr.Sel.Name == "Fatal" {
								pass.Reportf(ident.Pos(), "log.Fatal call")
							} else if ident.Name == "os" && selExpr.Sel.Name == "Exit" {
								pass.Reportf(ident.Pos(), "os.Exit call")
							}
						}
					}
				}
			}

			return true
		})
	}

	return nil, nil
}
