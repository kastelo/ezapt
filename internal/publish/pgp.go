package publish

import (
	"crypto"
	"fmt"
	"io"

	_ "crypto/sha256"

	_ "golang.org/x/crypto/ripemd160"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/clearsign"
	"golang.org/x/crypto/openpgp/packet"
)

type signer struct {
	keys []*packet.PrivateKey
}

func newSigner(keychain io.Reader) (*signer, error) {
	pr := packet.NewReader(keychain)
	s := &signer{}
	for {
		pkt, err := pr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if key, ok := pkt.(*packet.PrivateKey); ok {
			if !key.IsSubkey && key.PublicKey.PublicKey != nil {
				s.keys = append(s.keys, key)
			}
		}
	}
	return s, nil
}

type seekable interface {
	io.Reader
	io.Seeker
}

func (s *signer) DetachSign(in seekable, out io.Writer) error {
	if len(s.keys) == 0 {
		return fmt.Errorf("no private keys found")
	}
	cfg := &packet.Config{
		DefaultHash: crypto.SHA256,
	}
	for _, key := range s.keys {
		if _, err := in.Seek(0, io.SeekStart); err != nil {
			return err
		}
		signer := &openpgp.Entity{PrivateKey: key}
		if err := openpgp.DetachSign(out, signer, in, cfg); err != nil {
			return err
		}
	}
	return nil
}

func (s *signer) ClearSign(in seekable, out io.Writer) error {
	if len(s.keys) == 0 {
		return fmt.Errorf("no private keys found")
	}

	w, err := clearsign.EncodeMulti(out, s.keys, nil)
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, in); err != nil {
		return err
	}
	return w.Close()
}
