---
title: net functions
menu:
  main:
    parent: functions
---

The `net` namespace contains functions that can help deal with network-related
lookups and calculations. Some of these functions return specifically-typed
values that contain additional methods useful for formatting or further
calculations.

[RFC 4632]: http://tools.ietf.org/html/rfc4632
[RFC 4291]: http://tools.ietf.org/html/rfc4291
[`inet.af/netaddr`]: https://pkg.go.dev/inet.af/netaddr

## `net.LookupIP`

Resolve an IPv4 address for a given host name. When multiple IP addresses
are resolved, the first one is returned.

### Usage

```go
net.LookupIP name
```
```go
name | net.LookupIP
```

### Arguments

| name | description |
|------|-------------|
| `name` | _(required)_ The hostname to look up. This can be a simple hostname, or a fully-qualified domain name. |

### Examples

```console
$ gomplate -i '{{ net.LookupIP "example.com" }}'
93.184.216.34
```

## `net.LookupIPs`

Resolve all IPv4 addresses for a given host name. Returns an array of strings.

### Usage

```go
net.LookupIPs name
```
```go
name | net.LookupIPs
```

### Arguments

| name | description |
|------|-------------|
| `name` | _(required)_ The hostname to look up. This can be a simple hostname, or a fully-qualified domain name. |

### Examples

```console
$ gomplate -i '{{ join (net.LookupIPs "twitter.com") "," }}'
104.244.42.65,104.244.42.193
```

## `net.LookupCNAME`

Resolve the canonical name for a given host name. This does a DNS lookup for the
`CNAME` record type. If no `CNAME` is present, a canonical form of the given name
is returned -- e.g. `net.LookupCNAME "localhost"` will return `"localhost."`.

### Usage

```go
net.LookupCNAME name
```
```go
name | net.LookupCNAME
```

### Arguments

| name | description |
|------|-------------|
| `name` | _(required)_ The hostname to look up. This can be a simple hostname, or a fully-qualified domain name. |

### Examples

```console
$ gomplate -i '{{ net.LookupCNAME "www.amazon.com" }}'
d3ag4hukkh62yn.cloudfront.net.
```

## `net.LookupSRV`

