package reporter

import (
	"bytes"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"stress-test/internal/runner"
	"strings"
	"text/tabwriter"
	"unicode/utf8"

	"github.com/bytedance/gopkg/util/logger"
	gts "github.com/kopoli/go-terminal-size"
)

type Number interface {
	int | int64 | int32 | float32 | float64
}

func PrintReport(runnerResult *runner.StressResult, params runner.Params) {
	size, getErr := gts.GetSize()
	if getErr != nil {
		size = gts.Size{
			Width: 50,
		}
	}

	codes := slices.Sorted(maps.Keys(runnerResult.StatusCode))

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)

	frameLength := utf8.RuneCountInString("╔")
	var allLines []string

	allLines = append(allLines, fmt.Sprintf(" - Total requests: %d", params.Requests))
	allLines = append(allLines, " - Url: "+params.Url)
	allLines = append(allLines, " - Total time: "+runnerResult.Duration.String())
	allLines = append(allLines, "break")

	fmt.Fprintln(w, "STATUS\tDESCRIÇÃO\tQTD\t%")
	for _, code := range codes {
		percentage := float64(runnerResult.StatusCode[code]) / float64(params.Requests)
		fmt.Fprintln(w, fmt.Sprintf("%d\t%s\t%d\t%.2f", code, http.StatusText(code), runnerResult.StatusCode[code], percentage))
	}
	flushErr := w.Flush()
	if flushErr != nil {
		logger.Info(flushErr)
	}

	for line := range strings.SplitSeq(strings.TrimRight(buf.String(), "\n"), "\n") {
		allLines = append(allLines, line)
	}

	innerWidth := 0
	for _, l := range allLines {
		if w := utf8.RuneCountInString(l); w > innerWidth {
			innerWidth = w
		}
	}
	innerWidth += 4
	innerWidth = min(innerWidth, size.Width-2)

	fmt.Println("╔" + strings.Repeat("=", innerWidth-(2*frameLength)) + "╗")
	printCentered("Summary", innerWidth)
	for _, l := range allLines {
		if l == "break" {
			printLine(strings.Repeat("=", innerWidth-(2*frameLength)), innerWidth)
			continue
		}
		printLine(l, innerWidth)
	}
	fmt.Println("╚" + strings.Repeat("=", innerWidth-(2*frameLength)) + "╝")

}

func printLine(str string, width int) {
	frameLength := utf8.RuneCountInString("║")
	fmt.Println("║" + str + strings.Repeat(" ", width-utf8.RuneCountInString(str)-(frameLength*2)) + "║")
}

func printCentered(str string, width int) {
	frameLength := utf8.RuneCountInString("║")
	fmt.Println("║" + strings.Repeat(" ", (width/2)-utf8.RuneCountInString(str)-frameLength) + str + strings.Repeat(" ", (width/2)-frameLength) + "║")
}
