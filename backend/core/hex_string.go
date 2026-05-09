package core

import "encoding/hex"

func GetHexString(random_bytes []byte) string {

	s := hex.EncodeToString(random_bytes)
	return s
}
