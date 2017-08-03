---
layout: "docs"
page_title: "Server Configuration"
sidebar_current: "docs-configuration"
description: |-
  Vault server configuration reference.
---

# Vault Configuration

Outside of development mode, Vault servers are configured using a file.
The format of this file is [HCL](https://github.com/hashicorp/hcl) or JSON.
An example configuration is shown below:

```javascript
storage "consul" {
  address = "127.0.0.1:8500"
  path    = "vault"
}

listener "tcp" {
  address     = "127.0.0.1:8200"
  tls_disable = 1
}

telemetry {
  statsite_address = "127.0.0.1:8125"
  disable_hostname = true
}
```

After the configuration is written, use the `-config` flag with `vault server`
to specify where the configuration is.

## Parameters

- `storage` <tt>([StorageBackend][storage-backend]: \<required\>)</tt> –
  Configures the storage backend where Vault data is stored. Please see the
  [storage backends documentation][storage-backend] for the full list of
  available storage backends. Running Vault in HA mode would require
  coordination semantics to be supported by the backend. If the storage backend
  supports HA coordination, HA backend options can also be specified in this
  parameter block. If not, a separate `ha_storage` parameter should be
  configured with a backend that supports HA, along with corresponding HA
  options.

- `ha_storage` <tt>([StorageBackend][storage-backend]: nil)</tt> – Configures
  the storage backend where Vault HA coordination will take place. This must be
  an HA-supporting backend. If not set, HA will be attempted on the backend
  given in the `storage` parameter. This parameter is not required if the
  storage backend supports HA coordination and if HA specific options are
  already specified with `storage` parameter.

- `cluster_name` `(string: <generated>)` – Specifies the identifier for the
  Vault cluster. If omitted, Vault will generate a value. When connecting to
  Vault Enterprise, this value will be used in the interface.

- `listener` <tt>([Listener][listener]: \<required\>)</tt> – Configures how
  Vault is listening for API requests.

- `cache_size` `(string: "32000")` – Specifies the size of the read cache used
  by the physical storage subsystem. The value is in number of entries, so the
  total cache size depends on the size of stored entries.

- `disable_cache` `(bool: false)` – Disables all caches within Vault, including
  the read cache used by the physical storage subsystem. This will very
  significantly impact performance.

- `disable_mlock` `(bool: false)` – Disables the server from executing the
  `mlock` syscall. `mlock` prevents memory from being swapped to disk. Disabling
  `mlock` is not recommended in production, but is fine for local development
  and testing.

    Disabling `mlock` is not recommended unless the systems running Vault only
    use encrypted swap or do not use swap at all. Vault only supports memory
    locking on UNIX-like systems that support the mlock() syscall (Linux, FreeBSD, etc).
    Non UNIX-like systems (e.g. Windows, NaCL, Android) lack the primitives to keep a
    process's entire memory address space from spilling to disk and is therefore
    automatically disabled on unsupported platforms.

    On Linux, to give the Vault executable the ability to use the `mlock`
    syscall without running the process as root, run:

    ```shell
    sudo setcap cap_ipc_lock=+ep $(readlink -f $(which vault))
    ```

- `plugin_directory` `(string: "")` – A directory from which plugins are
  allowed to be loaded. Vault must have permission to read files in this
  directory to successfully load plugins.

- `telemetry` <tt>([Telemetry][telemetry]: nil)</tt> – Specifies the telemetry
  reporting system.

- `default_lease_ttl` `(string: "32d")` – Specifies the default lease duration
  for tokens and secrets. This is specified using a label suffix like `"30s"` or
  `"1h"`. This value cannot be larger than `max_lease_ttl`.

- `max_lease_ttl` `(string: "32d")` – Specifies the maximum possible lease
  duration for tokens and secrets. This is specified using a label
  suffix like `"30s"` or `"1h"`.

- `ui` `(bool: false, Enterprise-only)` – Enables the built-in web UI, which is
  available on all listeners (address + port) at the `/ui` path. Browsers accessing
  the standard Vault API address will automatically redirect there. This can also
  be provided via the environment variable `VAULT_UI`.

[storage-backend]: /docs/configuration/storage/index.html
[listener]: /docs/configuration/listener/index.html
[telemetry]: /docs/configuration/telemetry.html
