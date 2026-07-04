package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
)

func Header(text string) {
	fmt.Println(titleStyle.Render("== " + text))
}

func Dimf(format string, args ...any) {
	fmt.Println(dimStyle.Render(fmt.Sprintf(format, args...)))
}

func Success(text string) {
	fmt.Println(successStyle.Render("ok " + text))
}

func Warn(text string) {
	fmt.Println(warnStyle.Render("! " + text))
}

func Fatalf(format string, args ...any) {
	fmt.Fprintln(os.Stderr, errorStyle.Render("error "+fmt.Sprintf(format, args...)))
	os.Exit(1)
}
