---
title: sockaddr functions
menu:
  main:
    parent: functions
---

This namespace wraps the [`github.com/hashicorp/go-sockaddr`](https://github.com/hashicorp/go-sockaddr)
package, which makes it easy to discover information about a system's network
interfaces.

These functions are _partly_ documented here for convenience, but the canonical
documentation is at https://godoc.org/github.com/hashicorp/go-sockaddr.

Aside from some convenience functions, the general method of working with these 
functions is through a _pipeline_. There are _source_ functions, which select
interfaces ([`IfAddr`](https://godoc.org/github.com/hashicorp/go-sockaddr#IfAddr)),
and there are functions to further filter, refine, and finally to select
the specific attributes you're interested in.

To demonstrate how this can be used, here's an example that lists all of the IPv4 addresses available on the system:

```console
$ gomplate -i '{{ range (sockaddr.GetAllInterfaces | sockaddr.Include "type" "ipv4") -}}
{{ . | sockaddr.Attr "address" }}
{{end}}'
127.0.0.1
10.0.0.8
132.79.79.79
```

## `sockaddr.GetAllInterfaces`

Iterates over all available network interfaces and finds all available IP
addresses on each interface and converts them to `sockaddr.IPAddrs`, and returning
the result as an array of `IfAddr`.

Should be piped through a further function to refine and extract attributes.

## `sockaddr.GetDefaultInterfaces`

Returns `IfAddrs` of the addresses attached to the default route.

Should be piped through a further function to refine and extract attributes.

## `sockaddr.GetPrivateInterfaces`

Returns an array of `IfAddr`s containing every IP that matches
[RFC 6890][], is attached to the interface with
the default route, and is a forwardable IP address.

**Note:** [RFC 6890][] is a more exhaustive version of [RFC 1918][]
because it spans IPv4 and IPv6, however it does permit the inclusion of likely
undesired addresses such as multicast, therefore our definition of a "private"
address also excludes non-forwardable IP addresses (as defined by the IETF).

Should be piped through a further function to refine and extract attributes.

## `sockaddr.GetPublicInterfaces`

Returns an array of `IfAddr`s that do not match [RFC 6890][],
are attached to the default route, and are forwardable.

Should be piped through a further function to refine and extract attributes.

## `sockaddr.Sort`

Returns an array of `IfAddr`s sorted based on the given selector. Multiple sort
clauses can be passed in as a comma-delimited list without whitespace.

### Usage

```
<array-of-IfAddrs> | sockaddr.Sort selector
```

### Selectors

The valid selectors are:

| selector | sorts by... |
|----------|-------------|
| `address` | the network address |
| `default` | whether or not the `IfAddr` has a default route |
| `name` | the interface name |
| `port` | the port, if included in the `IfAddr` |
| `size` | the size of the network mask, smaller mask (larger number of hosts per network) to largest (e.g. a /24 sorts before a /32) |
| `type` | the type of the `IfAddr`. Order is Unix, IPv4, then IPv6 |

Each of these selectors sort _ascending_, but a _descending_ sort may be chosen
by prefixing the selector with a `-` (e.g. `-address`). You may prefix with a `+`
to make explicit that the sort is ascending.

`IfAddr`s that are not comparable will be at the end of the list and in a
non-deterministic order.

### Examples

To sort first by interface name, then by address (descending):
```console
$ gomplate -i '{{ sockaddr.GetAllInterfaces | sockaddr.Sort "name,-address" }}'
```

## `sockaddr.Exclude`

Returns an array of `IfAddr`s filtered by interfaces that do not match the given
selector's value.

### Usage

```
<array-of-IfAddrs> | sockaddr.Exclude selector value
```

### Selectors

The valid selectors are:

| selector | excludes by... |
|----------|-------------|
| `address` | the network address |
| `flag` | the specified flags (see below) |
| `name` | the interface name |
| `network` | being part of the given IP network (in net/mask format) |
| `port` | the port, if included in the `IfAddr` |
| `rfc` | being included in networks defined by the given RFC. See [the source code](https://github.com/hashicorp/go-sockaddr/blob/master/rfc.go#L38) for a list of valid RFCs | 
| `size` | the size of the network mask, as number of bits (e.g. `"24"` for a /24) |
| `type` | the type of the `IfAddr`. `unix`, `ipv4`, or `ipv6` |

#### supported flags

These flags are supported by the `flag` selector:
`broadcast`, `down`, `forwardable`, `global unicast`, `interface-local multicast`,
`link-local multicast`, `link-local unicast`, `loopback`, `multicast`, `point-to-point`,
`unspecified`, `up`

### Examples

To exclude all IPv6 interfaces:
```console
$ gomplate -i '{{ sockaddr.GetAllInterfaces | sockaddr.Exclude "type" "ipv6" }}'
```

## `sockaddr.Include`

Returns an array of `IfAddr`s filtered by interfaces that match the given
selector's value.

This is the inverse of `sockaddr.Exclude`. See [`sockaddr.Exclude`](#sockaddr.Exclude) for details.

## `sockaddr.Attr`

Returns the named attribute as a string.

### Usage

```
<array-of-IfAddrs> | sockaddr.Attr selector
```

### Examples

```console
$ gomplate -i '{{ range (sockaddr.GetAllInterfaces | sockaddr.Include "type" "ipv4") }}{{ . | sockaddr.Attr "name" }} {{end}}'
lo0 en0
```

## `sockaddr.Join`

Selects the given attribute from each `IfAddr` in the source array, and joins
the results with the given separator.

### Usage

```
<array-of-IfAddrs> | sockaddr.Join selector separator
```

### Examples

```console
$ gomplate -i '{{ sockaddr.GetAllInterfaces | sockaddr.Join "name" "," }}'
lo0,lo0,lo0,en0,en0
```

## `sockaddr.Limit`

Returns a slice of `IfAddr`s based on the specified limit.

### Usage

```
<array-of-IfAddrs> | sockaddr.Limit limit
```

### Examples

```console
$ gomplate -i '{{ sockaddr.GetAllInterfaces | sockaddr.Limit 2 | sockaddr.Join "name" "|" }}'
lo0|lo0
```

## `sockaddr.Offset`

Returns a slice of `IfAddr`s based on the specified offset.

### Usage

```
<array-of-IfAddrs> | sockaddr.Offset offset
```

### Examples

```console
$ gomplate -i '{{ sockaddr.GetAllInterfaces | sockaddr.Limit 2 | sockaddr.Offset 1 | sockaddr.Attr "address" }}'
::1
```

## `sockaddr.Unique`

Creates a unique array of `IfAddr`s based on the matching selector. Assumes the input has
already been sorted.

### Usage

```
<array-of-IfAddrs> | sockaddr.Unique selector
```

### Examples

```console
$ gomplate -i '{{ sockaddr.GetAllInterfaces | sockaddr.Unique "name" | sockaddr.Join "name" ", " }}'
lo0, en0
```

## `sockaddr.Math`

Applies a math operation to each `IfAddr` in the input. Any failure will result in zero results.

See [the source code](https://github.com/hashicorp/go-sockaddr/blob/master/ifaddrs.go#L725)
for details.

### Usage

```
<array-of-IfAddrs> | sockaddr.Math operation value
```

### Examples

```console
$ gomplate -i '{{ sockaddr.GetAllInterfaces | sockaddr.Math "address" "+5" | sockaddr.Attr "address" }}'
127.0.0.6
```

## `sockaddr.GetPrivateIP`

Returns a string with a single IP address that is part of [RFC 6890][] and has a
default route. If the system can't determine its IP address or find an [RFC 6890][]
IP address, an empty string will be returned instead.

### Usage

```
sockaddr.GetPrivateIP
```

### Examples

```console
$ gomplate -i '{{ sockaddr.GetPrivateIP }}'
10.0.0.28
```

## `sockaddr.GetPrivateIPs`

Returns a space-separated string with all IP addresses that are part of [RFC 6890][]
(regardless of whether or not there is a default route, unlike `GetPublicIP`).
If the system can't find any [RFC 6890][] IP addresses, an empty string will be
returned instead.

### Usage

```
sockaddr.GetPrivateIPs
```

### Examples

```console
$ gomplate -i '{{ sockaddr.GetPrivateIPs }}'
10.0.0.28 192.168.0.1
```

## `sockaddr.GetPublicIP`

Returns a string with a single IP address that is NOT part of [RFC 6890][] and
has a default route. If the system can't determine its IP address or find a
non-[RFC 6890][] IP address, an empty string will be returned instead.

### Usage

```
sockaddr.GetPublicIP
```

### Examples

```console
$ gomplate -i '{{ sockaddr.GetPublicIP }}'
8.1.2.3
```

## `sockaddr.GetPublicIPs`

Returns a space-separated string with all IP addresses that are NOT part of
[RFC 6890][] (regardless of whether or not there is a default route, unlike
`GetPublicIP`). If the system can't find any non-[RFC 6890][] IP addresses, an
empty string will be returned instead.

### Usage

```
sockaddr.GetPublicIPs
```

### Examples

```console
$ gomplate -i '{{ sockaddr.GetPublicIPs }}'
8.1.2.3 8.2.3.4
```

## `sockaddr.GetInterfaceIP`

Returns a string with a single IP address sorted by the size of the network
(i.e. IP addresses with a smaller netmask, larger network size, are sorted first).

### Usage

```
sockaddr.GetInterfaceIP name
```

### Examples

```console
$ gomplate -i '{{ sockaddr.GetInterfaceIP "en0" }}'
10.0.0.28
```

## `sockaddr.GetInterfaceIPs`

Returns a string with all IPs, sorted by the size of the network (i.e. IP
addresses with a smaller netmask, larger network size, are sorted first), on a
named interface.

### Usage

```
sockaddr.GetInterfaceIPs name
```

### Examples

```console
$ gomplate -i '{{ sockaddr.GetInterfaceIPs "en0" }}'
10.0.0.28 fe80::1f9a:5582:4b41:bd18
```

[RFC 1918]: http://tools.ietf.org/html/rfc1918
[RFC 6890]: http://tools.ietf.org/html/rfc6890
