package funcs

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"io/fs"
	"math/big"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/hairyhenderson/gomplate/v4/internal/datafs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func TestCreateSSHFuncs(t *testing.T) {
	t.Parallel()

	for i := 0; i < 10; i++ {
		// Run this a bunch to catch race conditions
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			fmap := CreateSSHFuncs(ctx)
			actual := fmap["ssh"].(func() any)

			assert.Equal(t, ctx, actual().(*SSHFuncs).ctx)
		})
	}
}

func TestPublicKey(t *testing.T) {
	t.Parallel()

	const (
		idRSAData     = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCnEosV4dTgI6CL4YgM4Tfzs6CKdvLL/tarxipWrgEcdwn0TqFn3PmvxSOQWXbQci1Rl2I+U6X3Z4qQ3fafEOlF/bDbwfnY/eUpr9dHnVe1FCbX0tVzCR7OMHg7vGnF3Mta5E9MXMBKupiukgH51hH6fosr90Cvuhj0vsmO3jQL+i1yQxgbc14RCMQuIUZqAA/1Y9JWtucYe4X2uRyby/m2qtHA08kjPTREVd1cMSTM6rCdxnjXgJn7I416ybWnNIwwYeU8q2aKNPIhndSnIBMdDQnnxRCQHgWZXGjF8K8dVl1r3lJWbg/XMXKDWwLXbhRXZwR7/6HDamsV9fkY5Sld9VfKesNiCjaWLlnbe3d6NbdveBcBO6DgDFcshvvtOyu4quBly8EJFpyfeo5V8XQTIVMcLxehXMZNlk0C0PGKQx4xHdxTwFw9IFPbuGNRqRIRwC0YEH3TR4+xBp/gxAedO6GSFC7X+feNqKydIqKlq82R9cnjJPuPLbVvWPB+r08PeJobl++6d9m8EQorpokS+ntqnr35QnIBDWLHk139KhWkOjDOvUHJd6pjOOLhSVapmKPOz1dST4QCweET59STvLHHjNVQfJtWI9zVl4X9S4SoiLDkUUyge+9UnqyA9bAr2P4NkVWZYgf3QnrqoWpRGHz1F7JgV+VmGOlh/Kmc6Q== email@example.com"
		idED25519Data = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBCLlDopq1aotlRUMw6oJ7Snr+qa+r5X8qxADTuYJumN other_email@example.com\n"
	)

	fsys := fstest.MapFS{
		"home/user/.ssh":                &fstest.MapFile{Mode: fs.ModeDir | 0o777},
		"home/user/.ssh/id_rsa.pub":     &fstest.MapFile{Data: []byte(idRSAData)},
		"home/user/.ssh/id_ed25519.pub": &fstest.MapFile{Data: []byte(idED25519Data)},
	}
	ff := &SSHFuncs{fs: datafs.WrapWdFS(fsys), homeDir: "/home/user"}

	defaultKey, err := ff.PublicKey(nil)
	require.NoError(t, err)
	require.NotNil(t, defaultKey)

	ed25519Key, err := ff.PublicKey("id_ed25519")
	require.NoError(t, err)
	require.NotNil(t, ed25519Key)

	assert.Equal(t, "ssh-rsa", defaultKey.Format())
	assert.Equal(t, idRSAData, defaultKey.Marshal())
	assert.Equal(t, "email@example.com", defaultKey.Comment())

	expectedED25519 := strings.Join(strings.Fields(idED25519Data), " ")
	assert.Equal(t, "ssh-ed25519", ed25519Key.Format())
	assert.Equal(t, expectedED25519, ed25519Key.Marshal())
	assert.Equal(t, "other_email@example.com", ed25519Key.Comment())

	fsys["home/user/.ssh/config"] = &fstest.MapFile{Data: []byte("IdentityFile /home/user/.ssh/id_ed25519")}
	ff.reset()
	defaultKey, err = ff.PublicKey(nil)
	require.NoError(t, err)
	require.NotNil(t, defaultKey)

	assert.Equal(t, "ssh-ed25519", defaultKey.Format())
	assert.Equal(t, expectedED25519, defaultKey.Marshal())
	assert.Equal(t, "other_email@example.com", defaultKey.Comment())

	fsys["home/user/.ssh/config"] = &fstest.MapFile{}
	ff.reset()
	defaultKey, err = ff.PublicKey(nil)
	require.NoError(t, err)
	require.NotNil(t, defaultKey)

	assert.Equal(t, "ssh-rsa", defaultKey.Format())
	assert.Equal(t, idRSAData, defaultKey.Marshal())
	assert.Equal(t, "email@example.com", defaultKey.Comment())
}

