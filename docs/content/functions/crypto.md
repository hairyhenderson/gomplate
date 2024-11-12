---
title: crypto functions
menu:
  main:
    parent: functions
---

A set of crypto-related functions to be able to perform hashing and (simple!)
encryption operations with `gomplate`.

_Note: These functions are mostly wrappers of existing functions in the Go
standard library. The authors of gomplate are not cryptographic experts,
however, and so can not guarantee correctness of implementation. It is
recommended to have your resident security experts inspect gomplate's code
before using gomplate for critical security infrastructure!_

## `crypto.SSH`

Namespace for ssh functions

_Added in gomplate [v4.2.0](https://github.com/hairyhenderson/gomplate/releases/tag/v4.2.0)_
### Usage

```
crypto.SSH
```


### Examples

```console
$ gomplate -i '{{ crypto.ssh }}'
<namespace SSH [PublicKey]>
```

## `crypto.SSH.PublicKey`

Loads [Secure Shell](https://en.wikipedia.org/wiki/Secure_Shell) public key

_Added in gomplate [v4.2.0](https://github.com/hairyhenderson/gomplate/releases/tag/v4.2.0)_
### Usage

```
crypto.SSH.PublicKey [name]
```
```
name | crypto.SSH.PublicKey
```

### Arguments

| name | description |
|------|-------------|
| `name` | _(optional)_ the name of the key in `~/.ssh` or the absolute path to it. The default value is defined by `IdentityFile` in `~/.ssh/config`. If not specified, `~/.ssh/id_rsa.pub` is used. |

### Examples

```console
$ cat ~/.ssh/id_rsa.pub
gxAedO6GSFC7X+feNqKydIqKlq82R9cnjJPuPLbVvWPB+r08PeJobl++6d9m8EQorpokS+ntqnr35QnIBDWLHk139KhWkOjDOvUHJd6pjOOLhSVapmKPOz1dST4QCweET59STvLHHjNVQfJtWI9zVl4X9S4SoiLDkUUyge+9UnqyA9bAr2P4NkVWZYgf3QnrqoWpRGHz1F7JgV+VmGOlh/Kmc6Q== email@example.com

$ gomplate -i '{{ crypto.SSH.PublicKey }}'
<namespace PublicKey [Blob Comment Format Marshal]>
```
```console
$ gomplate -i '{{ crypto.SSH.PublicKey.Marshal }}'
ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCnEosV4dTgI6CL4YgM4Tfzs6CKdvLL/tarxipWrgEcdwn0TqFn3PmvxSOQWXbQci1Rl2I+U6X3Z4qQ3fafEOlF/bDbwfnY/eUpr9dHnVe1FCbX0tVzCR7OMHg7vGnF3Mta5E9MXMBKupiukgH51hH6fosr90Cvuhj0vsmO3jQL+i1yQxgbc14RCMQuIUZqAA/1Y9JWtucYe4X2uRyby/m2qtHA08kjPTREVd1cMSTM6rCdxnjXgJn7I416ybWnNIwwYeU8q2aKNPIhndSnIBMdDQnnxRCQHgWZXGjF8K8dVl1r3lJWbg/XMXKDWwLXbhRXZwR7/6HDamsV9fkY5Sld9VfKesNiCjaWLlnbe3d6NbdveBcBO6DgDFcshvvtOyu4quBly8EJFpyfeo5V8XQTIVMcLxehXMZNlk0C0PGKQx4xHdxTwFw9IFPbuGNRqRIRwC0YEH3TR4+xBp/gxAedO6GSFC7X+feNqKydIqKlq82R9cnjJPuPLbVvWPB+r08PeJobl++6d9m8EQorpokS+ntqnr35QnIBDWLHk139KhWkOjDOvUHJd6pjOOLhSVapmKPOz1dST4QCweET59STvLHHjNVQfJtWI9zVl4X9S4SoiLDkUUyge+9UnqyA9bAr2P4NkVWZYgf3QnrqoWpRGHz1F7JgV+VmGOlh/Kmc6Q== email@example.com
```
```console
$ gomplate -i '{{ crypto.SSH.PublicKey.Blob | base64.Encode }}'
AAAAB3NzaC1yc2EAAAADAQABAAACAQCnEosV4dTgI6CL4YgM4Tfzs6CKdvLL/tarxipWrgEcdwn0TqFn3PmvxSOQWXbQci1Rl2I+U6X3Z4qQ3fafEOlF/bDbwfnY/eUpr9dHnVe1FCbX0tVzCR7OMHg7vGnF3Mta5E9MXMBKupiukgH51hH6fosr90Cvuhj0vsmO3jQL+i1yQxgbc14RCMQuIUZqAA/1Y9JWtucYe4X2uRyby/m2qtHA08kjPTREVd1cMSTM6rCdxnjXgJn7I416ybWnNIwwYeU8q2aKNPIhndSnIBMdDQnnxRCQHgWZXGjF8K8dVl1r3lJWbg/XMXKDWwLXbhRXZwR7/6HDamsV9fkY5Sld9VfKesNiCjaWLlnbe3d6NbdveBcBO6DgDFcshvvtOyu4quBly8EJFpyfeo5V8XQTIVMcLxehXMZNlk0C0PGKQx4xHdxTwFw9IFPbuGNRqRIRwC0YEH3TR4+xBp/gxAedO6GSFC7X+feNqKydIqKlq82R9cnjJPuPLbVvWPB+r08PeJobl++6d9m8EQorpokS+ntqnr35QnIBDWLHk139KhWkOjDOvUHJd6pjOOLhSVapmKPOz1dST4QCweET59STvLHHjNVQfJtWI9zVl4X9S4SoiLDkUUyge+9UnqyA9bAr2P4NkVWZYgf3QnrqoWpRGHz1F7JgV+VmGOlh/Kmc6Q==
```
```console
$ gomplate -i '{{ crypto.SSH.PublicKey.Comment }}'
email@example.com
```
```console
$ gomplate -i '{{ crypto.SSH.PublicKey.Format }}'
  ssh-rsa
```
```console
$ gomplate -i '{{ (crypto.SSH.PublicKey "e2e_id_ed25519").Marshal }}'
ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBCLlDopq1aotlRUMw6oJ7Snr+qa+r5X8qxADTuYJumN e2e_key
```

## `crypto.Bcrypt`

Uses the [bcrypt](https://en.wikipedia.org/wiki/Bcrypt) password hashing algorithm to generate the hash of a given string. Wraps the [`golang.org/x/crypto/brypt`](https://godoc.org/golang.org/x/crypto/bcrypt) package.

_Added in gomplate [v2.6.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.6.0)_
### Usage

```
crypto.Bcrypt [cost] input
```
```
input | crypto.Bcrypt [cost]
```

### Arguments

| name | description |
|------|-------------|
| `cost` | _(optional)_ the cost, as a number from `4` to `31` - defaults to `10` |
| `input` | _(required)_ the input to hash, usually a password |

### Examples

```console
$ gomplate -i '{{ "foo" | crypto.Bcrypt }}'
$2a$10$jO8nKZ1etGkKK7I3.vPti.fYDAiBqwazQZLUhaFoMN7MaLhTP0SLy
```
```console
$ gomplate -i '{{ crypto.Bcrypt 4 "foo" }}
$2a$04$zjba3N38sjyYsw0Y7IRCme1H4gD0MJxH8Ixai0/sgsrf7s1MFUK1C
```

## `crypto.DecryptAES` _(experimental)_
**Experimental:** This function is [_experimental_][experimental] and may be enabled with the [`--experimental`][experimental] flag.

[experimental]: ../config/#experimental

Decrypts the given input using the given key. By default,
uses AES-256-CBC, but supports 128- and 192-bit keys as well.

This function prints the output as a string. Note that this may result in
unreadable text if the decrypted payload is binary. See
[`crypto.DecryptAESBytes`](#cryptodecryptaesbytes-_experimental_) for another method.

This function is suitable for decrypting data that was encrypted by
Helm's `encryptAES` function, when the input is base64-decoded, and when
using 256-bit keys.

_Added in gomplate [v3.11.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.11.0)_
### Usage

```
crypto.DecryptAES key [keyBits] input
```
```
input | crypto.DecryptAES key [keyBits]
```

### Arguments

| name | description |
|------|-------------|
| `key` | _(required)_ the key to use for decryption |
| `keyBits` | _(optional)_ the key length to use - defaults to `256` |
| `input` | _(required)_ the input to decrypt |

### Examples

```console
$ gomplate -i '{{ base64.Decode "Gp2WG/fKOUsVlhcpr3oqgR+fRUNBcO1eZJ9CW+gDI18=" | crypto.DecryptAES "swordfish" 128 }}'
hello world
```

## `crypto.DecryptAESBytes` _(experimental)_
**Experimental:** This function is [_experimental_][experimental] and may be enabled with the [`--experimental`][experimental] flag.

[experimental]: ../config/#experimental

Decrypts the given input using the given key. By default,
uses AES-256-CBC, but supports 128- and 192-bit keys as well.

This function outputs the raw byte array, which may be sent as input to
other functions.

This function is suitable for decrypting data that was encrypted by
Helm's `encryptAES` function, when the input is base64-decoded, and when
using 256-bit keys.

_Added in gomplate [v3.11.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.11.0)_
### Usage

```
crypto.DecryptAESBytes key [keyBits] input
```
```
input | crypto.DecryptAESBytes key [keyBits]
```

### Arguments

| name | description |
|------|-------------|
| `key` | _(required)_ the key to use for decryption |
| `keyBits` | _(optional)_ the key length to use - defaults to `256` |
| `input` | _(required)_ the input to decrypt |

### Examples

```console
$ gomplate -i '{{ base64.Decode "Gp2WG/fKOUsVlhcpr3oqgR+fRUNBcO1eZJ9CW+gDI18=" | crypto.DecryptAES "swordfish" 128 }}'
hello world
```

## `crypto.EncryptAES` _(experimental)_
**Experimental:** This function is [_experimental_][experimental] and may be enabled with the [`--experimental`][experimental] flag.

[experimental]: ../config/#experimental

Encrypts the given input using the given key. By default,
uses AES-256-CBC, but supports 128- and 192-bit keys as well.

This function is suitable for encrypting data that will be decrypted by
Helm's `decryptAES` function, when the output is base64-encoded, and when
using 256-bit keys.

_Added in gomplate [v3.11.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.11.0)_
### Usage

```
crypto.EncryptAES key [keyBits] input
```
```
input | crypto.EncryptAES key [keyBits]
```

### Arguments

| name | description |
|------|-------------|
| `key` | _(required)_ the key to use for encryption |
| `keyBits` | _(optional)_ the key length to use - defaults to `256` |
| `input` | _(required)_ the input to encrypt |

### Examples

```console
$ gomplate -i '{{ "hello world" | crypto.EncryptAES "swordfish" 128 | base64.Encode }}'
MnRutHovsh/9JN3YrJtBVjZtI6xXZh33bCQS2iZ4SDI=
```

## `crypto.ECDSAGenerateKey` _(experimental)_
**Experimental:** This function is [_experimental_][experimental] and may be enabled with the [`--experimental`][experimental] flag.

[experimental]: ../config/#experimental

Generate a new Elliptic Curve Private Key and output in
PEM-encoded PKCS#1 ASN.1 DER form.

Go's standard NIST P-224, P-256, P-384, and P-521 elliptic curves are all
supported.

Default curve is P-256 and can be overridden with the optional `curve`
parameter.

_Added in gomplate [v3.11.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.11.0)_
### Usage

```
crypto.ECDSAGenerateKey [curve]
```
```
curve | crypto.ECDSAGenerateKey
```

### Arguments

| name | description |
|------|-------------|
| `curve` | _(optional)_ One of Go's standard NIST curves, P-224, P-256, P-384, or P-521 -
defaults to P-256.
 |

### Examples

```console
$ gomplate -i '{{ crypto.ECDSAGenerateKey }}'
-----BEGIN EC PRIVATE KEY-----
...
```

## `crypto.ECDSADerivePublicKey` _(experimental)_
**Experimental:** This function is [_experimental_][experimental] and may be enabled with the [`--experimental`][experimental] flag.

[experimental]: ../config/#experimental

Derive a public key from an elliptic curve private key and output in PKIX
ASN.1 DER form.

_Added in gomplate [v3.11.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.11.0)_
### Usage

```
crypto.ECDSADerivePublicKey key
```
```
key | crypto.ECDSADerivePublicKey
```

### Arguments

| name | description |
|------|-------------|
| `key` | _(required)_ the private key to derive a public key from |

### Examples

```console
$ gomplate -i '{{ crypto.ECDSAGenerateKey | crypto.ECDSADerivePublicKey }}'
-----BEGIN PUBLIC KEY-----
...
```
```console
$ gomplate -d key=priv.pem -i '{{ crypto.ECDSADerivePublicKey (include "key") }}'
-----BEGIN PUBLIC KEY-----
MIGbMBAGByqGSM49AgEGBSuBBAAjA4GGAAQBZvTS1wcCJSsGYQUVoSVctynkuhke
kikB38iNwx/80jzdm+Z8OmRGlwH6OE9NX1MyxjvYMimhcj6zkaOKh1/HhMABrfuY
+hIz6+EUt/Db51awO7iCuRly5L4TZ+CnMAsIbtUOqsqwSQDtv0AclAuogmCst75o
aztsmrD79OXXnhUlURI=
-----END PUBLIC KEY-----
```

## `crypto.Ed25519GenerateKey` _(experimental)_
**Experimental:** This function is [_experimental_][experimental] and may be enabled with the [`--experimental`][experimental] flag.

[experimental]: ../config/#experimental

Generate a new Ed25519 Private Key and output in
PEM-encoded PKCS#8 ASN.1 DER form.

_Added in gomplate [v4.0.0](https://github.com/hairyhenderson/gomplate/releases/tag/v4.0.0)_
### Usage

```
crypto.Ed25519GenerateKey
```


### Examples

```console
$ gomplate -i '{{ crypto.Ed25519GenerateKey }}'
-----BEGIN PRIVATE KEY-----
...
```

## `crypto.Ed25519GenerateKeyFromSeed` _(experimental)_
**Experimental:** This function is [_experimental_][experimental] and may be enabled with the [`--experimental`][experimental] flag.

[experimental]: ../config/#experimental

Generate a new Ed25519 Private Key from a random seed and output in
PEM-encoded PKCS#8 ASN.1 DER form.

_Added in gomplate [v4.0.0](https://github.com/hairyhenderson/gomplate/releases/tag/v4.0.0)_
### Usage

```
crypto.Ed25519GenerateKeyFromSeed encoding seed
```
```
seed | crypto.Ed25519GenerateKeyFromSeed encoding
```

### Arguments

| name | description |
|------|-------------|
| `encoding` | _(required)_ the encoding that the seed is in (`hex` or `base64`) |
| `seed` | _(required)_ the random seed encoded in either base64 or hex |

### Examples

```console
$ gomplate -i '{{ crypto.Ed25519GenerateKeyFromSeed "base64" "MDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDA=" }}'
-----BEGIN PRIVATE KEY-----
...
```

## `crypto.Ed25519DerivePublicKey` _(experimental)_
**Experimental:** This function is [_experimental_][experimental] and may be enabled with the [`--experimental`][experimental] flag.

[experimental]: ../config/#experimental

Derive a public key from an Ed25519 private key and output in PKIX
ASN.1 DER form.

_Added in gomplate [v4.0.0](https://github.com/hairyhenderson/gomplate/releases/tag/v4.0.0)_
### Usage

```
crypto.Ed25519DerivePublicKey key
```
```
key | crypto.Ed25519DerivePublicKey
```

### Arguments

| name | description |
|------|-------------|
| `key` | _(required)_ the private key to derive a public key from |

### Examples

```console
$ gomplate -i '{{ crypto.Ed25519GenerateKey | crypto.Ed25519DerivePublicKey }}'
-----BEGIN PUBLIC KEY-----
...
```
```console
$ gomplate -d key=priv.pem -i '{{ crypto.Ed25519DerivePublicKey (include "key") }}'
-----BEGIN PUBLIC KEY-----
...PK
```

## `crypto.PBKDF2`

Run the Password-Based Key Derivation Function &num;2 as defined in
[RFC 8018 (PKCS &num;5 v2.1)](https://tools.ietf.org/html/rfc8018#section-5.2).

This function outputs the binary result as a hexadecimal string.

_Added in gomplate [v2.3.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.3.0)_
### Usage

```
crypto.PBKDF2 password salt iter keylen [hashfunc]
```

### Arguments

| name | description |
|------|-------------|
| `password` | _(required)_ the password to use to derive the key |
| `salt` | _(required)_ the salt |
| `iter` | _(required)_ iteration count |
| `keylen` | _(required)_ desired length of derived key |
| `hashfunc` | _(optional)_ the hash function to use - must be one of the allowed functions (either in the SHA-1 or SHA-2 sets). Defaults to `SHA-1` |

### Examples

```console
$ gomplate -i '{{ crypto.PBKDF2 "foo" "bar" 1024 8 }}'
32c4907c3c80792b
```

## `crypto.RSADecrypt` _(experimental)_
**Experimental:** This function is [_experimental_][experimental] and may be enabled with the [`--experimental`][experimental] flag.

[experimental]: ../config/#experimental

Decrypt an RSA-encrypted input and print the output as a string. Note that
this may result in unreadable text if the decrypted payload is binary. See
[`crypto.RSADecryptBytes`](#cryptorsadecryptbytes-_experimental_) for a safer method.

The private key must be a PEM-encoded RSA private key in PKCS#1, ASN.1 DER
form, which typically begins with `-----BEGIN RSA PRIVATE KEY-----`.

The input text must be plain ciphertext, as a byte array, or safely
convertible to a byte array. To decrypt base64-encoded input, you must
first decode with the [`base64.DecodeBytes`](../base64/#base64decodebytes)
function.

_Added in gomplate [v3.8.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.8.0)_
### Usage

```
crypto.RSADecrypt key input
```
```
input | crypto.RSADecrypt key
```

### Arguments

| name | description |
|------|-------------|
| `key` | _(required)_ the private key to decrypt the input with |
| `input` | _(required)_ the encrypted input |

### Examples

```console
$ gomplate -c pubKey=./testPubKey -c privKey=./testPrivKey \
  -i '{{ $enc := "hello" | crypto.RSAEncrypt .pubKey -}}
  {{ crypto.RSADecrypt .privKey $enc }}'
hello
```
```console
$ export ENCRYPTED="ScTcX1NZ6p/EeDIf6R7FKLcDFjvP98YgiBhyhPE4jtehajIyTKP1GL8C72qbAWrgdQ6A2cSVjoyo3viqf/PZxpcBDUUMDJuemTaJqUUjMWaDuPG37mQbmRtcvFTuUhw1qSbKyHorDOgTX5d4DvWV4otycGtBT6dXhnmmb5V72J/w3z68vtTJ21m9wREFD7LrYVHdFFtRZiIyMBAF0ngQ+hcujrxilnmgzPkEAg6E7Ccctn28Ie2c4CojrwRbNNxXNlIWCCkC/8Vq8qlDfZ70a+BsTmJDuScE6BZbTyteo9uGYrLn+bTIHNDj90AeLCKUTyWLUJ5Edi9LhlKVBoJUNQ=="
$ gomplate -c ciphertext=env:///ENCRYPTED -c privKey=./testPrivKey \
  -i '{{ base64.DecodeBytes .ciphertext | crypto.RSADecrypt .privKey }}'
hello
```

## `crypto.RSADecryptBytes` _(experimental)_
**Experimental:** This function is [_experimental_][experimental] and may be enabled with the [`--experimental`][experimental] flag.

[experimental]: ../config/#experimental

Decrypt an RSA-encrypted input and output the decrypted byte array.

The private key must be a PEM-encoded RSA private key in PKCS#1, ASN.1 DER
form, which typically begins with `-----BEGIN RSA PRIVATE KEY-----`.

The input text must be plain ciphertext, as a byte array, or safely
convertible to a byte array. To decrypt base64-encoded input, you must
first decode with the [`base64.DecodeBytes`](../base64/#base64decodebytes)
function.

See [`crypto.RSADecrypt`](#cryptorsadecrypt-_experimental_) for a function that outputs
a string.

_Added in gomplate [v3.8.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.8.0)_
### Usage

```
crypto.RSADecryptBytes key input
```
```
input | crypto.RSADecryptBytes key
```

### Arguments

| name | description |
|------|-------------|
| `key` | _(required)_ the private key to decrypt the input with |
| `input` | _(required)_ the encrypted input |

### Examples

```console
$ gomplate -c pubKey=./testPubKey -c privKey=./testPrivKey \
  -i '{{ $enc := "hello" | crypto.RSAEncrypt .pubKey -}}
  {{ crypto.RSADecryptBytes .privKey $enc }}'
[104 101 108 108 111]
```
```console
$ gomplate -c pubKey=./testPubKey -c privKey=./testPrivKey \
  -i '{{ $enc := "hello" | crypto.RSAEncrypt .pubKey -}}
  {{ crypto.RSADecryptBytes .privKey $enc | conv.ToString }}'
hello
```

## `crypto.RSAEncrypt` _(experimental)_
**Experimental:** This function is [_experimental_][experimental] and may be enabled with the [`--experimental`][experimental] flag.

[experimental]: ../config/#experimental

Encrypt the input with RSA and the padding scheme from PKCS#1 v1.5.

This function is suitable for encrypting data that will be decrypted by
[Terraform's `rsadecrypt` function](https://www.terraform.io/docs/configuration/functions/rsadecrypt.html).

The key should be a PEM-encoded RSA public key in PKIX ASN.1 DER form,
which typically begins with `BEGIN PUBLIC KEY`. RSA public keys in PKCS#1
ASN.1 DER form are also supported (beginning with `RSA PUBLIC KEY`).

The output will not be encoded, so consider
[base64-encoding](../base64/#base64encode) it for display.

_Note:_ Output encrypted with this function will _not_ be deterministic,
so encrypting the same input twice will not result in the same ciphertext.

_Warning:_ Using this function may not be safe. See the warning on Go's
[`rsa.EncryptPKCS1v15`](https://pkg.go.dev/crypto/rsa/#EncryptPKCS1v15)
documentation.

_Added in gomplate [v3.8.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.8.0)_
### Usage

```
crypto.RSAEncrypt key input
```
```
input | crypto.RSAEncrypt key
```

### Arguments

| name | description |
|------|-------------|
| `key` | _(required)_ the public key to encrypt the input with |
| `input` | _(required)_ the encrypted input |

### Examples

```console
$ gomplate -c pubKey=./testPubKey \
  -i '{{ "hello" | crypto.RSAEncrypt .pubKey | base64.Encode }}'
ScTcX1NZ6p/EeDIf6R7FKLcDFjvP98YgiBhyhPE4jtehajIyTKP1GL8C72qbAWrgdQ6A2cSVjoyo3viqf/PZxpcBDUUMDJuemTaJqUUjMWaDuPG37mQbmRtcvFTuUhw1qSbKyHorDOgTX5d4DvWV4otycGtBT6dXhnmmb5V72J/w3z68vtTJ21m9wREFD7LrYVHdFFtRZiIyMBAF0ngQ+hcujrxilnmgzPkEAg6E7Ccctn28Ie2c4CojrwRbNNxXNlIWCCkC/8Vq8qlDfZ70a+BsTmJDuScE6BZbTyteo9uGYrLn+bTIHNDj90AeLCKUTyWLUJ5Edi9LhlKVBoJUNQ==
```
```console
$ gomplate -c pubKey=./testPubKey \
  -i '{{ $enc := "hello" | crypto.RSAEncrypt .pubKey -}}
  Ciphertext in hex: {{ printf "%x" $enc }}'
71729b87cccabb248b9e0e5173f0b12c01d9d2a0565bad18aef9d332ce984bde06acb8bb69334a01446f7f6430077f269e6fbf2ccacd972fe5856dd4719252ebddf599948d937d96ea41540dad291b868f6c0cf647dffdb5acb22cd33557f9a1ddd0ee6c1ad2bbafc910ba8f817b66ea0569afc06e5c7858fd9dc2638861fe7c97391b2f190e4c682b4aa2c9b0050081efe18b10aa8c2b2b5f5b68a42dcc06c9da35b37fca9b1509fddc940eb99f516a2e0195405bcb3993f0fa31bc038d53d2e7231dff08cc39448105ed2d0ac52d375cb543ca8a399f807cc5d007e2c44c69876d189667eee66361a393c4916826af77479382838cd4e004b8baa05636805a
```

## `crypto.RSAGenerateKey` _(experimental)_
**Experimental:** This function is [_experimental_][experimental] and may be enabled with the [`--experimental`][experimental] flag.

[experimental]: ../config/#experimental

Generate a new RSA Private Key and output in PEM-encoded PKCS#1 ASN.1 DER
form.

Default key length is 4096 bits, which should be safe enough for most
uses, but can be overridden with the optional `bits` parameter.

In order to protect against [CWE-326](https://cwe.mitre.org/data/definitions/326.html),
keys shorter than `2048` bits may not be generated.

The output is a string, suitable for use with the other `crypto.RSA*`
functions.

_Added in gomplate [v3.8.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.8.0)_
### Usage

```
crypto.RSAGenerateKey [bits]
```
```
bits | crypto.RSAGenerateKey
```

### Arguments

| name | description |
|------|-------------|
| `bits` | _(optional)_ Length in bits of the generated key. Must be at least `2048`. Defaults to `4096` |

### Examples

```console
$ gomplate -i '{{ crypto.RSAGenerateKey }}'
-----BEGIN RSA PRIVATE KEY-----
...
```
```console
$ gomplate -i '{{ $key := crypto.RSAGenerateKey 2048 -}}
  {{ $pub := crypto.RSADerivePublicKey $key -}}
  {{ $enc := "hello" | crypto.RSAEncrypt $pub -}}
  {{ crypto.RSADecrypt $key $enc }}'
hello
```

## `crypto.RSADerivePublicKey` _(experimental)_
**Experimental:** This function is [_experimental_][experimental] and may be enabled with the [`--experimental`][experimental] flag.

[experimental]: ../config/#experimental

Derive a public key from an RSA private key and output in PKIX ASN.1 DER
form.

The output is a string, suitable for use with other `crypto.RSA*`
functions.

_Added in gomplate [v3.8.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.8.0)_
### Usage

```
crypto.RSADerivePublicKey key
```
```
key | crypto.RSADerivePublicKey
```

### Arguments

| name | description |
|------|-------------|
| `key` | _(required)_ the private key to derive a public key from |

### Examples

```console
$ gomplate -i '{{ crypto.RSAGenerateKey | crypto.RSADerivePublicKey }}'
-----BEGIN PUBLIC KEY-----
...
```
```console
$ gomplate -c privKey=./privKey.pem \
  -i '{{ $pub := crypto.RSADerivePublicKey .privKey -}}
  {{ $enc := "hello" | crypto.RSAEncrypt $pub -}}
  {{ crypto.RSADecrypt .privKey $enc }}'
hello
```

## `crypto.SHA1`, `crypto.SHA224`, `crypto.SHA256`, `crypto.SHA384`, `crypto.SHA512`, `crypto.SHA512_224`, `crypto.SHA512_256`

Compute a checksum with a SHA-1 or SHA-2 algorithm as defined in [RFC 3174](https://tools.ietf.org/html/rfc3174) (SHA-1) and [FIPS 180-4](http://nvlpubs.nist.gov/nistpubs/FIPS/NIST.FIPS.180-4.pdf) (SHA-2).

These functions output the binary result as a hexadecimal string.

_Warning: SHA-1 is cryptographically broken and should not be used for secure applications._

_Added in gomplate [v2.3.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.3.0)_
### Usage
```
crypto.SHA1 input
crypto.SHA224 input
crypto.SHA256 input
crypto.SHA384 input
crypto.SHA512 input
crypto.SHA512_224 input
crypto.SHA512_256 input
```

### Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ the data to hash - can be binary data or text |

### Examples

```console
$ gomplate -i '{{ crypto.SHA1 "foo" }}'
f1d2d2f924e986ac86fdf7b36c94bcdf32beec15
```
```console
$ gomplate -i '{{ crypto.SHA512 "bar" }}'
cc06808cbbee0510331aa97974132e8dc296aeb795be229d064bae784b0a87a5cf4281d82e8c99271b75db2148f08a026c1a60ed9cabdb8cac6d24242dac4063
```

## `crypto.SHA1Bytes`, `crypto.SHA224Bytes`, `crypto.SHA256Bytes`, `crypto.SHA384Bytes`, `crypto.SHA512Bytes`, `crypto.SHA512_224Bytes`, `crypto.SHA512_256Bytes`

Compute a checksum with a SHA-1 or SHA-2 algorithm as defined in [RFC 3174](https://tools.ietf.org/html/rfc3174) (SHA-1) and [FIPS 180-4](http://nvlpubs.nist.gov/nistpubs/FIPS/NIST.FIPS.180-4.pdf) (SHA-2).

These functions output the raw binary result, suitable for piping to other functions.

_Warning: SHA-1 is cryptographically broken and should not be used for secure applications._

_Added in gomplate [v3.11.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.11.0)_
### Usage
```
crypto.SHA1Bytes input
crypto.SHA224Bytes input
crypto.SHA256Bytes input
crypto.SHA384Bytes input
crypto.SHA512Bytes input
crypto.SHA512_224Bytes input
crypto.SHA512_256Bytes input
```

### Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ the data to hash - can be binary data or text |

### Examples

```console
$ gomplate -i '{{ crypto.SHA256Bytes "foo" | base64.Encode }}'
LCa0a2j/xo/5m0U8HTBBNBNCLXBkg7+g+YpeiGJm564=
```

## `crypto.WPAPSK`

This is really an alias to [`crypto.PBKDF2`](#cryptopbkdf2) with the
values necessary to convert ASCII passphrases to the WPA pre-shared keys for use with WiFi networks.

This can be used, for example, to help generate a configuration for [wpa_supplicant](http://w1.fi/wpa_supplicant/).

_Added in gomplate [v2.3.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.3.0)_
### Usage

```
crypto.WPAPSK ssid password
```

### Arguments

| name | description |
|------|-------------|
| `ssid` | _(required)_ the WiFi SSID (network name) - must be less than 32 characters |
| `password` | _(required)_ the password - must be between 8 and 63 characters |

### Examples

```console
$ PW=abcd1234 gomplate -i '{{ crypto.WPAPSK "mynet" (getenv "PW") }}'
2c201d66f01237d17d4a7788051191f31706844ac3ffe7547a66c902f2900d34
```
