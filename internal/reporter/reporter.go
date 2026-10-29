package reporter

import (
	"fmt"
	"net/http"
	"reflect"
	"stress-test/internal/runner"
	"strings"

	gts "github.com/kopoli/go-terminal-size"
)

type Number interface {
	int | int64 | int32 | float32 | float64
}

func PrintReport(runnerResult *runner.StressResult, params runner.Params) {
	size, err := gts.GetSize()
	if err != nil {
		size = gts.Size{
			Width: 50,
		}
	}

	successRequests := len(runnerResult.Response[http.StatusOK]) + len(runnerResult.Response[http.StatusCreated]) + len(runnerResult.Response[http.StatusAccepted])

	fmt.Println("╔" + strings.Repeat("=", size.Width-2) + "╗")
	printCentered("Summary", size)
	printLine(" - Total requests: "+string(rune(params.Requests)), size)
	printLine(" - Url: "+params.Url, size)
	printLine(" - Success: "+string(rune(successRequests)), size)
	fmt.Println("╚" + strings.Repeat("=", size.Width-2) + "╝")

}

func printLine(str string, size gts.Size) {
	fmt.Println("║" + str + strings.Repeat(" ", size.Width-len(str)-2) + "║")
}

func printLineNumber[V Number](number V, size gts.Size) {
	switch reflect.TypeOf(number).Name() {
	case "int":

	}
}

func printCentered(str string, size gts.Size) {
	fmt.Println("║" + strings.Repeat(" ", (size.Width/2)-len(str)-1) + str + strings.Repeat(" ", (size.Width/2)-1) + "║")
}
