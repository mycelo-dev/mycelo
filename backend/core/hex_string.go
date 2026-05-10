package core

import "encoding/hex"

// GetHexString encodes raw bytes as a lowercase hexadecimal string.
func GetHexString(random_bytes []byte) string {

	s := hex.EncodeToString(random_bytes)
	return s
}
