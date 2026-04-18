package main

import (
	"fmt"
	"os"

	"codeberg.org/l-lin/lazygh/internal/app"
	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"codeberg.org/l-lin/lazygh/internal/tui"
)

func main() {
	if err := app.New(tui.NewProgram(githubcli.NewClient())).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "lazygh: %v\n", err)
		os.Exit(1)
	}
}
