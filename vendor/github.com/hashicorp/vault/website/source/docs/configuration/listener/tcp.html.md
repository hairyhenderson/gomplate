---
layout: "docs"
page_title: "TCP - Listeners - Configuration"
sidebar_current: "docs-configuration-listener-tcp"
description: |-
  The TCP listener configures Vault to listen on the specified TCP address and
  port.
---

# `tcp` Listener

The TCP listener configures Vault to listen on a TCP address/port.

```hcl
listener "tcp" {
  address = "127.0.0.1:8200"
}
```

## `tcp` Listener Parameters

- `address` `(string: "127.0.0.1:8200")` – Specifies the address to bind to for
  listening.

- `cluster_address` `(string: "127.0.0.1:8201")` – Specifies the address to bind
  to for cluster server-to-server requests. This defaults to one port higher
  than the value of `address`. This does not usually need to be set, but can be
  useful in case Vault servers are isolated from each other in such a way that
  they need to hop through a TCP load balancer or some other scheme in order to
  talk.

- `tls_disable` `(string: "false")` – Specifies if TLS will be disabled. Vault
  assumes TLS by default, so you must explicitly disable TLS to opt-in to
  insecure communication.

- `tls_cert_file` `(string: <required-if-enabled>, reloads-on-SIGHUP)` –
  Specifies the path to the certificate for TLS. To configure the listener to
  use a CA certificate, concatenate the primary certificate and the CA
  certificate together. The primary certificate should appear first in the
  combined file.

- `tls_key_file` `(string: <required-if-enabled>, reloads-on-SIGHUP)` –
  Specifies the path to the private key for the certificate.

- `tls_min_version` `(string: "tls12")` – Specifies the minimum supported
  version of TLS. Accepted values are "tls10", "tls11" or "tls12".

    ~> **Warning**: TLS 1.1 and lower are generally considered insecure.

- `tls_cipher_suites` `(string: "")` – Specifies the list of supported
  ciphersuites as a comma-separated-list. The list of all available ciphersuites
  is available in the [Golang TLS documentation][golang-tls].

- `tls_prefer_server_cipher_suites` `(string: "false")` – Specifies to prefer the
  server's ciphersuite over the client ciphersuites.

- `tls_require_and_verify_client_cert` `(string: "false")` – Turns on client
  authentication for this listener; the listener will require a presented
  client cert that successfully validates against system CAs.

## `tcp` Listener Examples

### Configuring TLS

This example shows enabling a TLS listener.

```hcl
listener "tcp" {
  tls_cert_file = "/etc/certs/nomad.crt"
  tls_key_file  = "/etc/certs/nomad.key"
}
```

[golang-tls]: https://golang.org/src/crypto/tls/cipher_suites.go
