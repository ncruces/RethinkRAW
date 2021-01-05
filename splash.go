package main

import (
	"os"

	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
)

func printSplash() {
	if !isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		return
	}

	const (
		rebeccapurple = "\033[48;2;102;51;153m"
		black         = "\033[48;5;16m"
		indigo        = "\033[48;5;54m"
		white         = "\033[48;5;231m"
		text          = "\033[38;5;231m"
		reset         = "\033[0m"
	)

	var purple string
	if colors := os.Getenv("COLORTERM"); colors == "truecolor" || colors == "24bit" {
		purple = rebeccapurple
	} else {
		purple = indigo
	}

	out := colorable.NewColorableStdout()

	var icon = []string{
		"................",
		"................",
		"...        .....",
		"... ******  ....",
		"... *******  ...",
		"... **    ** ...",
		"... **    ** ...",
		"... **   **  ...",
		"... ******   ...",
		"... **  ***  ...",
		"... **   *** ...",
		"... **    *  ...",
		"...    .     ...",
		"...    ..   ....",
		"................",
	}

	var old string
	for i := range icon {
		old = ""
		out.Write([]byte(reset))
		out.Write([]byte("\n  "))
		for _, c := range icon[i] {
			var col string
			switch c {
			default:
				col = white
			case ' ':
				col = black
			case '.':
				col = purple
			}
			if col != old {
				out.Write([]byte(col))
				old = col
			}
			out.Write([]byte("  "))
		}
	}
	out.Write([]byte(reset))
	out.Write([]byte("\n  "))
	out.Write([]byte(purple))
	out.Write([]byte(text))
	out.Write([]byte("           RethinkRAW           "))
	out.Write([]byte(reset))
	out.Write([]byte("\n"))
	out.Write([]byte("\n"))
}
