package publish

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
)

func newHashes(r io.Reader, size int64) (*hashes, error) {
	m5 := md5.New()
	s1 := sha1.New()
	s2 := sha256.New()
	s5 := sha512.New()
	mw := io.MultiWriter(m5, s1, s2, s5)
	if _, err := io.Copy(mw, r); err != nil {
		return nil, fmt.Errorf("hashing: %w", err)
	}

	return &hashes{
		Size:   size,
		MD5sum: hex.EncodeToString(m5.Sum(nil)),
		SHA1:   hex.EncodeToString(s1.Sum(nil)),
		SHA256: hex.EncodeToString(s2.Sum(nil)),
		SHA512: hex.EncodeToString(s5.Sum(nil)),
	}, nil
}

type hashes struct {
	Size   int64
	MD5sum string
	SHA1   string
	SHA256 string
	SHA512 string
}
