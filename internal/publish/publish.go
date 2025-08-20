package publish

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"slices"

	yaml "go.yaml.in/yaml/v4"
	"kastelo.dev/ezapt/internal/pgp"
)

type CLI struct {
	Root string `required:"" help:"Path to root directory" type:"existingdir" env:"EZAPT_ROOT"`
	pgp.CLI
}

func (c *CLI) Run() error {
	cfgBs, err := os.ReadFile(filepath.Join(c.Root, "ezapt.yaml"))
	if err != nil {
		return err
	}
	var cfg Config
	if err := yaml.Unmarshal(cfgBs, &cfg); err != nil {
		return err
	}

	sign, err := c.Signer()
	if err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	pkgs, err := scanPackages(c.Root, filepath.Join(c.Root, cfg.PoolDir))
	if err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	for _, dist := range cfg.Distributions {
		if err := c.processDistribution(cfg, dist, pkgs, sign); err != nil {
			return err
		}
	}

	for _, pkg := range pkgs {
		if pkg.kept == 0 {
			slog.Info("Deleting unkept package", "file", pkg.Filename)
			os.Remove(filepath.Join(c.Root, pkg.Filename))
		}
	}

	return nil
}

func (c *CLI) processDistribution(cfg Config, dist ConfigDistribution, pkgs []*packageFile, sign *pgp.Signer) error {
	distDir := filepath.Join(c.Root, cfg.DistsDir, dist.Name)
	slog.Info("Processing distribution", "name", dist.Name, "path", distDir)

	if err := os.MkdirAll(distDir, 0o755); err != nil {
		return err
	}

	for _, comp := range dist.Components {
		if err := c.processComponent(comp, pkgs, distDir); err != nil {
			return err
		}
	}

	if err := writeRelease(distDir); err != nil {
		return fmt.Errorf("publish: %w", err)
	}
	if err := signRelease(distDir, sign); err != nil {
		return fmt.Errorf("publish: %w", err)
	}
	return nil
}

func (*CLI) processComponent(comp ConfigComponent, pkgs []*packageFile, distDir string) error {
	filteredPkgs, err := filterPackages(comp, pkgs)
	if err != nil {
		return err
	}

	compDir := filepath.Join(distDir, comp.Name)
	slog.Info("Processing component", "name", comp.Name, "path", compDir, "pkgs", len(filteredPkgs))
	if err := writePackages(compDir, filteredPkgs, comp.KeepVersions); err != nil {
		return fmt.Errorf("publish: %w", err)
	}
	return nil
}

func filterPackages(comp ConfigComponent, pkgs []*packageFile) ([]*packageFile, error) {
	var filteredPkgs []*packageFile
	for _, pat := range comp.FilePatterns {
		patExp, err := regexp.Compile(pat)
		if err != nil {
			return nil, err
		}
		for _, pkg := range pkgs {
			if patExp.MatchString(pkg.Filename) && !slices.Contains(filteredPkgs, pkg) {
				filteredPkgs = append(filteredPkgs, pkg)
			}
		}
	}
	return filteredPkgs, nil
}