Resolve a DNS [`SRV` service record](https://en.wikipedia.org/wiki/SRV_record).
This implementation supports the canonical [RFC2782](https://tools.ietf.org/html/rfc2782)
form (i.e. `_Service._Proto.Name`), but other forms are also supported, such as
those served by [Consul's DNS interface](https://www.consul.io/docs/agent/dns.html#standard-lookup).

When multiple records are returned, this function returns the first.

A [`net.SRV`](https://golang.org/pkg/net/#SRV) data structure is returned. The
following properties are available:
- `Target` - _(string)_ the hostname where the service can be reached
- `Port` - _(uint16)_ the service's port
- `Priority`, `Weight` - see [RFC2782](https://tools.ietf.org/html/rfc2782) for details

### Usage

```go
net.LookupSRV name
```
```go
name | net.LookupSRV
```

### Arguments

| name | description |
|------|-------------|
| `name` | _(required)_ The service name to look up |

### Examples

```console
$ gomplate -i '{{ net.LookupSRV "_sip._udp.sip.voice.google.com" | toJSONPretty "  " }}'
{
  "Port": 5060,
  "Priority": 10,
  "Target": "sip-anycast-1.voice.google.com.",
  "Weight": 1
}
```

## `net.LookupSRVs`

Resolve a DNS [`SRV` service record](https://en.wikipedia.org/wiki/SRV_record).
This implementation supports the canonical [RFC2782](https://tools.ietf.org/html/rfc2782)
form (i.e. `_Service._Proto.Name`), but other forms are also supported, such as
those served by [Consul's DNS interface](https://www.consul.io/docs/agent/dns.html#standard-lookup).

This function returns all available SRV records.

An array of [`net.SRV`](https://golang.org/pkg/net/#SRV) data structures is
returned. For each element, the following properties are available:
- `Target` - _(string)_ the hostname where the service can be reached
- `Port` - _(uint16)_ the service's port
- `Priority`, `Weight` - see [RFC2782](https://tools.ietf.org/html/rfc2782) for details

### Usage

```go
net.LookupSRVs name
```
```go
name | net.LookupSRVs
```

### Arguments

| name | description |
|------|-------------|
| `name` | _(required)_ The hostname to look up. This can be a simple hostname, or a fully-qualified domain name. |

### Examples

_input.tmpl:_
```
{{ range (net.LookupSRVs "_sip._udp.sip.voice.google.com") -}}
priority={{.Priority}}/port={{.Port}}
{{- end }}
```

```console
$ gomplate -f input.tmpl
priority=10/port=5060
priority=20/port=5060
```

## `net.LookupTXT`

Resolve a DNS [`TXT` record](https://en.wikipedia.org/wiki/SRV_record).

This function returns all available TXT records as an array of strings.

### Usage

```go
net.LookupTXT name
```
```go
name | net.LookupTXT
```

### Arguments

| name | description |
|------|-------------|
| `name` | _(required)_ The host name to look up |

### Examples

```console
$ gomplate -i '{{net.LookupTXT "example.com" | data.ToJSONPretty "  " }}'
[
  "v=spf1 -all"
]
```

## `net.ParseIP`

Parse the given string as an IP address (a `netaddr.IP` from the
[`inet.af/netaddr`](https://pkg.go.dev/inet.af/netaddr) package).

Any of `netaddr.IP`'s methods may be called on the resulting value. See
[the docs](https://pkg.go.dev/inet.af/netaddr) for details.

### Usage

```go
net.ParseIP ip
```
```go
ip | net.ParseIP
```

### Arguments

| name | description |
|------|-------------|
| `ip` | _(required)_ The IP string to parse. It must be either an IPv4 or IPv6 address. |

### Examples

```console
$ gomplate -i '{{ (net.ParseIP "192.168.0.1").IsPrivate }}'
true
$ gomplate -i '{{ $ip := net.ParseIP (net.LookupIP "example.com") -}}
  {{ $ip.Prefix 12 }}'
93.176.0.0/12
```

## `net.ParseIPPrefix`

Parse the given string as an IP address prefix (CIDR) representing an IP
network (a `netaddr.IPPrefix` from the
[`inet.af/netaddr`][] package).

The string can be in the form `"192.168.1.0/24"` or `"2001::db8::/32"`,
the CIDR notations defined in [RFC 4632][] and [RFC 4291][].

Any of `netaddr.IPPrefix`'s methods may be called on the resulting value.
See [the docs][`inet.af/netaddr`] for details.

### Usage

```go
net.ParseIPPrefix ipprefix
```
```go
ipprefix | net.ParseIPPrefix
```

### Arguments

| name | description |
|------|-------------|
| `ipprefix` | _(required)_ The IP address prefix to parse. It must represent either an IPv4 or IPv6 prefix, containing a `/`. |

### Examples

```console
$ gomplate -i '{{ (net.ParseIPPrefix "192.168.0.0/24").Range }}'
192.168.0.0-192.168.0.255
$ gomplate -i '{{ $ip := net.ParseIP (net.LookupIP "example.com") -}}
  {{ $net := net.ParseIPPrefix "93.184.0.0/16" -}}
  {{ $net.Contains $ip }}'
true
$ gomplate -i '{{ $net := net.ParseIPPrefix "93.184.0.0/12" -}}
  {{ $net.Range }}'
93.176.0.0-93.191.255.255
```

## `net.ParseIPRange`

Parse the given string as an inclusive range of IP addresses from the same
address family (a `netaddr.IPRange` from the [`inet.af/netaddr`][] package).

The string must contain a hyphen (`-`).

Any of `netaddr.IPRange`'s methods may be called on the resulting value.
See [the docs][`inet.af/netaddr`] for details.

### Usage

```go
net.ParseIPRange iprange
```
```go
iprange | net.ParseIPRange
```

### Arguments

| name | description |
|------|-------------|
| `iprange` | _(required)_ The IP address range to parse. It must represent either an IPv4 or IPv6 range, containing a `-`. |

### Examples

```console
$ gomplate -i '{{ (net.ParseIPRange "192.168.0.0-192.168.0.255").To }}'
192.168.0.255
$ gomplate -i '{{ $range := net.ParseIPRange "1.2.3.0-1.2.3.233" -}}
  {{ $range.Prefixes }}'
[1.2.3.0/25 1.2.3.128/26 1.2.3.192/27]
```
