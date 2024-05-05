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
[`github.com/inetaf/netaddr`]: https://pkg.go.dev/github.com/inetaf/netaddr
[`net`]: https://pkg.go.dev/net

## `net.LookupIP`

Resolve an IPv4 address for a given host name. When multiple IP addresses
are resolved, the first one is returned.

_Added in gomplate [v1.9.0](https://github.com/hairyhenderson/gomplate/releases/tag/v1.9.0)_
### Usage

```
net.LookupIP name
```
```
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

_Added in gomplate [v1.9.0](https://github.com/hairyhenderson/gomplate/releases/tag/v1.9.0)_
### Usage

```
net.LookupIPs name
```
```
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

_Added in gomplate [v1.9.0](https://github.com/hairyhenderson/gomplate/releases/tag/v1.9.0)_
### Usage

```
net.LookupCNAME name
```
```
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

_Added in gomplate [v1.9.0](https://github.com/hairyhenderson/gomplate/releases/tag/v1.9.0)_
### Usage

```
net.LookupSRV name
```
```
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

_Added in gomplate [v1.9.0](https://github.com/hairyhenderson/gomplate/releases/tag/v1.9.0)_
### Usage

```
net.LookupSRVs name
```
```
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

_Added in gomplate [v1.9.0](https://github.com/hairyhenderson/gomplate/releases/tag/v1.9.0)_
### Usage

```
net.LookupTXT name
```
```
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

## `net.ParseAddr`_(unreleased)_
**Unreleased:** _This function is in development, and not yet available in released builds of gomplate._

Parse the given string as an IP address (a
[`netip.Addr`](https://pkg.go.dev/net/netip#Addr)).

Any of `netip.Addr`'s methods may be called on the resulting value. See
[the docs](https://pkg.go.dev/net/netip#Addr) for details.

### Usage

```
net.ParseAddr addr
```
```
addr | net.ParseAddr
```

### Arguments

| name | description |
|------|-------------|
| `addr` | _(required)_ The IP string to parse. It must be either an IPv4 or IPv6 address. |

### Examples

```console
$ gomplate -i '{{ (net.ParseAddr "192.168.0.1").IsPrivate }}'
true
$ gomplate -i '{{ $ip := net.ParseAddr (net.LookupIP "example.com") -}}
  {{ $ip.Prefix 12 }}'
93.176.0.0/12
```

## `net.ParseIP` _(deprecated)_
**Deprecation Notice:** Use [`net.ParseAddr`](#net-parseaddr) instead.

Parse the given string as an IP address (a `netaddr.IP` from the
[`github.com/inetaf/netaddr`](https://pkg.go.dev/github.com/inetaf/netaddr) package).

Any of `netaddr.IP`'s methods may be called on the resulting value. See
[the docs](https://pkg.go.dev/github.com/inetaf/netaddr) for details.

_Added in gomplate [v3.10.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.10.0)_
### Usage

```
net.ParseIP ip
```
```
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

## `net.ParsePrefix`_(unreleased)_
**Unreleased:** _This function is in development, and not yet available in released builds of gomplate._

Parse the given string as an IP address prefix (CIDR) representing an IP
network (a [`netip.Prefix`](https://pkg.go.dev/net/netip#Prefix)).

The string can be in the form `"192.168.1.0/24"` or `"2001::db8::/32"`,
the CIDR notations defined in [RFC 4632][] and [RFC 4291][].

Any of `netip.Prefix`'s methods may be called on the resulting value. See
[the docs](https://pkg.go.dev/net/netip#Prefix) for details.

### Usage

```
net.ParsePrefix prefix
```
```
prefix | net.ParsePrefix
```

### Arguments

| name | description |
|------|-------------|
| `prefix` | _(required)_ The IP address prefix to parse. It must represent either an IPv4 or IPv6 prefix, containing a `/`. |

### Examples

```console
$ gomplate -i '{{ (net.ParsePrefix "192.168.0.0/24").Range }}'
192.168.0.0-192.168.0.255
$ gomplate -i '{{ $ip := net.ParseAddr (net.LookupIP "example.com") -}}
  {{ $net := net.ParsePrefix "93.184.0.0/16" -}}
  {{ $net.Contains $ip }}'
true
```

## `net.ParseIPPrefix` _(deprecated)_
**Deprecation Notice:** Use [`net.ParsePrefix`](#net-parseprefix) instead.

Parse the given string as an IP address prefix (CIDR) representing an IP
network (a `netaddr.IPPrefix` from the
[`github.com/inetaf/netaddr`][] package).

The string can be in the form `"192.168.1.0/24"` or `"2001::db8::/32"`,
the CIDR notations defined in [RFC 4632][] and [RFC 4291][].

Any of `netaddr.IPPrefix`'s methods may be called on the resulting value.
See [the docs][`github.com/inetaf/netaddr`] for details.

_Added in gomplate [v3.10.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.10.0)_
### Usage

```
net.ParseIPPrefix ipprefix
```
```
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

## `net.ParseRange`_(unreleased)_ _(experimental)_
**Unreleased:** _This function is in development, and not yet available in released builds of gomplate._
**Experimental:** This function is [_experimental_][experimental] and may be enabled with the [`--experimental`][experimental] flag.

[experimental]: ../config/#experimental

Parse the given string as an inclusive range of IP addresses from the same
address family (a [`netipx.IPRange`](https://pkg.go.dev/go4.org/netipx#IPRange)
from the [`go4.org/netipx`](https://pkg.go.dev/go4.org/netipx) module).

The string must contain a hyphen (`-`).

Any of `netipx.IPRange`'s methods may be called on the resulting value.
See [the docs](https://pkg.go.dev/go4.org/netipx#IPRange) for details.

Note that this function is experimental for now, because it uses a
[third-party module](https://pkg.go.dev/go4.org/netipx) which may be
brought into the standard library in the future, which may require
breaking changes to this function.

### Usage

```
net.ParseRange iprange
```
```
iprange | net.ParseRange
```

### Arguments

| name | description |
|------|-------------|
| `iprange` | _(required)_ The IP address range to parse. It must represent either an IPv4 or IPv6 range, containing a `-`. |

### Examples

```console
$ gomplate -i '{{ (net.ParseRange "192.168.0.0-192.168.0.255").To }}'
192.168.0.255
$ gomplate -i '{{ $range := net.ParseRange "1.2.3.0-1.2.3.233" -}}
  {{ $range.Prefixes }}'
[1.2.3.0/25 1.2.3.128/26 1.2.3.192/27 1.2.3.224/29 1.2.3.232/31]
```

## `net.ParseIPRange` _(deprecated)_
**Deprecation Notice:** Use [`net.ParseRange`](#net-parserange) instead.

Parse the given string as an inclusive range of IP addresses from the same
address family (a `netaddr.IPRange` from the [`github.com/inetaf/netaddr`][] package).

The string must contain a hyphen (`-`).

Any of `netaddr.IPRange`'s methods may be called on the resulting value.
See [the docs][`github.com/inetaf/netaddr`] for details.

_Added in gomplate [v3.10.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.10.0)_
### Usage

```
net.ParseIPRange iprange
```
```
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

## `net.CIDRHost` _(experimental)_
**Experimental:** This function is [_experimental_][experimental] and may be enabled with the [`--experimental`][experimental] flag.

[experimental]: ../config/#experimental

Calculates a full host IP address for a given host number within a given IP network address prefix.

The IP network can be in the form `"192.168.1.0/24"` or `"2001::db8::/32"`,
the CIDR notations defined in [RFC 4632][] and [RFC 4291][].

Any of `netip.Addr`'s methods may be called on the resulting value. See
[the docs](https://pkg.go.dev/net/netip#Addr) for details.

_Added in gomplate [v3.11.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.11.0)_
### Usage

```
net.CIDRHost hostnum prefix
```
```
prefix | net.CIDRHost hostnum
```

### Arguments

| name | description |
|------|-------------|
| `hostnum` | _(required)_ Is a whole number that can be represented as a binary integer with no more than the number of digits remaining in the address after the given prefix. |
| `prefix` | _(required)_ Must be given in CIDR notation. It must represent either an IPv4 or IPv6 prefix, containing a `/`. String or [`net.IPNet`](https://pkg.go.dev/net#IPNet) object returned from `net.ParseIPPrefix` can by used. |

### Examples

```console
$ gomplate -i '{{ "10.12.127.0/20" | net.CIDRHost 268 }}'
10.12.113.12
```

## `net.CIDRNetmask` _(experimental)_
**Experimental:** This function is [_experimental_][experimental] and may be enabled with the [`--experimental`][experimental] flag.

[experimental]: ../config/#experimental

The result is a subnet address formatted in the conventional dotted-decimal IPv4 address syntax or hexadecimal IPv6 address syntax, as expected by some software.

Any of `netip.Addr`'s methods may be called on the resulting value. See
[the docs](https://pkg.go.dev/net/netip#Addr) for details.

_Added in gomplate [v3.11.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.11.0)_
### Usage

```
net.CIDRNetmask prefix
```
```
prefix | net.CIDRNetmask
```

### Arguments

| name | description |
|------|-------------|
| `prefix` | _(required)_ Must be given in CIDR notation. It must represent either an IPv4 or IPv6 prefix, containing a `/`. String or [`net.IPNet`](https://pkg.go.dev/net#IPNet) object returned from `net.ParseIPPrefix` can by used. |

### Examples

```console
$ gomplate -i '{{ net.CIDRNetmask "10.12.127.0/20" }}'
255.255.240.0
$ gomplate -i '{{ net.CIDRNetmask "fd00:fd12:3456:7890:00a2::/72" }}'
ffff:ffff:ffff:ffff:ff00::
```

## `net.CIDRSubnets` _(experimental)_
**Experimental:** This function is [_experimental_][experimental] and may be enabled with the [`--experimental`][experimental] flag.

[experimental]: ../config/#experimental

Calculates a subnet address within given IP network address prefix.

Any of `netip.Prefix`'s methods may be called on the resulting values. See
[the docs](https://pkg.go.dev/net/netip#Prefix) for details.

_Added in gomplate [v3.11.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.11.0)_
### Usage

```
net.CIDRSubnets newbits prefix
```
```
prefix | net.CIDRSubnets newbits
```

### Arguments

| name | description |
|------|-------------|
| `newbits` | _(required)_ Is the number of additional bits with which to extend the prefix. For example, if given a prefix ending in `/16` and a `newbits` value of `4`, the resulting subnet address will have length `/20`. |
| `prefix` | _(required)_ Must be given in CIDR notation. It must represent either an IPv4 or IPv6 prefix, containing a `/`. String or [`net.IPNet`](https://pkg.go.dev/net#IPNet) object returned from `net.ParseIPPrefix` can by used. |

### Examples

```console
$ gomplate -i '{{ index ("10.0.0.0/16" | net.CIDRSubnets 2) 1 }}'
10.0.64.0/18
$ gomplate -i '{{ net.CIDRSubnets 2 "10.0.0.0/16" -}}'
[10.0.0.0/18 10.0.64.0/18 10.0.128.0/18 10.0.192.0/18]
```

## `net.CIDRSubnetSizes` _(experimental)_
**Experimental:** This function is [_experimental_][experimental] and may be enabled with the [`--experimental`][experimental] flag.

[experimental]: ../config/#experimental

Calculates a sequence of consecutive IP address ranges within a particular CIDR prefix.

Any of `netip.Prefix`'s methods may be called on the resulting values. See
[the docs](https://pkg.go.dev/net/netip#Prefix) for details.

_Added in gomplate [v3.11.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.11.0)_
### Usage

```
net.CIDRSubnetSizes newbits... prefix
```
```
prefix | net.CIDRSubnetSizes newbits...
```

### Arguments

| name | description |
|------|-------------|
| `newbits...` | _(required)_ Numbers of additional network prefix bits for returned address range. |
| `prefix` | _(required)_ Must be given in CIDR notation. It must represent either an IPv4 or IPv6 prefix, containing a `/`. String or [`net.IPNet`](https://pkg.go.dev/net#IPNet) object returned from `net.ParseIPPrefix` can by used. |

### Examples

```console
$ gomplate -i '{{ net.CIDRSubnetSizes 4 4 8 4 "10.1.0.0/16" -}}'
[10.1.0.0/20 10.1.16.0/20 10.1.32.0/24 10.1.48.0/20]
```
