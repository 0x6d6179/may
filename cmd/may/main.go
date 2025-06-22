package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/0x6d6179/may/internal/cmd/root"
	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/ui"
)

func main() {
	f := factory.New()
	cmd := root.NewCmdRoot(f)
	if err := cmd.Execute(); err != nil {
		if errors.Is(err, ui.ErrAborted) {
			fmt.Fprintln(f.IO.ErrOut, "→  command aborted")
			os.Exit(0)
		}
		fmt.Fprintln(f.IO.ErrOut, err)
		os.Exit(1)
	}
}
