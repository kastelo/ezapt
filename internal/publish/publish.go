package publish

import (
	"fmt"
	"path/filepath"
)

type CLI struct {
	Dists        string   `arg:"" required:"" help:"Path to dists directory" type:"existingdir" env:"EZAPT_DISTS"`
	KeepVersions int      `help:"Number of versions to keep" default:"2" env:"EZAPT_KEEP_VERSIONS"`
	Keyring      string   `required:"" help:"Path to GPG keyring" type:"existingfile" env:"EZAPT_KEYRING"`
	SignUser     []string `help:"GPG user to sign with" env:"EZAPT_SIGN_USER" default:"37C84554E7E0A261E4F76E1ED26E6ED000654A3E,FBA2E162F2F44657B38F0309E5665F9BD5970C47"`
}

func (c *CLI) Run() error {
	keyring, err := filepath.Abs(c.Keyring)
	if err != nil {
		return fmt.Errorf("publish: %w", err)
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

	for _, dist := range dists {
		if err := writeRelease(dist); err != nil {
			return fmt.Errorf("publish: %w", err)
		}
		if err := signRelease(dist, keyring, c.SignUser); err != nil {
			return fmt.Errorf("publish: %w", err)
		}
	}

	return nil
}
