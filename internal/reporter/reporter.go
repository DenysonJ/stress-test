package reporter

import (
	"fmt"
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

	fmt.Println("╔" + strings.Repeat("=", size.Width-2) + "╗")
	printCentered("Summary", size)
	printLineNumber(" - Total requests: ", params.Requests, size)
	printLine(" - Url: "+params.Url, size)
	printLineNumber(" - Error: ", len(runnerResult.Error), size)
	printLineNumber(" - Request Success: ", runnerResult.GetNumberOfSuccess(), size)
	printLineNumber(" - Request Error: ", runnerResult.GetNumberOfErrors(), size)
	printLineNumber(" - Total Duration: ", runnerResult.Duration.Seconds(), size)
	fmt.Println("╚" + strings.Repeat("=", size.Width-2) + "╝")

}

func printLine(str string, size gts.Size) {
	fmt.Println("║" + str + strings.Repeat(" ", size.Width-len(str)-2) + "║")
}

func printLineNumber[V Number](str string, number V, size gts.Size) {
	switch reflect.TypeOf(number).Name() {
	case "int":
		printLine(str+fmt.Sprintf("%d", number), size)
	case "int32":
		printLine(str+fmt.Sprintf("%d", number), size)
	case "int64":
		printLine(str+fmt.Sprintf("%d", number), size)
	case "float32":
		printLine(str+fmt.Sprintf("%.2f", number), size)
	case "float64":
		printLine(str+fmt.Sprintf("%.2f", number), size)
	}
}

func printCentered(str string, size gts.Size) {
	fmt.Println("║" + strings.Repeat(" ", (size.Width/2)-len(str)-1) + str + strings.Repeat(" ", (size.Width/2)-1) + "║")
}
