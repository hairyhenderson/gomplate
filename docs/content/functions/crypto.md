---
title: crypto functions
menu:
  main:
    parent: functions
---

A set of crypto-related functions to be able to perform hashing and (simple!) encryption operations with `gomplate`.

_Note: These functions are mostly wrappers of existing functions in the Go standard library. The authors of gomplate are not cryptographic experts, however, and so can not guarantee correctness of implementation. It is recommended to have your resident security experts inspect gomplate's code before using gomplate for critical security infrastructure!_

## `crypto.Bcrypt`

Uses the [bcrypt](https://en.wikipedia.org/wiki/Bcrypt) password hashing algorithm to generate the hash of a given string. Wraps the [`golang.org/x/crypto/brypt`](https://godoc.org/golang.org/x/crypto/bcrypt) package.

### Usage

```go
crypto.Bcrypt [cost] input
```
```go
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

## `crypto.PBKDF2`

Run the Password-Based Key Derivation Function &num;2 as defined in
[RFC 8018 (PKCS &num;5 v2.1)](https://tools.ietf.org/html/rfc8018#section-5.2).

This function outputs the binary result as a hexadecimal string.

### Usage

```go
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

## `crypto.SHA1`, `crypto.SHA224`, `crypto.SHA256`, `crypto.SHA384`, `crypto.SHA512`, `crypto.SHA512_224`, `crypto.SHA512_256`

Compute a checksum with a SHA-1 or SHA-2 algorithm as defined in [RFC 3174](https://tools.ietf.org/html/rfc3174) (SHA-1) and [FIPS 180-4](http://nvlpubs.nist.gov/nistpubs/FIPS/NIST.FIPS.180-4.pdf) (SHA-2).

These functions output the binary result as a hexadecimal string.

_Note: SHA-1 is cryptographically broken and should not be used for secure applications._

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

## `crypto.WPAPSK`

This is really an alias to [`crypto.PBKDF2`](#crypto.PBKDF2) with the
values necessary to convert ASCII passphrases to the WPA pre-shared keys for use with WiFi networks.

This can be used, for example, to help generate a configuration for [wpa_supplicant](http://w1.fi/wpa_supplicant/).

### Usage

```go
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
