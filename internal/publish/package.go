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
	kept int
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

func scanPackages(root, repo string) ([]*packageFile, error) {
	var packages []*packageFile

	files, err := filepath.Glob(filepath.Join(repo, "*.deb"))
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		info, err := os.Lstat(file)
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			continue
		}

		slog.Info("Scanning package", "path", file)
		pkg, err := newPackage(file)
		if err != nil {
			continue
		}
		pkg.Filename, _ = filepath.Rel(root, file)
		packages = append(packages, pkg)
	}

	return packages, nil
}

func writePackages(repo string, pkgs []*packageFile, keep int) error {
	slices.SortFunc(pkgs, func(a, b *packageFile) int {
		if r := cmp.Compare(a.Package, b.Package); r != 0 {
			return r
		}
		return compareVersions(a.Version, b.Version)
	})

	perArch := make(map[string][]*packageFile)
	for _, pkg := range pkgs {
		arch := pkg.Architecture.String()
		perArch[arch] = append(perArch[arch], pkg)
	}

	for arch, pkgs := range perArch {
		dir := filepath.Join(repo, "binary-"+arch)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
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

		var prevPkg string
		count := 0
		for _, pkg := range pkgs {
			if pkg.Package != prevPkg {
				pkg.kept++
				prevPkg = pkg.Package
				count = 1
			} else if count < keep {
				pkg.kept++
				count++
			} else {
				continue
			}
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
