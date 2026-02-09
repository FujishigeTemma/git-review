package output

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

const (
	colorRed    = "\033[0;31m"
	colorGreen  = "\033[0;32m"
	colorYellow = "\033[0;33m"
	colorCyan   = "\033[0;36m"
	colorBold   = "\033[1m"
	colorReset  = "\033[0m"
)

// Output handles formatted terminal output with optional color support.
type Output struct {
	Stdout io.Writer
	Stderr io.Writer
	Color  bool
}

// New creates an Output with TTY-based color detection.
func New() *Output {
	return &Output{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Color:  term.IsTerminal(int(os.Stdout.Fd())),
	}
}

func (o *Output) colorize(color, msg string) string {
	if !o.Color {
		return msg
	}
	return color + msg + colorReset
}

func (o *Output) Info(msg string)  { fmt.Fprintln(o.Stdout, o.colorize(colorCyan, msg)) }
func (o *Output) Warn(msg string)  { fmt.Fprintln(o.Stderr, o.colorize(colorYellow, "Warning: "+msg)) }
func (o *Output) Ok(msg string)    { fmt.Fprintln(o.Stdout, o.colorize(colorGreen, msg)) }
func (o *Output) Err(msg string)   { fmt.Fprintln(o.Stderr, o.colorize(colorRed, "Error: "+msg)) }
func (o *Output) Bold(msg string) string   { return o.colorize(colorBold, msg) }
func (o *Output) Green(msg string) string  { return o.colorize(colorGreen, msg) }
func (o *Output) Yellow(msg string) string { return o.colorize(colorYellow, msg) }

func (o *Output) Printf(format string, args ...any) {
	fmt.Fprintf(o.Stdout, format, args...)
}
