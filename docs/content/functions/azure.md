---
title: azure functions
menu:
  main:
    parent: functions
---

The functions in the `azure` namespace interface with various Microsoft Azure
APIs to make it possible for a template to render differently based on the Azure 
environment and metadata.

### Configuring Azure

A number of environment variables can be used to control how gomplate communicates
with Azure APIs.

| Environment Variable | Description |
| -------------------- | ----------- |
| `AZURE_META_ENDPOINT` | _(Default `http://169.254.169.254`)_ Sets the base address of the instance metadata service. |
| `AZURE_TIMEOUT` | _(Default `500`)_ Adjusts timeout for API requests, in milliseconds. |

## `azure.Meta`

Queries [Azure Instance Metadata Service](https://learn.microsoft.com/en-us/azure/virtual-machines/instance-metadata-service) for information.

For times when running outside Azure, or when the metadata API can't be reached, a `default` value can be provided.

### Usage

```go
azure.Meta key [format] [apiVersion] [default]
```

### Arguments

| name | description |
|------|-------------|
| `key` | _(required)_ the metadata key to query. To receive all available values, use an empty key with format json. |
| `format` | _(optional)_ the format of the metadata to query. Allowed values: `text`, `json`. Use `json` to receive the full tree of properties. |
| `apiVersion` | _(optional)_ Specify the api version to use to query the IMDS. Defaults to 2021-12-13 |
| `default` | _(optional)_ the default value |

### Examples

```console
$ echo '{{ azure.Meta "compute/resourceId" }}' | gomplate
/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/name/providers/Microsoft.Compute/virtualMachines/name
```
```console
$ echo '{{ azure.Meta "network/interface/0/ipv4/ipAddress/0/privateIpAddress" }}' | gomplate
10.0.192.5
```
```console
$ echo '{{ azure.Meta "compute/tagsList" "json" | jsonArray | jsonpath `$[?(@.name=="owner")].value` }}' | gomplate
me
```
```console
$ echo '{{ azure.Meta "compute/vmId" "text"  }}' | gomplate
/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/name/providers/Microsoft.Compute/virtualMachines/name
```
