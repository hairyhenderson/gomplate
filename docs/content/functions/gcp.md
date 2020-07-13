---
title: gcp functions
menu:
  main:
    parent: functions
---

The functions in the `gcp` namespace interface with various Google Cloud Platform
APIs to make it possible for a template to render differently based on the GCP
environment and metadata.

### Configuring GCP

A number of environment variables can be used to control how gomplate communicates
with GCP APIs.

| Environment Variable | Description |
| -------------------- | ----------- |
| `GCP_META_ENDPOINT` | _(Default `http://metadata.google.internal`)_ Sets the base address of the instance metadata service. |
| `GCP_TIMEOUT` | _(Default `500`)_ Adjusts timeout for API requests, in milliseconds. |

## `gcp.Meta`

Queries GCP [Instance Metadata](https://cloud.google.com/compute/docs/storing-retrieving-metadata) for information.

For times when running outside GCP, or when the metadata API can't be reached, a `default` value can be provided.

### Usage

```go
gcp.Meta key [default]
```

### Arguments

| name | description |
|------|-------------|
| `key` | _(required)_ the metadata key to query |
| `default` | _(optional)_ the default value |

### Examples

```console
$ echo '{{gcp.Meta "id"}}' | gomplate
1334999446930701104
```
```console
$ echo '{{gcp.Meta "network-interfaces/0/ip"}}' | gomplate
10.128.0.23
```
