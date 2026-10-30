package runner

import (
	"crypto/sha256"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

const MaxBodyPerCode = 5

const sampleMaxRunes = 50

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

type BodyStat struct {
	Code          int
	Count         int
	DistinctCount int
	AvgSize       int
	Sample        string
}

// ErrorStat é a contagem de um tipo de erro de rede (agrupado por mensagem).
type ErrorStat struct {
	Message string
	Count   int
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

func (sr StressResult) BodySummary() []BodyStat {
	codes := make([]int, 0, len(sr.Response))
	for code := range sr.Response {
		codes = append(codes, code)
	}
	sort.Ints(codes)

	stats := make([]BodyStat, 0, len(codes))
	for _, code := range codes {
		bodies := sr.Response[code]
		stat := BodyStat{
			Code:          code,
			Count:         sr.StatusCode[code],
			DistinctCount: len(bodies),
		}

		if len(bodies) > 0 {
			totalSize := 0
			for _, b := range bodies {
				totalSize += len(b)
			}
			stat.AvgSize = totalSize / len(bodies)
			stat.Sample = sampleOf(bodies[0])
		}

		stats = append(stats, stat)
	}

	return stats
}

// ErrorSummary agrupa os erros de rede (Error) por mensagem e devolve a
// contagem de cada um, ordenado do mais frequente para o menos frequente.
func (sr StressResult) ErrorSummary() []ErrorStat {
	counts := make(map[string]int)
	for _, e := range sr.Error {
		counts[e.Error()]++
	}

	stats := make([]ErrorStat, 0, len(counts))
	for msg, n := range counts {
		stats = append(stats, ErrorStat{Message: msg, Count: n})
	}

	sort.Slice(stats, func(i, j int) bool {
		if stats[i].Count != stats[j].Count {
			return stats[i].Count > stats[j].Count
		}
		return stats[i].Message < stats[j].Message
	})

	return stats
}

func sampleOf(body []byte) string {
	flattened := strings.Join(strings.Fields(string(body)), " ")

	runes := []rune(flattened)
	if len(runes) > sampleMaxRunes {
		return string(runes[:sampleMaxRunes]) + "..."
	}

	return flattened
}

func (sr StressResult) ExistBody(body []byte, code int) bool {
	shaBody := sha256.Sum256(body)

	for _, r := range sr.Response[code] {
		rs := sha256.Sum256(r)

		if shaBody == rs {
			return true
		}
	}

	return false
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
			if !result.ExistBody(response.Body, response.Code) && len(result.Response[response.Code]) < MaxBodyPerCode {
				result.Response[response.Code] = append(result.Response[response.Code], response.Body)
			}
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
