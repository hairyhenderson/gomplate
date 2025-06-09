---
title: Best Practices in Templating
weight: 11
menu: main
---

this guide is based on my experience migrating helm values generation of our 100+ microservices from consul-template to gomplate. while some examples come from this context, the principles apply broadly to writing maintainable, reusable templates in gomplate.

### using environment variables effectively

you can and should create your own environment variables.
for example, we had cluster, environment, service, and terraform outputs as environment variables that acted as canonical sources for configuration values. in a multi-repo structure, however, these values may need to be managed via external systems using a [datasource](https://docs.gomplate.ca/datasources/) like vault or consul.

limit this to the overarching umbrellas that configs fall under, such that as many templates as possible can benefit from it, rather than scoping it to be used for individual templates.

in the below example, we declare a custom environment variable, then allow ourselves to reference it.

in your gomplate configuration:

```gotemplate
{{ $ENVIRONMENT := env.Getenv "ENVIRONMENT" }}
{{- printf "%s/envs/%s/values.gotemplate" $PREFIX $ENVIRONMENT | defineDatasource "ENVIRONMENT_DS"  -}}

{{- $TEMPLATE_VARS := coll.Dict "ENVIRONMENT" $ENVIRONMENT -}}

{{- $ctx := coll.Dict "GLOBAL_VARS" $TEMPLATE_VARS
                      "ENVIRONMENT" (ds "ENVIRONMENT_DS")  -}}

{{- tmpl.Exec "template" $ctx -}}

```

in your template:

```gotemplate
---
env:
  name: {{ .ENVIRONMENT.name }}
  dependency:
    host: {{ .ENVIRONMENT.dependency.host }}
    port: {{ .ENVIRONMENT.dependency.port }}
```

### managing config divergence

let's say you're in a situation where multiple groups of configs behave differently from one another, unpredictably. you could, of course, try to conditional your way out of it:

```gotemplate
db:
  {{- if or (eq .CLUSTER.name "megadev") (eq .CLUSTER.name "dev-gov") (eq .CLUSTER.name "int-gov") }}
  auth: true
  {{- else }}
  auth: false
  {{- end }}
  name: {{ index .TF_OUTPUTS_ALL_SERVICES "cloud_provisioner" "db-name" "value" }}
  region: {{ .CLUSTER.region }}
  {{- if or (eq .CLUSTER.name "megadev") (eq .CLUSTER.name "gov-dev") (eq .CLUSTER.name "int-gov") }}
  user: cloudprovisioner_iam
  {{- else }}
  user: {{ index .TF_OUTPUTS_ALL_SERVICES "cloud_provisioner" "db-user-name" "value" }}
  {{- end }}
```

this is repetitive and error-prone. it is also difficult to understand.

firstly, verify the config in question cannot be referenced another way via a custom environment variable, like a terraform output. if not, you can, instead, create a configuration file json and datasource it like so (it can even be a [template](http://docs.gomplate.ca/functions/tmpl/) itself!). json is a favorite because it allows one to untangle values rather cleanly.

**config.json:**

```json
{
  "cluster1": {
    "replica_counts": 3,
    "oidc_issuer_format": "{{ .CLUSTER.type.user_management.redirectPath }}%s",
    "other_thing": "a_unique_string"
  }
}
```

**values.gtmpl:**

```gotemplate
{{- defineDatasource "config" "/path/to/config/file" -}}
{{- $config := ds "config" }}
envVars:
  - name: OIDC_ISSUER_FORMAT
    VALUE: {{ tpl (index $config.oidc_issuer_format .CLUSTER.name) }}
```

### working with conditionals

inline logic is great in boolean settings, but quickly becomes difficult to read and write.
 instead, use variables to reference it. in this example, the variable is defined before the gotemplate start, but it can easily be defined in another .gtmpl and used as a datasource. ** note: fact check/ex.

 ```gotemplate

replicas: {{- if eq .ENVIRONMENT.name "prod" }} 3 {{ else }} 1 {{ end }}


```

becomes:

```gotemplate

{{- $replicas := 1 -}}
{{- if eq .ENVIRONMENT.name "prod" }}{{- $replicas = 3 -}}{{- end }}
replicas: {{ $replicas }}


```

### saving your fingers (avoiding repetitive typing)

a common idea present in engineering is to group similar configurations, both in their source and their target locations. it is not uncommon to then find oneself with a block of highly-nested configuration that needs calling.

an inexperienced gomplater would iterate down the configs. however, an issue arises when the configs' home must change - it will be cumbersome to update the template. it is also more difficult to read.

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

becomes

```gotemplate

{{- with .CLUSTER.type -}}
envVars:
  - name: REGION
    value: {{ .crnRegion }}
  - name: PARTITION
    value: {{ .crnPartition }}
  - name: SERVICEDEPENDENCY_HOST
    value: {{ .dep.host }}
  - name: SERVICEDEPENDENCY_PORT
    value: {{ .dep.port }}
{{- end -}}

```

### minimizing repetition

so your template is as good as it's going to get, but what about the next one? imagine that every single one of your microservices requires the above snippet. a ```with``` statement might make it easier to read once, but do you want to write it out n+1 times? use functions!

```gotemplate
example where several templates require the same block of configs
use a function to tpl the configs
```

### closing thoughts

not exhaustive. provide tldr