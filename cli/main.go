package main

import (
	"fmt"
	"os"

	"github.com/guionardo/multi-git/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}