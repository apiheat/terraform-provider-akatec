package akatec

import (
	"crypto/sha1"
	"encoding/hex"
)

func getSHAString(rdata string) string {
	h := sha1.New()
	h.Write([]byte(rdata))

	sha1hashtest := hex.EncodeToString(h.Sum(nil))
	return sha1hashtest
}
