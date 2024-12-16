package pgp

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
)

type CLI struct {
	Keyring       string `group:"Keyring" help:"Path to GPG keyring" type:"existingfile" env:"EZAPT_KEYRING"`
	KeyringBase64 string `group:"Keyring" help:"GPG keyring as base64" env:"EZAPT_KEYRING_BASE64"`
}

func (c *CLI) Signer() (*Signer, error) {
	var keyringReader io.Reader
	if c.KeyringBase64 != "" {
		keyringReader = base64.NewDecoder(base64.StdEncoding, strings.NewReader(c.KeyringBase64))
	} else if c.Keyring != "" {
		fd, err := os.Open(c.Keyring)
		if err != nil {
			return nil, err
		}
		defer fd.Close()
		keyringReader = fd
	} else {
		return nil, fmt.Errorf("no keyring provided")
	}

	return NewSigner(keyringReader)
}
