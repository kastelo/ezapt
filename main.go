package main

import (
	"log/slog"
	"os"

	"github.com/alecthomas/kong"
	"kastelo.dev/ezapt/internal/publish"
)

func main() {
	var cli publish.CLI
	ctx := kong.Parse(&cli)
	if err := ctx.Run(); err != nil {
		slog.Error("Failed to run", "error", err)
		os.Exit(1)
	}
}
