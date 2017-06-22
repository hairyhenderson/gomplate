---
title: aws functions
menu:
  main:
    parent: functions
---

The functions in the `aws` namespace interface with various Amazon Web Services
APIs to make it possible for a template to render differently based on the AWS
environment and metadata.

#### `AWS_TIMEOUT` variable

In some cases AWS APIs may be slower to respond. The timeout for these requests
can be adjusted by setting the `AWS_TIMEOUT` environment variable. The value
must be in milliseconds, and defaults to 500 milliseconds.

## `aws.EC2Meta`

**Alias:** _(to be deprecated)_ `ec2meta`

Queries AWS [EC2 Instance Metadata](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html) for information. This only retrieves data in the `meta-data` path -- for data in the `dynamic` path use `aws.EC2Dynamic`.

This only works when running `gomplate` on an EC2 instance. If the EC2 instance metadata API isn't available, the tool will timeout and fail.

#### Example

```console
$ echo '{{aws.EC2Meta "instance-id"}}' | gomplate
i-12345678
```

## `aws.EC2Dynamic`

**Alias:** _(to be deprecated)_ `ec2dynamic`

Queries AWS [EC2 Instance Dynamic Metadata](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html) for information. This only retrieves data in the `dynamic` path -- for data in the `meta-data` path use `aws.EC2Meta`.

This only works when running `gomplate` on an EC2 instance. If the EC2 instance metadata API isn't available, the tool will timeout and fail.

#### Example

```console
$ echo '{{ (aws.EC2Dynamic "instance-identity/document" | json).region }}' | ./gomplate
us-east-1
```

## `aws.EC2Region`

**Alias:** _(to be deprecated)_ `ec2region`

Queries AWS to get the region. An optional default can be provided, or returns
`unknown` if it can't be determined for some reason.

#### Example

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

**Alias:** _(to be deprecated)_ `ec2tag`

Queries the AWS EC2 API to find the value of the given [user-defined tag](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/Using_Tags.html). An optional default
can be provided.

#### Example

```console
$ echo 'This server is in the {{ aws.EC2Tag "Account" }} account.' | ./gomplate
foo
$ echo 'I am a {{ aws.EC2Tag "classification" "meat popsicle" }}.' | ./gomplate
I am a meat popsicle.
```