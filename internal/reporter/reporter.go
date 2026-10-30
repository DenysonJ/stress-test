package reporter

import (
	"bytes"
	"fmt"
	"net/http"
	"sort"
	"stress-test/internal/runner"
	"strings"
	"text/tabwriter"
	"unicode/utf8"

	gts "github.com/kopoli/go-terminal-size"
)

const frameWidth = 2

const breakMarker = "\x00break\x00"

const barWidth = 20

func PrintReport(runnerResult *runner.StressResult, params runner.Params) {
	size, getErr := gts.GetSize()
	if getErr != nil {
		size = gts.Size{Width: 80}
	}

	lines := make([]string, 0)

	// Resumo geral.
	lines = append(lines,
		fmt.Sprintf(" - URL: %s", params.Url),
		fmt.Sprintf(" - Total requests:   %d", params.Requests),
		fmt.Sprintf(" - Concurrency:      %d", params.Concurrency),
		fmt.Sprintf(" - Successful (2xx): %d", runnerResult.GetNumberOfSuccess()),
		fmt.Sprintf(" - Failed (>=400):   %d", runnerResult.GetNumberOfErrors()),
		fmt.Sprintf(" - Network errors:   %d", len(runnerResult.Error)),
		fmt.Sprintf(" - Total time:       %s", runnerResult.Duration),
		breakMarker,
	)

	// Tabela de distribuição de status.
	lines = append(lines, statusTable(runnerResult, params)...)
	lines = append(lines, breakMarker)

	// Tabela de resumo dos bodies.
	lines = append(lines, bodyTable(runnerResult)...)

	innerWidth := boxWidth(lines, size.Width)

	fmt.Println("╔" + strings.Repeat("=", innerWidth-frameWidth) + "╗")
	printCentered("Summary", innerWidth)
	for _, line := range lines {
		if line == breakMarker {
			printLine(strings.Repeat("=", innerWidth-frameWidth), innerWidth)
			continue
		}
		printLine(line, innerWidth)
	}
	fmt.Println("╚" + strings.Repeat("=", innerWidth-frameWidth) + "╝")
}

// statusTable monta as linhas (já alinhadas) da distribuição de códigos HTTP.
func statusTable(runnerResult *runner.StressResult, params runner.Params) []string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)

	codes := sortedCodes(runnerResult.StatusCode)
	maxCount := 0
	for _, count := range runnerResult.StatusCode {
		if count > maxCount {
			maxCount = count
		}
	}

	fmt.Fprintln(w, "STATUS\tDESCRIPTION\tCOUNT\t%\tDISTRIBUTION")
	for _, code := range codes {
		count := runnerResult.StatusCode[code]
		percentage := 0.0
		if params.Requests > 0 {
			percentage = float64(count) / float64(params.Requests) * 100
		}
		fmt.Fprintf(w, "%d\t%s\t%d\t%.2f%%\t%s\n", code, http.StatusText(code), count, percentage, bar(count, maxCount))
	}
	w.Flush()

	return bufferLines(&buf)
}

// bar devolve uma barra de blocos proporcional a count em relação a maxCount.
// Se houver ao menos uma ocorrência, mostra no mínimo um bloco.
func bar(count, maxCount int) string {
	if maxCount <= 0 {
		return ""
	}
	n := count * barWidth / maxCount
	if n == 0 && count > 0 {
		n = 1
	}
	return strings.Repeat("█", n)
}

// bodyTable monta as linhas (já alinhadas) do resumo dos bodies.
func bodyTable(runnerResult *runner.StressResult) []string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, "STATUS\tDISTINCT\tAVG SIZE\tSAMPLE")
	for _, stat := range runnerResult.BodySummary() {
		fmt.Fprintf(w, "%d\t%d\t%dB\t%s\n", stat.Code, stat.DistinctCount, stat.AvgSize, stat.Sample)
	}
	w.Flush()

	return bufferLines(&buf)
}

// sortedCodes devolve as chaves do map ordenadas, para a saída ser estável.
func sortedCodes(statusCode map[int]int) []int {
	codes := make([]int, 0, len(statusCode))
	for code := range statusCode {
		codes = append(codes, code)
	}
	sort.Ints(codes)
	return codes
}

// bufferLines quebra o conteúdo do tabwriter em linhas, sem a quebra final.
func bufferLines(buf *bytes.Buffer) []string {
	trimmed := strings.TrimRight(buf.String(), "\n")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "\n")
}

// boxWidth escolhe a largura interna da moldura: o conteúdo mais largo (com uma
// margem), nunca excedendo a largura do terminal.
func boxWidth(lines []string, terminalWidth int) int {
	widest := utf8.RuneCountInString("Summary")
	for _, line := range lines {
		if w := utf8.RuneCountInString(line); w > widest {
			widest = w
		}
	}

	width := widest + frameWidth + 2 // margem de 1 espaço de cada lado
	if terminalWidth > frameWidth {
		width = min(width, terminalWidth)
	}
	return width
}

// printLine imprime uma linha dentro da moldura, truncando por runes se preciso
func printLine(str string, width int) {
	inner := width - frameWidth
	if utf8.RuneCountInString(str) > inner {
		str = string([]rune(str)[:inner])
	}
	pad := inner - utf8.RuneCountInString(str)
	fmt.Println("║" + str + strings.Repeat(" ", pad) + "║")
}

// printCentered imprime uma linha centralizada dentro da moldura.
func printCentered(str string, width int) {
	inner := width - frameWidth
	if utf8.RuneCountInString(str) > inner {
		str = string([]rune(str)[:inner])
	}
	total := inner - utf8.RuneCountInString(str)
	left := total / 2
	right := total - left
	fmt.Println("║" + strings.Repeat(" ", left) + str + strings.Repeat(" ", right) + "║")
}
