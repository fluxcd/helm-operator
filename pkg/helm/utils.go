package helm

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/google/go-cmp/cmp"
)

func Diff(j *Release, k *Release) string {
	return cmp.Diff(j.Values, k.Values) + cmp.Diff(j.Chart, k.Chart)
}

func Checksum(b []byte) string {
	hasher := sha256.New()
	hasher.Write(b)
	return hex.EncodeToString(hasher.Sum(nil))
}
