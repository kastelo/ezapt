package publish

import (
	"fmt"
	"path/filepath"
)

type CLI struct {
	Dists        string `arg:"" required:"" help:"Path to dists directory" type:"existingdir"`
	Keyring      string `required:"" help:"Path to GPG keyring" type:"existingfile"`
	KeepVersions int    `help:"Number of versions to keep" default:"2"`
}

func (c *CLI) Run() error {
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

	for _, dist := range dists {
		if err := writeRelease(dist); err != nil {
			return fmt.Errorf("publish: %w", err)
		}
		if err := signRelease(dist); err != nil {
			return fmt.Errorf("publish: %w", err)
		}
	}

	return nil
}
