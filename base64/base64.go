package base64

import (
	b64 "encoding/base64"
)

// Encode - Encode data in base64 format
func Encode(in []byte) (string, error) {
	return b64.StdEncoding.EncodeToString(in), nil
}

// Decode - Decode a base64-encoded string
func Decode(in string) ([]byte, error) {
	o, err := b64.StdEncoding.DecodeString(in)
	if err != nil {
		// maybe it's in the URL variant?
		o, err = b64.URLEncoding.DecodeString(in)
		if err != nil {
			// ok, just give up...
			return nil, err
		}
	}
	return o, nil
}
