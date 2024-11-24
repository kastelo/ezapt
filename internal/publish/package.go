package publish

import (
	"cmp"
	"compress/gzip"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	"golang.org/x/mod/semver"
	"pault.ag/go/debian/deb"
	"pault.ag/go/debian/version"
)

func newPackage(name string) (*packageFile, error) {
	fd, err := os.Open(name)
	if err != nil {
		return nil, fmt.Errorf("reading package: %w", err)
	}

	info, err := fd.Stat()
	if err != nil {
		return nil, fmt.Errorf("reading package: %w", err)
	}

	hashes, err := newHashes(fd, info.Size())
	if err != nil {
		return nil, fmt.Errorf("reading package: %w", err)
	}

	df, err := deb.Load(fd, name)
	if err != nil {
		return nil, fmt.Errorf("reading package: %w", err)
	}

	return &packageFile{
		Filename: name,
		Control:  &df.Control,
		hashes:   hashes,
	}, nil
}

type packageFile struct {
	Filename string
	*deb.Control
	*hashes
}

func (p *packageFile) MarshalTo(w io.Writer) error {
	return packageTemplate.Execute(w, p)
}

var packageTemplate = template.Must(template.New("package").Parse(`Package: {{.Package}}
Version: {{.Version}}
Architecture: {{.Architecture}}
Maintainer: {{.Maintainer}}
Installed-Size: {{.InstalledSize}}
Depends: {{.Depends}}
Filename: {{.Filename}}
Size: {{.Size}}
MD5sum: {{.MD5sum}}
SHA1: {{.SHA1}}
SHA256: {{.SHA256}}
SHA512: {{.SHA512}}
Section: default
Priority: {{.Priority}}
Homepage: {{.Homepage}}
Description: {{.Description}}
License: MPL-2
Vendor: {{.Maintainer}}

`))

func scanPackages(repo string) (map[string][]*packageFile, error) {
	// directory -> set of packages
	packages := make(map[string][]*packageFile)

	parent := filepath.Dir(repo)
	err := filepath.Walk(repo, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(parent, path)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".deb" {
			return nil
		}

		slog.Info("Scanning package", "path", relPath)
		pkg, err := newPackage(path)
		if err != nil {
			return err
		}
		pkg.Filename = relPath
		packagesDir := filepath.Dir(path)
		packages[packagesDir] = append(packages[packagesDir], pkg)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scanning packages: %w", err)
	}
	return packages, nil
}

func trimPackages(repo string, packages map[string][]*packageFile, keepVersions int) {
	parent := filepath.Dir(repo)
	for dir, pkgs := range packages {
		perPkg := make(map[string][]*packageFile)
		for _, pkg := range pkgs {
			perPkg[pkg.Package] = append(perPkg[pkg.Package], pkg)
		}
		packages[dir] = nil
		for _, pkgs := range perPkg {
			slices.SortFunc(pkgs, func(a, b *packageFile) int {
				return compareVersions(a.Version, b.Version)
			})
			if len(pkgs) > keepVersions {
				for _, pkg := range pkgs[keepVersions:] {
					slog.Info("Removing package", "package", pkg.Package, "version", pkg.Version)
					if err := os.Remove(filepath.Join(parent, pkg.Filename)); err != nil {
						slog.Error("Failed to remove package", "package", pkg.Package, "version", pkg.Version, "error", err)
					}
				}
				pkgs = pkgs[:keepVersions]
			}
			packages[dir] = append(packages[dir], pkgs...)
		}
	}
}

func writePackages(repo string, packages map[string][]*packageFile) error {
	parent := filepath.Dir(repo)
	for dir, pkgs := range packages {
		relPath, err := filepath.Rel(parent, dir)
		if err != nil {
			return err
		}
		slog.Info("Writing packages", "dir", relPath)

		slices.SortFunc(pkgs, func(a, b *packageFile) int {
			if r := cmp.Compare(a.Package, b.Package); r != 0 {
				return r
			}
			return compareVersions(a.Version, b.Version)
		})

		pkgFile, err := os.Create(filepath.Join(dir, "Packages"))
		if err != nil {
			return err
		}

		pkgGzFile, err := os.Create(filepath.Join(dir, "Packages.gz"))
		if err != nil {
			return err
		}

		gw := gzip.NewWriter(pkgGzFile)
		mw := io.MultiWriter(pkgFile, gw)

		for _, pkg := range pkgs {
			if err := pkg.MarshalTo(mw); err != nil {
				return err
			}
		}

		if err := pkgFile.Close(); err != nil {
			return err
		}
		if err := gw.Close(); err != nil {
			return err
		}
		if err := pkgGzFile.Close(); err != nil {
			return err
		}
	}

	return nil
}

func compareVersions(a, b version.Version) int {
	if a.Epoch != b.Epoch {
		return -cmp.Compare(a.Epoch, b.Epoch)
	}
	av := "v" + strings.ReplaceAll(a.String(), "~", "-")
	bv := "v" + strings.ReplaceAll(b.String(), "~", "-")
	return -semver.Compare(av, bv)
}
