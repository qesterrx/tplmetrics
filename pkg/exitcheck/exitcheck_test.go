package exitcheck

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestExitAnalyzer(t *testing.T) {
	// analysistest.TestData() ожидает директорию "testdata"
	analysistest.Run(t, analysistest.TestData(), ExitAnalyzer, "./...")
}


