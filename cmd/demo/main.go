package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/suxin2017/bubbles-diff-view/diffview"
)

func main() {
	left, right, leftTitle, rightTitle, err := loadInputs(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	m := diffview.New(diffview.Options{
		Title:           "bubbles diff view",
		LeftTitle:       leftTitle,
		RightTitle:      rightTitle,
		ShowLineNumbers: true,
		Width:           120,
		Height:          32,
	})
	m.SetDiffStrings(left, right)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "program failed: %v\n", err)
		os.Exit(1)
	}
}

func loadInputs(args []string) (left, right, leftTitle, rightTitle string, err error) {
	if len(args) == 3 {
		leftBytes, readErr := os.ReadFile(args[1])
		if readErr != nil {
			return "", "", "", "", fmt.Errorf("read left file: %w", readErr)
		}
		rightBytes, readErr := os.ReadFile(args[2])
		if readErr != nil {
			return "", "", "", "", fmt.Errorf("read right file: %w", readErr)
		}
		return string(leftBytes), string(rightBytes), args[1], args[2], nil
	}

	left = "" +
		"package main\n" +
		"\n" +
		"import \"fmt\"\n" +
		"\n" +
		"func main() {\n" +
		"    fmt.Println(\"hello\")\n" +
		"}\n"

	right = "" +
		"package main\n" +
		"\n" +
		"import \"fmt\"\n" +
		"\n" +
		"func main() {\n" +
		"    message := \"hello bubbles\"\n" +
		"    fmt.Println(message)\n" +
		"}\n"

	return left, right, "example-left", "example-right", nil
}
