---
title: net functions
menu:
  main:
    parent: functions
---

## `net.LookupIP`

Resolve an IPv4 address for a given host name. When multiple IP addresses
are resolved, the first one is returned.

**Note:** Unresolvable hostnames will result in an error, and `gomplate` will panic.

### Usage

```go
net.LookupIP name
```

### Arguments

| name   | description |
|--------|-------|
| `name` | The hostname to look up. This can be a simple hostname, or a fully-qualified domain name. |

### Examples

```console
$ gomplate -i '{{ net.LookupIP "example.com" }}'
93.184.216.34
```

## `net.LookupIPs`

Resolve all IPv4 addresses for a given host name. Returns an array of strings.

**Note:** Unresolvable hostnames will result in an error, and `gomplate` will panic.

### Usage

```go
net.LookupIPs name
```

### Arguments

| name   | description |
|--------|-------|
| `name` | The hostname to look up. This can be a simple hostname, or a fully-qualified domain name. |

### Examples

```console
$ gomplate -i '{{ join (net.LookupIPs "twitter.com") "," }}'  
104.244.42.65,104.244.42.193
```

## `net.LookupCNAME`

Resolve the canonical name for a given host name. This does a DNS lookup for the
`CNAME` record type. If no `CNAME` is present, a canonical form of the given name
is returned -- e.g. `net.LookupCNAME "localhost"` will return `"localhost."`.

**Note:** Unresolvable hostnames will result in an error, and `gomplate` will panic.

### Usage

```go
net.LookupCNAME name
```

### Arguments

| name   | description |
|--------|-------|
| `name` | The hostname to look up. This can be a simple hostname, or a fully-qualified domain name. |

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

**Note:** Unresolvable hostnames will result in an error, and `gomplate` will panic.

### Usage

```go
net.LookupSRV name
```

### Arguments

| name   | description |
|--------|-------|
| `name` | The service name to look up |

### Examples

```console
$ gomplate -i '{{ net.LookupSRV "_sip._udp.sip.voice.google.com" | toJSONPretty "  " }}
'
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

**Note:** Unresolvable hostnames will result in an error, and `gomplate` will panic.

### Usage

```go
net.LookupSRVs name
```

### Arguments

| name   | description |
|--------|-------|
| `name` | The service name to look up |

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

**Note:** Unresolvable hostnames will result in an error, and `gomplate` will panic.

### Usage

```go
net.LookupTXT name
```

### Arguments

| name   | description |
|--------|-------|
| `name` | The host name to look up |

### Examples

```console
$ gomplate -i '{{net.LookupTXT "example.com" | toJSONPretty "\t" }}'
[
	"$Id: example.com 4415 2015-08-24 20:12:23Z davids $",
	"v=spf1 -all"
]
```
