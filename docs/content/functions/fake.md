---
title: fake functions
menu:
  main:
    parent: functions
---

Functions for generating realistic fake data, such as names, email
addresses, street addresses, job titles, internet values, and product names.
These are useful for test fixtures, load tests, demos, and sample
configuration.

The values come from [gofakeit](https://github.com/brianvoe/gofakeit). Most
are picked at random from built-in English-language data sets, and the
address data is US-centric. Every call returns a new value, but values are
not guaranteed to be unique: the same name or email can come up twice.
Each call is also independent of the others, so values don't match each
other: [`fake.City`](#fakecity) and [`fake.StateAbr`](#fakestateabr) can
name a city and a state that don't go together.

For random strings and numbers, see the [`random`](../random/) namespace.
For UUIDs, see [`uuid.V4`](../uuid/#uuidv4), and for MAC addresses see
[`net.GenerateMAC`](../net/#netgeneratemac).

| Group | Functions |
|-------|-----------|
| Person | [`fake.Name`](#fakename), [`fake.FirstName`](#fakefirstname), [`fake.MiddleName`](#fakemiddlename), [`fake.LastName`](#fakelastname), [`fake.NamePrefix`](#fakenameprefix), [`fake.NameSuffix`](#fakenamesuffix), [`fake.Hobby`](#fakehobby) |
| Contact | [`fake.Email`](#fakeemail), [`fake.Phone`](#fakephone), [`fake.PhoneFormatted`](#fakephoneformatted) |
| Address | [`fake.Street`](#fakestreet), [`fake.City`](#fakecity), [`fake.State`](#fakestate), [`fake.StateAbr`](#fakestateabr), [`fake.Zip`](#fakezip), [`fake.Country`](#fakecountry), [`fake.CountryAbr`](#fakecountryabr), [`fake.TimeZoneRegion`](#faketimezoneregion) |
| Company and job | [`fake.CompanySuffix`](#fakecompanysuffix), [`fake.JobTitle`](#fakejobtitle), [`fake.BuzzWord`](#fakebuzzword), [`fake.BS`](#fakebs), [`fake.Slogan`](#fakeslogan) |
| Internet | [`fake.DomainName`](#fakedomainname), [`fake.DomainSuffix`](#fakedomainsuffix), [`fake.URL`](#fakeurl), [`fake.IPv4Address`](#fakeipv4address), [`fake.IPv6Address`](#fakeipv6address), [`fake.HTTPMethod`](#fakehttpmethod), [`fake.UserAgent`](#fakeuseragent) |
| Product and text | [`fake.ProductName`](#fakeproductname), [`fake.ProductCategory`](#fakeproductcategory), [`fake.ProductDescription`](#fakeproductdescription), [`fake.Word`](#fakeword), [`fake.Phrase`](#fakephrase) |
| Reproducible output | [`fake.Seeded`](#fakeseeded) |

### About randomness

Values come from a [PCG](https://pkg.go.dev/math/rand/v2#PCG)
pseudo-random number generator that is seeded from the operating system's
cryptographic random number generator, so the output is different every
time gomplate runs. For output that is the same on every run, use
[`fake.Seeded`](#fakeseeded). Like the [`random`](../random/) functions,
these functions are not suitable for security-sensitive uses, such as
generating passwords or keys.

## `fake.Name`

Generates a person's full name: a first name and a last name separated
by a space.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.Name
```


### Examples

```console
$ gomplate -i '{{ fake.Name }}'
Benton Stevens
```

## `fake.FirstName`

Generates a person's first name.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.FirstName
```


### Examples

```console
$ gomplate -i '{{ fake.FirstName }}'
Katrina
```

## `fake.MiddleName`

Generates a person's middle name.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.MiddleName
```


### Examples

```console
$ gomplate -i '{{ fake.MiddleName }}'
Anastasia
```

## `fake.LastName`

Generates a person's last name.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.LastName
```


### Examples

```console
$ gomplate -i '{{ fake.LastName }}'
Hicks
```

## `fake.NamePrefix`

Generates a name prefix or title, such as `Mr.`, `Ms.`, or `Dr.`.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.NamePrefix
```


### Examples

```console
$ gomplate -i '{{ fake.NamePrefix }} {{ fake.LastName }}'
Ms. Russell
```

## `fake.NameSuffix`

Generates a name suffix, such as `Jr.`, `III`, or `MD`.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.NameSuffix
```


### Examples

```console
$ gomplate -i '{{ fake.Name }}, {{ fake.NameSuffix }}'
Tina Crawford, II
```

## `fake.Hobby`

Generates the name of a hobby.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.Hobby
```


### Examples

```console
$ gomplate -i '{{ fake.Hobby }}'
Knitting
```

## `fake.Email`

Generates an email address, built from a random first and last name.
The address is well-formed, but its domain is made up and may or may
not exist, so don't send mail to it.

The address isn't related to the output of any other function, such as
[`fake.Name`](#fakename). To build an address that matches a name, use
the reserved `example.com` domain, as in the second example.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.Email
```


### Examples

```console
$ gomplate -i '{{ fake.Email }}'
emilebird@flatley.io
```
```console
$ gomplate -i '{{ $first := fake.FirstName }}{{ $last := fake.LastName }}{{ $first }} {{ $last }} <{{ print $first "." $last "@example.com" | strings.ToLower }}>'
Nichole Gottlieb <nichole.gottlieb@example.com>
```

## `fake.Phone`

Generates a 10-digit phone number with no formatting.

See [`fake.PhoneFormatted`](#fakephoneformatted) for a formatted number.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.Phone
```


### Examples

```console
$ gomplate -i '{{ fake.Phone }}'
7488179788
```

## `fake.PhoneFormatted`

Generates a North American-style phone number in one of several common
formats, such as `555-555-5555`, `(555)555-5555`, `1-555-555-5555`, or
`555.555.5555`.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.PhoneFormatted
```


### Examples

```console
$ gomplate -i '{{ fake.PhoneFormatted }}'
(056)965-7276
```

## `fake.Street`

Generates a street address line: a house number followed by a street
name.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.Street
```


### Examples

```console
$ gomplate -i '{{ fake.Street }}'
34737 West Wellsport
```

## `fake.City`

Generates the name of a US city.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.City
```


### Examples

```console
$ gomplate -i '{{ fake.City }}'
Fremont
```

## `fake.State`

Generates the full name of a US state.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.State
```


### Examples

```console
$ gomplate -i '{{ fake.State }}'
Washington
```

## `fake.StateAbr`

Generates the two-letter postal abbreviation of a US state.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.StateAbr
```


### Examples

```console
$ gomplate -i '{{ fake.StateAbr }}'
FL
```

## `fake.Zip`

Generates a five-digit US ZIP code.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.Zip
```


### Examples

```console
$ gomplate -i '{{ fake.City }}, {{ fake.StateAbr }} {{ fake.Zip }}'
Santa Ana, NE 79443
```

## `fake.Country`

Generates the name of a country.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.Country
```


### Examples

```console
$ gomplate -i '{{ fake.Country }}'
Uruguay
```

## `fake.CountryAbr`

Generates a two-letter [ISO 3166-1 alpha-2](https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2)
country code.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.CountryAbr
```


### Examples

```console
$ gomplate -i '{{ fake.CountryAbr }}'
LC
```

## `fake.TimeZoneRegion`

Generates the name of a time zone from the
[IANA time zone database](https://www.iana.org/time-zones), such as
`America/Chicago`. These are the names that the `TZ` environment
variable and most software accept.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.TimeZoneRegion
```


### Examples

```console
$ gomplate -i '{{ fake.TimeZoneRegion }}'
America/Jujuy
```

## `fake.CompanySuffix`

Generates a company name suffix: `Inc`, `LLC`, `Group`, or `and Sons`.
Combine it with [`fake.LastName`](#fakelastname) for a fictional
company name.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.CompanySuffix
```


### Examples

```console
$ gomplate -i '{{ fake.LastName }} {{ fake.CompanySuffix }}'
Swanson and Sons
```

## `fake.JobTitle`

Generates a job title.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.JobTitle
```


### Examples

```console
$ gomplate -i '{{ fake.JobTitle }}'
Stock Clerk
```

## `fake.BuzzWord`

Generates a business buzzword.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.BuzzWord
```


### Examples

```console
$ gomplate -i '{{ fake.BuzzWord }}'
Reactive
```

## `fake.BS`

Generates a word of business jargon.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.BS
```


### Examples

```console
$ gomplate -i '{{ fake.BS }}'
magnetic
```

## `fake.Slogan`

Generates a short company slogan.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.Slogan
```


### Examples

```console
$ gomplate -i '{{ fake.Slogan }}'
Precision from the start.
```

## `fake.DomainName`

Generates a domain name, such as `example.com`. The domain is made up
and may or may not exist.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.DomainName
```


### Examples

```console
$ gomplate -i '{{ fake.DomainName }}'
customerexpedite.biz
```

## `fake.DomainSuffix`

Generates a top-level domain, without the leading dot, such as `com`
or `org`.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.DomainSuffix
```


### Examples

```console
$ gomplate -i 'api.{{ fake.LastName | strings.ToLower }}.{{ fake.DomainSuffix }}'
api.mckenzie.info
```

## `fake.URL`

Generates an `http` or `https` URL with a made-up domain name and a
path of one to four segments.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.URL
```


### Examples

```console
$ gomplate -i '{{ fake.URL }}'
https://www.executivenetworks.io/mesh
```

## `fake.IPv4Address`

Generates an IPv4 address in dotted-decimal form. Each octet is picked
at random, so the address can fall in any range, including private,
loopback, and reserved ones.

For an address within a particular network, build one with
[`random.Number`](../random/#randomnumber) instead, for example
`{{ printf "10.0.%d.%d" (random.Number 0 255) (random.Number 1 254) }}`.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.IPv4Address
```


### Examples

```console
$ gomplate -i '{{ fake.IPv4Address }}'
97.11.41.53
```

## `fake.IPv6Address`

Generates an IPv6 address, written as eight groups of hexadecimal digits
separated by colons. As with [`fake.IPv4Address`](#fakeipv4address), the
address can fall in any range.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.IPv6Address
```


### Examples

```console
$ gomplate -i '{{ fake.IPv6Address }}'
4eda:5f98:b628:a566:1e58:d6e4:b754:8990
```

## `fake.HTTPMethod`

Generates an HTTP request method, such as `GET`, `POST`, or `DELETE`.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.HTTPMethod
```


### Examples

```console
$ gomplate -i '{{ fake.HTTPMethod }}'
PUT
```

## `fake.UserAgent`

Generates a web browser `User-Agent` string, in the style of Chrome,
Firefox, Safari, or Opera. The browser versions are not current, so
don't rely on these strings for browser detection.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.UserAgent
```


### Examples

```console
$ gomplate -i '{{ fake.UserAgent }}'
Opera/10.68 (Macintosh; PPC Mac OS X 10_9_3; en-US) Presto/2.8.277 Version/10.00
```

## `fake.ProductName`

Generates a product name.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.ProductName
```


### Examples

```console
$ gomplate -i '{{ fake.ProductName }}'
Tactical Wall Outlet Glide
```

## `fake.ProductCategory`

Generates a product category, in lower case.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.ProductCategory
```


### Examples

```console
$ gomplate -i '{{ fake.ProductCategory }}'
video equipment
```

## `fake.ProductDescription`

Generates a short product description.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.ProductDescription
```


### Examples

```console
$ gomplate -i '{{ fake.ProductDescription }}'
Whether you're using it at outdoor activities or on the go, this scenic Softwood Dice offers ultra-fast and ensures flexibility. Crafted from synthetic, it's perfect for tech enthusiasts.
```

## `fake.Word`

Generates a single English word.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.Word
```


### Examples

```console
$ gomplate -i '{{ fake.Word }}'
up
```

## `fake.Phrase`

Generates a common English phrase or saying, such as `time will tell`.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.Phrase
```


### Examples

```console
$ gomplate -i '{{ fake.Phrase }}'
stop the car
```

## `fake.Seeded`

Returns a copy of the `fake` namespace whose output is determined by
`seed`, for fixtures and sample data that must be the same on every
run. Call the other `fake` functions on the result: the same seed
always produces the same sequence of values.

The seed can be any value. It is converted to a string first, so `42`
and `"42"` are the same seed.

To keep each record the same when others are added or removed, give
each record its own seed, such as its index or ID, instead of drawing
every record from one seeded namespace.

Seeded output is reproducible with the same version of gomplate, but
may change in a later release, for example when the underlying data
sets change.

_<span class="release-check" data-tag="v5.4.0">Added in gomplate v5.4.0</span>_
### Usage

```
fake.Seeded seed
```
```
seed | fake.Seeded
```

### Arguments

| name | description |
|------|-------------|
| `seed` | _(required)_ the value that the generated values are derived from |

### Examples

```console
$ gomplate -i '{{ $f := fake.Seeded 42 }}{{ $f.Name }}, {{ $f.JobTitle }}'
Jaron Kling, Optometrist
```
```console
$ gomplate -i '{{ range $id := coll.Slice "a1" "b2" "c3" }}{{ $f := fake.Seeded $id }}{{ $id }}: {{ $f.Name }}, {{ $f.City }}
{{ end }}'
a1: Alexie Muller, Mesa
b2: Astrid Shields, Hialeah
c3: Declan Lane, Detroit
```
