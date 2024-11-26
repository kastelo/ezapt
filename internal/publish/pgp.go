package publish

import (
	"crypto"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"

	"github.com/ProtonMail/go-crypto/openpgp/clearsign"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
	openpgp "github.com/ProtonMail/go-crypto/openpgp/v2"
)

type signer struct {
	entities []*openpgp.Entity
}

func newSigner(keychain io.Reader) (*signer, error) {
	pr := packet.NewReader(keychain)
	s := &signer{}
	for {
		ent, err := openpgp.ReadEntity(pr)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		slog.Info("Loaded key", "fingerprint", hex.EncodeToString(ent.PrimaryKey.Fingerprint))
		s.entities = append(s.entities, ent)
	}
	return s, nil
}

type seekable interface {
	io.Reader
	io.Seeker
}

func (s *signer) DetachSign(in seekable, out io.Writer) error {
	if len(s.entities) == 0 {
		return fmt.Errorf("no entities")
	}
	cfg := &packet.Config{
		DefaultHash: crypto.SHA256,
	}
	if err := openpgp.DetachSign(out, s.entities, in, cfg); err != nil {
		return err
	}
	return nil
}

func (s *signer) ClearSign(in seekable, out io.Writer) error {
	if len(s.entities) == 0 {
		return fmt.Errorf("no entities")
	}

	keys := make([]*packet.PrivateKey, len(s.entities))
	for i, e := range s.entities {
		keys[i] = e.PrivateKey
	}
	w, err := clearsign.EncodeMulti(out, keys, nil)
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, in); err != nil {
		return err
	}
	return w.Close()
}
