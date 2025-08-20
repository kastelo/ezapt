package publish

import (
	"compress/gzip"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"
	"time"

	"kastelo.dev/ezapt/internal/pgp"
)

type release struct {
	Codename      string
	Description   string
	Label         string
	Origin        string
	Suite         string
	Architectures []string
	Components    []string
	Date          string
	Files         map[string]*hashes
}

func newRelease() *release {
	return &release{
		Date:        time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 MST"),
		Files:       make(map[string]*hashes),
		Codename:    "debian",
		Description: "Syncthing",
		Label:       "Syncthing",
		Origin:      "Syncthing",
		Suite:       "syncthing",
	}
}

func (r *release) AddFile(name string, hashes *hashes) {
	r.Files[name] = hashes
	parts := strings.Split(name, "/")
	if len(parts) > 1 {
		comp := parts[0]
		if !slices.Contains(r.Components, comp) {
			r.Components = append(r.Components, comp)
		}
		arch := parts[1]
		if strings.HasPrefix(arch, "binary-") {
			arch = arch[7:]
			if !slices.Contains(r.Architectures, arch) {
				r.Architectures = append(r.Architectures, arch)
			}
		}
	}
}

func (r *release) MarshalTo(w io.Writer) error {
	return releaseTemplate.Execute(w, r)
}

var releaseTemplate = template.Must(template.New("release").Parse(`Architectures: {{range .Architectures}}{{.}} {{end}}
Codename: {{.Codename}}
Components: {{range .Components}}{{.}} {{end}}
Date: {{.Date}}
Description: {{.Description}}
Label: {{.Label}}
Origin: {{.Origin}}
Suite: {{.Suite}}
MD5Sum:
{{range $file, $hashes := .Files}} {{$hashes.MD5sum}} {{printf "%16d" $hashes.Size}} {{$file}}
{{end -}}
SHA1:
{{range $file, $hashes := .Files}} {{$hashes.SHA1}} {{printf "%16d" $hashes.Size}} {{$file}}
{{end -}}
SHA256:
{{range $file, $hashes := .Files}} {{$hashes.SHA256}} {{printf "%16d" $hashes.Size}} {{$file}}
{{end -}}
SHA512:
{{range $file, $hashes := .Files}} {{$hashes.SHA512}} {{printf "%16d" $hashes.Size}} {{$file}}
{{end -}}
`))

func writeRelease(dist string) error {
	rel := newRelease()
	err := filepath.Walk(dist, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(dist, path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		switch filepath.Base(path) {
		case "Packages", "Packages.gz":
		default:
			return nil
		}

		slog.Info("Scanning release", "path", relPath)

		fd, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fd.Close()
		hashes, err := newHashes(fd, info.Size())
		if err != nil {
			return err
		}
		rel.AddFile(relPath, hashes)
		return nil
	})
	if err != nil {
		return err
	}

	slog.Info("Writing release", "repo", dist)

	releaseFile, err := os.Create(filepath.Join(dist, "Release"))
	if err != nil {
		return err
	}
	releaseGzFile, err := os.Create(filepath.Join(dist, "Release.gz"))
	if err != nil {
		return err
	}
	gw := gzip.NewWriter(releaseGzFile)
	mw := io.MultiWriter(releaseFile, gw)
	if err := rel.MarshalTo(mw); err != nil {
		return err
	}
	if err := releaseFile.Close(); err != nil {
		return err
	}
	if err := gw.Close(); err != nil {
		return err
	}
	if err := releaseGzFile.Close(); err != nil {
		return err
	}
	return nil
}

func signRelease(dist string, s *pgp.Signer) error {
	in, err := os.Open(filepath.Join(dist, "Release"))
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(filepath.Join(dist, "Release.gpg"))
	if err != nil {
		return err
	}
	if err := s.DetachSign(in, out, false); err != nil {
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	if err := compress(out.Name()); err != nil {
		return err
	}

	if _, err := in.Seek(0, io.SeekStart); err != nil {
		return err
	}

	out, err = os.Create(filepath.Join(dist, "InRelease"))
	if err != nil {
		return err
	}
	if err := s.ClearSign(in, out); err != nil {
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	if err := compress(out.Name()); err != nil {
		return err
	}
	return nil
}

func compress(name string) error {
	fd, err := os.Open(name)
	if err != nil {
		return err
	}
	defer fd.Close()
	gz, err := os.Create(name + ".gz")
	if err != nil {
		return err
	}
	defer gz.Close()
	gw := gzip.NewWriter(gz)
	if _, err := io.Copy(gw, fd); err != nil {
		return err
	}
	if err := gw.Close(); err != nil {
		return err
	}
	return nil
}
