package sign

import (
	"fmt"
	"os"

	"kastelo.dev/ezapt/internal/pgp"
)

type CLI struct {
	Files []string `arg:"" help:"Files to sign" type:"existingfile" env:"EZAPT_FILES"`
	pgp.CLI
	Detach bool `help:"Detach signature" env:"EZAPT_DETACH"`
	ASCII  bool `help:"ASCII armor" env:"EZAPT_ASCII"`
}

func (c *CLI) Run() error {
	sign, err := c.Signer()
	if err != nil {
		return fmt.Errorf("sign: %w", err)
	}

	for _, file := range c.Files {
		in, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("open: %w", err)
		}
		defer in.Close()
		out, err := os.Create(file + ".asc")
		if err != nil {
			return fmt.Errorf("create: %w", err)
		}
		if c.Detach {
			err = sign.DetachSign(in, out, c.ASCII)
		} else {
			err = sign.ClearSign(in, out)
		}
		if err != nil {
			return fmt.Errorf("sign: %w", err)
		}
		if err := out.Close(); err != nil {
			return fmt.Errorf("close: %w", err)
		}
	}

	return nil
}
