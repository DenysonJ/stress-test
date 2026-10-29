package runner

import (
	"errors"
	"flag"
	"net/url"
)

type Params struct {
	Url         string
	Requests    int
	Concurrency int
}

func Parser() (*Params, error) {
	urlParam := flag.String("url", "", "URL")
	requests := flag.Int("requests", 0, "Number of requests")
	concurrency := flag.Int("concurrency", 0, "Number of concurrent requests")

	flag.Parse()

	if *urlParam == "" {
		return nil, errors.New("url flag is required")
	}
	if *requests < 1 {
		return nil, errors.New("number of requests flag is required")
	}
	if *concurrency < 1 {
		return nil, errors.New("concurrency flag is required")
	}

	if *requests < *concurrency {
		return nil, errors.New("number of requests must be greater than concurrency")
	}
	parsedUrl, parserErr := url.ParseRequestURI(*urlParam)
	if parserErr != nil {
		return nil, parserErr
	}

	return &Params{parsedUrl.String(), *requests, *concurrency}, nil
}
