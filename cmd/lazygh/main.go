package main

import (
	"fmt"
	"os"

	"codeberg.org/l-lin/lazygh/internal/app"
)

func main() {
	if err := app.New(os.Stdout).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "lazygh: %v\n", err)
		os.Exit(1)
	}
}