func TestRSAPublicKeyVerify(t *testing.T) {
	t.Parallel()

	keyPair, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	publicKey := &keyPair.PublicKey

	fsys := fstest.MapFS{
		"home/user/.ssh":                &fstest.MapFile{Mode: fs.ModeDir | 0o777},
		"home/user/.ssh/config":         &fstest.MapFile{Data: []byte("IdentityFile /home/user/.ssh/id_ed25519")},
		"home/user/.ssh/id_ed25519.pub": &fstest.MapFile{Data: marshalAuthorizedKey(publicKey)},
	}
	ff := &SSHFuncs{fs: datafs.WrapWdFS(fsys), homeDir: "/home/user"}

	defaultKey, err := ff.PublicKey(nil)
	require.NoError(t, err)
	require.NotNil(t, defaultKey)

	msg := []byte("My message")

	msgHash := sha512.New()
	_, err = msgHash.Write(msg)
	require.NoError(t, err)

	digest := msgHash.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, keyPair, crypto.SHA512, digest)
	require.NoError(t, err)

	sshSignature := &ssh.Signature{
		Format: ssh.KeyAlgoRSASHA512,
		Blob:   signature,
	}
	assert.NoError(t, defaultKey.MustVerify(msg, sshSignature, nil))
	assert.NoError(t, defaultKey.MustVerify(msg, sshSignature.Blob, sshSignature.Format))
	assert.NoError(t, defaultKey.MustVerify(msg, base64Encode(sshSignature.Blob), sshSignature.Format))
}

func TestED25519Verify(t *testing.T) {
	t.Parallel()

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	fsys := fstest.MapFS{
		"home/user/.ssh":            &fstest.MapFile{Mode: fs.ModeDir | 0o777},
		"home/user/.ssh/id_rsa.pub": &fstest.MapFile{Data: marshalAuthorizedKey(publicKey)},
	}
	ff := &SSHFuncs{fs: datafs.WrapWdFS(fsys), homeDir: "/home/user"}

	defaultKey, err := ff.PublicKey(nil)
	require.NoError(t, err)
	require.NotNil(t, defaultKey)

	msg := []byte("My message")
	signature := ed25519.Sign(privateKey, msg)

	sshSignature := &ssh.Signature{
		Format: ssh.KeyAlgoED25519,
		Blob:   signature,
	}
	assert.NoError(t, defaultKey.MustVerify(msg, sshSignature, nil))
	assert.NoError(t, defaultKey.MustVerify(msg, sshSignature.Blob, sshSignature.Format))
	assert.NoError(t, defaultKey.MustVerify(msg, base64Encode(sshSignature.Blob), sshSignature.Format))
}

func marshalAuthorizedKey(key any) []byte {
	var sshKey agent.Key

	switch key := key.(type) {
	case *rsa.PublicKey:
		sshKey.Format = ssh.KeyAlgoRSA
		sshKey.Blob = ssh.Marshal(struct {
			Format string
			E      *big.Int
			N      *big.Int
		}{
			sshKey.Format,
			big.NewInt(int64(key.E)),
			key.N,
		})
	case ed25519.PublicKey:
		sshKey.Format = ssh.KeyAlgoED25519
		sshKey.Blob = ssh.Marshal(struct {
			Format string
			Blob   []byte
		}{
			sshKey.Format,
			key,
		})
	//TODO:
	//case *ecdsa.PublicKey:
	//	switch bitSize := key.Params().BitSize; bitSize {
	//	case 256:
	//		sshKey.Format = ssh.KeyAlgoECDSA256
	//	case 384:
	//		sshKey.Format = ssh.KeyAlgoECDSA384
	//	case 521:
	//		sshKey.Format = ssh.KeyAlgoECDSA521
	//	default:
	//		panic(fmt.Errorf("ecdsa: unsupported bit size %d", bitSize))
	//	}
	default:
		panic(fmt.Errorf("unsupported key type %T", key))
	}

	return ssh.MarshalAuthorizedKey(&sshKey)
}

func base64Encode(blob []byte) []byte {
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(blob)))
	base64.StdEncoding.Encode(buf, blob)
	return buf
}
