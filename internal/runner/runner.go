package runner

import (
	"io"
	"net/http"
	"time"
)

type responseResult struct {
	Code int
	Body []byte
}
type responseResults map[int][][]byte

type StressRunner struct {
	httpClient     *http.Client
	url            string
	numberRequests int
	workers        chan bool
	responses      chan responseResult
	requestsErr    chan error
}

type StressResult struct {
	StatusCode map[int]int
	Response   responseResults
	Error      []error
	Duration   time.Duration
}

func (sr StressResult) GetNumberOfErrors() int {
	result := 0
	for code, n := range sr.StatusCode {
		if code >= 400 {
			result += n
		}
	}
	return result
}

func (sr StressResult) GetNumberOfSuccess() int {
	result := 0
	for code, n := range sr.StatusCode {
		if code >= 200 && code < 300 {
			result += n
		}
	}
	return result
}

func NewStressRunner(params Params, httpClient *http.Client) *StressRunner {
	return &StressRunner{
		httpClient:     httpClient,
		url:            params.Url,
		numberRequests: params.Requests,
		workers:        make(chan bool, params.Concurrency),
		responses:      make(chan responseResult, params.Concurrency),
		requestsErr:    make(chan error, params.Concurrency),
	}
}

func (r *StressRunner) Start() *StressResult {
	now := time.Now()
	go func() {
		for range r.numberRequests {
			r.workers <- true
			go r.worker()
		}
	}()

	result := &StressResult{
		StatusCode: map[int]int{},
		Response:   responseResults{},
		Error:      []error{},
	}

	for range r.numberRequests {
		select {
		case err := <-r.requestsErr:
			result.Error = append(result.Error, err)
		case response := <-r.responses:
			result.Response[response.Code] = append(result.Response[response.Code], response.Body)
			result.StatusCode[response.Code] += 1
		}
	}

	result.Duration = time.Now().Sub(now)

	return result
}

func (r *StressRunner) worker() {
	defer func() { <-r.workers }()

	response, err := r.httpClient.Get(r.url)
	if err != nil {
		r.requestsErr <- err
		return
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		r.requestsErr <- err
		return
	}
	result := responseResult{
		Code: response.StatusCode,
		Body: responseBody,
	}
	r.responses <- result
}
