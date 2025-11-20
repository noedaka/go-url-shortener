package analyzer

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
			callExpr, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			if ident, ok := callExpr.Fun.(*ast.Ident); ok && ident.Name == "panic" {
				pass.Reportf(ident.Pos(), "panic call")
				return true
			}

			selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
			if !ok || pass.Pkg.Name() == "main" {
				return true
			}

			pkgIdent, ok := selExpr.X.(*ast.Ident)
			if !ok {
				return true
			}

			switch {
			case pkgIdent.Name == "log" && selExpr.Sel.Name == "Fatal":
				pass.Reportf(selExpr.Pos(), "log.Fatal call")
			case pkgIdent.Name == "os" && selExpr.Sel.Name == "Exit":
				pass.Reportf(selExpr.Pos(), "os.Exit call")
			}

			return true
		})
	}

	return nil, nil
}
