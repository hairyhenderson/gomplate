package funcs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hairyhenderson/gomplate/v4/base64"
	"github.com/hairyhenderson/gomplate/v4/conv"
	"github.com/kevinburke/ssh_config"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// CreateSSHFuncs -
func CreateSSHFuncs(ctx context.Context) map[string]any {
	ns := newSSHFuncs(ctx)

	return map[string]any{
		"ssh": func() any { return ns },
	}
}

func newSSHFuncs(ctx context.Context) *SSHFuncs {
	ns := &SSHFuncs{
		ctx: ctx,
		fs:  getFS(ctx),
	}

	ns.self = ns
	return ns
}

type SSHFuncs struct {
	namespace

	ctx context.Context
	fs  fs.FS

	homeDir     string
	homeDirOnce sync.Once

	conf     *sshClientConf
	confOnce sync.Once
}

// PublicKey -
func (f *SSHFuncs) PublicKey(args ...any) (key *PublicKey, err error) {
	slog.Debug("PublicKey start", slog.Any("args", args))
	defer func() { slog.Debug("PublicKey end", slog.Any("key", key), slog.Any("err", err)) }()

	var nameAny any
	if len(args) > 0 {
		nameAny = args[0]
	}

	keyPath, err := f.toPublicKeyPath(nameAny)
	if err != nil {
		return nil, err
	}

	keyFile, err := f.fs.Open(keyPath + ".pub")
	if err != nil {
		return nil, fmt.Errorf("open key file: %w", err)
	}
	defer keyFile.Close()

	//TODO: add support for pem-encoded keys
	rawKeyData, err := io.ReadAll(keyFile)
	if err != nil {
		return nil, fmt.Errorf("read key file: %w", err)
	}

	out, comment, _, _, err := ssh.ParseAuthorizedKey(rawKeyData)
	if err != nil {
		return nil, fmt.Errorf("ssh: parse authorized key: %w", err)
	}

	agentKey := agent.Key{
		Format:  out.Type(),
		Blob:    out.Marshal(),
		Comment: comment,
	}

	return newPublicKey(agentKey), nil
}

type PublicKey struct {
	namespace
	k agent.Key
}

// Format -
func (k PublicKey) Format() string { return k.k.Format }

// Blob -
func (k PublicKey) Blob() []byte { return k.k.Blob }

// Comment -
func (k PublicKey) Comment() string { return k.k.Comment }

// Marshal -
func (k PublicKey) Marshal() string {
	return k.k.String()
}

// Verify -
// TODO: documentation
func (k PublicKey) Verify(args ...any) bool {
	return k.MustVerify(args...) == nil
}

// MustVerify -
// TODO: documentation
func (k PublicKey) MustVerify(args ...any) error {
	var data, signature, format any
	if len(args) > 1 {
		data, signature = args[0], args[1]
	}
	if len(args) > 2 {
		format = args[2]
	}

	sshSig, err := toSSHSignature(signature, format)
	if err != nil {
		return err
	}

	return k.k.Verify(toBytes(data), sshSig)
}

func newPublicKey(k agent.Key) *PublicKey {
	key := &PublicKey{k: k}
	key.self = key
	return key
}

func (f *SSHFuncs) toPublicKeyPath(nameAny any) (string, error) {
	if nameAny == nil {
		conf, err := f.getConfig()
		if err != nil {
			return "", fmt.Errorf("get config: %w", err)
		}

		return conf.IdentityFile, nil
	}

	if path := conv.ToString(nameAny); filepath.IsAbs(path) {
		// we will add it in fs.Open argument anyway
		return strings.TrimSuffix(path, ".pub"), nil
	}

	homeDir, err := f.getHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}

	return filepath.Join(homeDir, ".ssh", conv.ToString(nameAny)), nil

}

type sshClientConf struct {
	IdentityFile string
}

func (c *sshClientConf) setDefaults(homeDir string) *sshClientConf {
	if c.IdentityFile == "" {
		c.IdentityFile = filepath.Join(homeDir, ".ssh", "id_rsa")
	}

	return c
}

func (f *SSHFuncs) getConfig() (*sshClientConf, error) {
	var err error
	f.confOnce.Do(func() {
		var homeDir string
		homeDir, err = f.getHomeDir()
		if err != nil {
			err = fmt.Errorf("get home dir: %w", err)
			return
		}

		f.conf, err = f.getConfigOnce(homeDir)
		if f.conf == nil {
			f.conf = new(sshClientConf)
		}

		f.conf.setDefaults(homeDir)
	})

	return f.conf, err
}

func (f *SSHFuncs) getConfigOnce(homeDir string) (*sshClientConf, error) {
	file, err := f.fs.Open(filepath.Join(homeDir, ".ssh", "config"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open config file: %w", err)
	}
	defer file.Close()

	cfg, err := ssh_config.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("fail to read ssh config: %w", err)
	}

	identityFiles, err := cfg.GetAll("", "IdentityFile")
	if err != nil {
		return nil, fmt.Errorf("get IdentityFile directives: %w", err)
	}

	conf := &sshClientConf{
		IdentityFile: getLast(identityFiles),
	}

	return conf, nil
}

func (f *SSHFuncs) getHomeDir() (string, error) {
	var err error
	if f.homeDir != "" {
		return f.homeDir, nil
	}

	f.homeDirOnce.Do(func() {
		f.homeDir, err = os.UserHomeDir()
	})

	return f.homeDir, err
}

func (f *SSHFuncs) reset() {
	f.homeDirOnce = sync.Once{}
	f.confOnce = sync.Once{}
}

func (f *SSHFuncs) trace(msg string, attrs ...slog.Attr) {
	trace(f.ctx, msg, attrs...)
}

func toSSHSignature(sig, f any) (*ssh.Signature, error) {
	format := conv.ToString(f)
	if f == nil {
		format = ssh.KeyAlgoRSA
	}

	switch sig := sig.(type) {
	case *ssh.Signature:
		return sig, nil
	case string, []byte, byter, fmt.Stringer:
		return &ssh.Signature{Format: format, Blob: maybeBase64toBytes(sig)}, nil
	default:
		return nil, fmt.Errorf("could not convert %T to *ssh.Signature", sig)
	}
}

func getLast[T any](list []T) T {
	if len(list) == 0 {
		var zero T
		return zero
	}

	return list[len(list)-1]
}

func maybeBase64toBytes(in any) []byte {
	decoded, err := base64.Decode(conv.ToString(in))
	if err == nil {
		return decoded
	}

	return toBytes(in)
}
