package add

import (
	"log/slog"
	"os"
	"path/filepath"

	"pault.ag/go/debian/deb"
)

type CLI struct {
	Source string `required:"" help:"Path to packages to add" type:"existingdir"`
	Repo   string `required:"" help:"Path to repository directory" type:"existingdir"`
}

func (c *CLI) Run() error {
	return filepath.Walk(c.Source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".deb" {
			return nil
		}

		slog.Info("Scanning package", "path", path)
		deb, cl, err := deb.LoadFile(path)
		if err != nil {
			return err
		}
		defer cl()

		newPath := filepath.Join(c.Repo, "binary-"+deb.Control.Architecture.String(), filepath.Base(path))
		slog.Info("Adding package", "path", newPath)
		if err := os.MkdirAll(filepath.Dir(newPath), 0o700); err != nil {
			return err
		}
		if err := os.Link(path, newPath); err != nil {
			return err
		}
		return nil
	})
}
