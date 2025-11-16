package analyzer

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestPanicFatalExitChecker(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), PanicFatalExitChecker, "./...")
}
