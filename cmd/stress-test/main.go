package main

import (
	"net/http"
	"os"
	"stress-test/internal/reporter"
	"stress-test/internal/runner"

	"github.com/bytedance/gopkg/util/logger"
)

func main() {
	params, parserErr := runner.Parser()
	if parserErr != nil {
		logger.Error(parserErr)
		os.Exit(1)
	}

	runnerService := runner.NewStressRunner(*params, http.DefaultClient)
	result := runnerService.Start()

	reporter.PrintReport(result, *params)
}
