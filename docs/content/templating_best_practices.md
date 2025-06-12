---
title: Best Practices in Templating
weight: 11
menu: main
---

This guide provides practical recommendations for writing clear, reusable, and scalable templates in Gomplate. While many examples come from Helm values generation in a microservices environment, the concepts apply to a broad range of templating use cases.

### Using Environment Variables Effectively

You can and should create your own context using environment variables -- but choose them wisely. They should be overarching umbrellas that templates can share.

**Good candidates for context variables:**

- Cluster information
- Environment information
- Service identifiers
- Shared config paths
- Terraform outputs

In the below example, we declare a custom environment variable in a global template (a [configuration](https://docs.gomplate.ca/config/) file will also do the trick)
and set up the context:

```gotemplate
{{- $env := env.Getenv "ENVIRONMENT" -}}
{{- defineDatasource "env_config" (printf "%s/envs/%s/values.yaml" $PREFIX $env) -}}

{{- $ctx := dict 
  "ENVIRONMENT" (ds "env_config")
-}}

{{- tmpl.Exec "service" $ctx -}}
```

Now in your templates, you can reference these values cleanly:

```gotemplate
service:
  environment: {{ .ENVIRONMENT.name }}
  config:
    host: {{ .ENVIRONMENT.dependency.host }}
    port: {{ .ENVIRONMENT.dependency.port }}
```

### Managing Config Divergence

When dealing with multiple groups of configs that behave differently from one another, you might be tempted to conditional your way out of it:

```gotemplate
db:
  {{- if or (eq .CLUSTER.name "dev1") (eq .CLUSTER.name "int2") (eq .CLUSTER.name "gov") }}
  auth: true
  {{- else }}
  auth: false
  {{- end }}
  name: {{ index .TF_OUTPUTS_ALL_SERVICES "cloud_provisioner" "db-name" "value" }}
  region: {{ .CLUSTER.region }}
  {{- if or (eq .CLUSTER.name "dev1") (eq .CLUSTER.name "int2") (eq .CLUSTER.name "gov") }}
  user: cloudprovisioner_iam
  {{- else }}
  user: {{ index .TF_OUTPUTS_ALL_SERVICES "cloud_provisioner" "db-user-name" "value" }}
  {{- end }}
```

This approach is repetitive, error-prone, and difficult to maintain. Instead:

1. Check if the config exists in environment variables or external [datasources](https://docs.gomplate.ca/datasources/) like Consul.
2. If not, create a dedicated config file (which can itself be a template!)

The below example demonstrates this idea. Note that while JSON is shown here, YAML or CSV formats could be used depending on your specific needs.

**config.json:**

```json
{
  "db_auth": {
    "dev1": true,
    "int2": true,
    "gov": true,
    "prod": false
  },
  "db_users": {
    "dev1": "cloudprovisioner_iam",
    "int2": "cloudprovisioner_iam",
    "gov": "cloudprovisioner_iam",
    "prod": "{{ index .TF_OUTPUTS_ALL_SERVICES "cloud_provisioner" "db-user-name" "value" }}"
  }
}
```

**values.gtmpl:**

```gotemplate
{{- defineDatasource "config" "/path/to/config/file" -}}
{{- $config := ds "config" }}
db:
  auth: {{ index $config.db_auth .CLUSTER.name }}
  user: {{ tpl (index $config.db_users .CLUSTER.name) . }}
  name: {{ index .TF_OUTPUTS_ALL_SERVICES "cloud_provisioner" "db-name" "value" }}
  region: {{ .CLUSTER.region }}
```

### Working with Conditionals

Inline logic works well for simple booleans, but quickly becomes difficult to read and maintain.
Instead, use variables:

```gotemplate
replicas: {{- if eq .ENVIRONMENT.name "prod" }} 3 {{ else }} 1 {{ end }}
```

becomes:

```gotemplate
{{- $replicas := 1 -}}
{{- if eq .ENVIRONMENT.name "prod" }}{{- $replicas = 3 -}}{{- end }}
---
replicas: {{ $replicas }}
```

For more complex conditionals that you want to reuse across templates, consider using a datasource with the logic pre-computed:

**config.yaml:**

```yaml
replicas:
  dev: 1
  staging: 2
  prod: 3
```

**values.gtmpl:**

```gotemplate
{{- defineDatasource "config" "config.yaml" -}}
{{- $config := ds "config" -}}
replicas: {{ index $config.replicas .ENVIRONMENT.name }}
```

### Reducing Path Repetition

When working with nested configurations, you might find yourself repeating the same path multiple times:

```gotemplate

envVars:
  - name: REGION
    value: {{ .CLUSTER.type.region }}
  - name: PARTITION
    value: {{ .CLUSTER.type.partition }}
  - name: SERVICEDEPENDENCY_HOST
    value: {{ .CLUSTER.type.dep.host }}
  - name: SERVICEDEPENDENCY_PORT
    value: {{ .CLUSTER.type.dep.port }}

```

Using `with` blocks cuts repetition, makes structure easier to read, and reduces chances of errors:

```gotemplate

{{- with .CLUSTER.type -}}
envVars:
  - name: REGION
    value: {{ .region }}
  - name: PARTITION
    value: {{ .partition }}
  - name: SERVICEDEPENDENCY_HOST
    value: {{ .dep.host }}
  - name: SERVICEDEPENDENCY_PORT
    value: {{ .dep.port }}
{{- end -}}

```

### Closing Thoughts

This guide covered several key practices for writing maintainable Gomplate templates:

- Use environment variables for high-level, shared configuration
- Extract complex configuration into datasources instead of embedding conditionals
- Keep conditionals clean by using variables or moving logic to datasources
- Use `with` blocks to reduce repetitive path references

Remember that templates should be:

- Easy to read and maintain
- Modular and reusable
- Well-structured with configuration separate from logic
- Consistent in their approach to similar problems

While these patterns emerged from a Helm values use case, they apply to any complex templating needs with Gomplate.

For more advanced usage, check out:
- [Datasource docs](https://docs.gomplate.ca/datasources/) for additional data formats and sources
- [Functions reference](https://docs.gomplate.ca/functions/) for the full range of templating capabilities
- [Configuration guide](https://docs.gomplate.ca/config/) for template reuse patterns