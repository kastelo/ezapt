package main

import (
	"log/slog"

	"github.com/alecthomas/kong"
	"kastelo.dev/ezapt/internal/add"
	"kastelo.dev/ezapt/internal/publish"
)

type CLI struct {
	Publish publish.CLI `cmd:"" help:"Publish packages to a repository."`
	Add     add.CLI     `cmd:"" help:"Add packages to a repository."`
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli)
	if err := ctx.Run(); err != nil {
		slog.Error("Failed to run", "error", err)
	}
}
