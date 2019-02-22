---
title: aws functions
menu:
  main:
    parent: functions
---

The functions in the `aws` namespace interface with various Amazon Web Services
APIs to make it possible for a template to render differently based on the AWS
environment and metadata.

### Configuring AWS

A number of environment variables can be used to control how gomplate communicates
with AWS APIs. A few are documented here for convenience. See [the `aws-sdk-go` documentation](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html)
for details.

| Environment Variable | Description |
| -------------------- | ----------- |
| `AWS_TIMEOUT` | _(Default `500`)_ Adjusts timeout for API requests, in milliseconds. Not part of the AWS SDK. |
| `AWS_PROFILE` | Profile name the SDK should use when loading shared config from the configuration files. If not provided `default` will be used as the profile name. |
| `AWS_REGION` | Specifies where to send requests. See [this list](https://docs.aws.amazon.com/general/latest/gr/rande.html). Note that the region must be set for AWS functions to work correctly, either through this variable, or a configuration profile. |

## `aws.EC2Meta`

**Alias:** `ec2meta`

Queries AWS [EC2 Instance Metadata](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html) for information. This only retrieves data in the `meta-data` path -- for data in the `dynamic` path use `aws.EC2Dynamic`.

For times when running outside EC2, or when the metadata API can't be reached, a `default` value can be provided.

### Usage

```go
aws.EC2Meta key [default]
```

### Arguments

| name | description |
|------|-------------|
| `key` | _(required)_ the metadata key to query |
| `default` | _(optional)_ the default value |

### Examples

```console
$ echo '{{aws.EC2Meta "instance-id"}}' | gomplate
i-12345678
```

## `aws.EC2Dynamic`

**Alias:** `ec2dynamic`

Queries AWS [EC2 Instance Dynamic Metadata](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html) for information. This only retrieves data in the `dynamic` path -- for data in the `meta-data` path use `aws.EC2Meta`.

For times when running outside EC2, or when the metadata API can't be reached, a `default` value can be provided.

### Usage

```go
aws.EC2Dynamic key [default]
```

### Arguments

| name | description |
|------|-------------|
| `key` | _(required)_ the dynamic metadata key to query |
| `default` | _(optional)_ the default value |

### Examples

```console
$ echo '{{ (aws.EC2Dynamic "instance-identity/document" | json).region }}' | gomplate
us-east-1
```

## `aws.EC2Region`

**Alias:** `ec2region`

Queries AWS to get the region. An optional default can be provided, or returns
`unknown` if it can't be determined for some reason.

### Usage

```go
aws.EC2Region [default]
```

### Arguments

| name | description |
|------|-------------|
| `default` | _(optional)_ the default value |

### Examples

_In EC2_
```console
$ echo '{{ aws.EC2Region }}' | ./gomplate
us-east-1
```
_Not in EC2_
```console
$ echo '{{ aws.EC2Region }}' | ./gomplate
unknown
$ echo '{{ aws.EC2Region "foo" }}' | ./gomplate
foo
```

## `aws.EC2Tag`

**Alias:** `ec2tag`

Queries the AWS EC2 API to find the value of the given [user-defined tag](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/Using_Tags.html). An optional default
can be provided.

### Usage

```go
aws.EC2Tag tag [default]
```

### Arguments

| name | description |
|------|-------------|
| `tag` | _(required)_ the tag to query |
| `default` | _(optional)_ the default value |

### Examples

```console
$ echo 'This server is in the {{ aws.EC2Tag "Account" }} account.' | ./gomplate
foo
```
```console
$ echo 'I am a {{ aws.EC2Tag "classification" "meat popsicle" }}.' | ./gomplate
I am a meat popsicle.
```
