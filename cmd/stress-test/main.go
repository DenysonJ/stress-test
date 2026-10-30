package main

import (
	"net/http"
	"os"
	"time"

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

	httpClient := &http.Client{Timeout: 30 * time.Second}

	runnerService := runner.NewStressRunner(*params, httpClient)
	result := runnerService.Start()

	reporter.PrintReport(result, *params)
}
