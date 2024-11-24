package publish

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"pault.ag/go/debian/deb"
)

type CLI struct {
	Dists        string `required:"" help:"Path to dists directory" type:"existingdir" env:"EZAPT_DISTS"`
	KeepVersions int    `help:"Number of versions to keep" default:"2" env:"EZAPT_KEEP_VERSIONS"`
	Keyring      string `required:"" help:"Path to GPG keyring" type:"existingfile" env:"EZAPT_KEYRING"`
	Add          string `help:"Path to packages to add" type:"existingdir" env:"EZAPT_ADD"`
}

func (c *CLI) Run() error {
	if c.Add != "" {
		if err := c.add(); err != nil {
			return fmt.Errorf("add: %w", err)
		}
	}

	pkgs, err := scanPackages(c.Dists)
	if err != nil {
		return fmt.Errorf("publish: %w", err)
	}
	trimPackages(c.Dists, pkgs, c.KeepVersions)

	if err := writePackages(c.Dists, pkgs); err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	dists, err := filepath.Glob(filepath.Join(c.Dists, "*"))
	if err != nil {
		return fmt.Errorf("publish: globbing: %w", err)
	}

	fd, err := os.Open(c.Keyring)
	if err != nil {
		return fmt.Errorf("publish: %w", err)
	}
	sign, err := newSigner(fd)
	fd.Close()
	if err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	for _, dist := range dists {
		if err := writeRelease(dist); err != nil {
			return fmt.Errorf("publish: %w", err)
		}
		if err := signRelease(dist, sign); err != nil {
			return fmt.Errorf("publish: %w", err)
		}
	}

	return nil
}

func (c *CLI) add() error {
	return filepath.Walk(c.Add, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".deb" {
			return nil
		}

		deb, cl, err := deb.LoadFile(path)
		if err != nil {
			return err
		}
		defer cl()

		// If we have foo/syncthing/candidate/whatever.deb, grab the
		// syncthing/candidate part
		repoPath, err := filepath.Rel(c.Add, path)
		if err != nil {
			return err
		}
		repoPath = filepath.Dir(repoPath)

		newPath := filepath.Join(c.Dists, repoPath, "binary-"+deb.Control.Architecture.String(), filepath.Base(path))
		slog.Info("Adding package", "package", deb.Control.Package, "version", deb.Control.Version, "architecture", deb.Control.Architecture, "to", newPath)
		if err := os.MkdirAll(filepath.Dir(newPath), 0o700); err != nil {
			return err
		}
		if err := os.Link(path, newPath); errors.Is(err, os.ErrExist) {
			// never mind
		} else if err != nil {
			return err
		}
		return nil
	})
}
