package base64

import (
	b64 "encoding/base64"
	"log"
)

// Encode - Encode data in base64 format
func Encode(in []byte) string {
	return b64.StdEncoding.EncodeToString(in)
}

// Decode - Decode a base64-encoded string
func Decode(in string) []byte {
	o, err := b64.StdEncoding.DecodeString(in)
	if err != nil {
		// maybe it's in the URL variant?
		o, err = b64.URLEncoding.DecodeString(in)
		if err != nil {
			// ok, just give up...
			log.Fatal(err)
		}
	}
	return o
}
