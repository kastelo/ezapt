package main

import (
	"log/slog"
	"os"

	"github.com/alecthomas/kong"
	"kastelo.dev/ezapt/internal/publish"
	"kastelo.dev/ezapt/internal/sign"
)

type CLI struct {
	Publish publish.CLI `cmd:"" help:"Publish a repository." default:""`
	Sign    sign.CLI    `cmd:"" help:"Sign files."`
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli)
	if err := ctx.Run(); err != nil {
		slog.Error("Failed to run", "error", err)
		os.Exit(1)
	}
}
