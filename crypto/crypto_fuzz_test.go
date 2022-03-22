package crypto

import (
	"bytes"
	"testing"
)

func FuzzCryptAESCBC(f *testing.F) {
	key := bytes.Repeat([]byte{'a'}, 16)
	f.Add(key, []byte("some random text to be encrypted"))

	key = bytes.Repeat([]byte{'a'}, 32)
	f.Add(key, []byte(`more random
		text to be
		encrypted, with non-ascii characters:
		ΤΟΙΣ πᾶσι χρόνος καὶ καιρὸς τῷ παντὶ πράγματι ὑπὸ τὸν οὐρανόν. καιρὸς τοῦ
		τεκεῖν καὶ καιρὸς τοῦ ἀποθανεῖν, καιρὸς τοῦ φυτεῦσαι καὶ καιρὸς τοῦ ἐκτῖλαι τὸ
		πεφυτευμένον, καιρὸς τοῦ ἀποκτεῖναι καὶ καιρὸς τοῦ ἰάσασθαι, καιρὸς τοῦ παρακολουθεῖν
		and also:
		何事にも定まった時があります。
		生まれる時、死ぬ時、植える時、収穫の時、育てる時、枯れる時、葉が降る時、落ちる時、落とされる時、
		そして、それらの時には、それぞれの人々は、それぞれの幸福を知っています。
		殺す時、病気が治る時、壊す時、やり直す時.`))

	f.Fuzz(func(t *testing.T, key, in []byte) {
		// make sure we have a key of the right length
		var k []byte
		switch {
		case len(key) <= 16:
			k = make([]byte, 16)
		case len(key) <= 24:
			k = make([]byte, 24)
		default:
			k = make([]byte, 32)
		}
		copy(k, key)

		out, err := EncryptAESCBC(k, in)
		if err != nil {
			t.Error(err)
			return
		}
		if len(in) == 0 && len(out) != 0 {
			t.Error("empty input should return empty output")
			return
		}

		dec, err := DecryptAESCBC(k, out)
		if err != nil {
			t.Error(err)
		}
		if !bytes.Equal(in, dec) {
			t.Errorf("%q != %q (encrypted: %q)", in, dec, out)
		}
	})
}
