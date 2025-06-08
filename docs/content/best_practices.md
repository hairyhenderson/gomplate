---
title: Best Practices in Templating
weight: 11
menu: main
---

this guide is based on my experience migrating helm values generation of our 100+ microservices from consul-template to gomplate. while some examples come from this context, the principles apply broadly to writing maintainable, reusable templates in gomplate.

### using environment variables effectively

you can create your own environment variables using .Env - use it wisely  
we had cluster, environment, service, and terraform outputs as environment variables that acted as "sources" of information as that is where the specific configurations lived. in turn, SERVICE, ENVIRONMENT, etc. had their own existing central location.
something like this: `{{- $ENVIRONMENT := env.Getenv "ENVIRONMENT" }}` in a central configuration file. in a multi-repo structure, however, these values may need to be managed via external systems using a [datasource](https://docs.gomplate.ca/datasources/) like vault or consul.

limit this to the overarching umbrellas that configs fall under such that as many templates as possible can benefit from it, rather than scoping it to be used for individual templates. in the below block, we use our "custom" 

```yaml
---
ingress:
  hosts:
  - '*.{{ .CLUSTER.type.domain }}'
env:
  crnPartition: {{ .CLUSTER.type.crnPartition }}
  crnRegion: {{ .CLUSTER.type.crnRegion }}
  envName: {{ .ENVIRONMENT.name }}
  user_management:
    host: {{ .CLUSTER.type.user_management.host }}
    port: {{ .CLUSTER.type.user_management.port }}
  ENV: {{ .ENVIRONMENT.name }}
  data:
    host: {{ .SERVICE.data_app.host }}
    port: {{ .SERVICE.data_app.port }}
  publicApiEndpoint: https://{{ .CLUSTER.type.apiEndpoint }}

```

### managing config divergence

let's say you're in a situation where multiple groups of configs behave differently from one another in a seemingly random way. you could, of course, try to conditional your way out of it:

```yaml
db:
  {{- if or (eq .CLUSTER.name "devs-dev") (eq .CLUSTER.name "dev-gov") (eq .CLUSTER.name "int-gov") }}
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

my recommendation: i would always check if it can be a terraform output. however, many config values are not easily produced by terraform. you can, instead, create a configuration file json and datasource it like so (it can even be a [template](http://docs.gomplate.ca/functions/tmpl/) itself!). i like json for this because you can separate values out rather cleanly.

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

```yaml
{{- defineDatasource "config" "/path/to/config/file" -}}
{{- $config := ds "config" }}
envVars:
  - name: OIDC_ISSUER_FORMAT
    VALUE: {{ tpl (index $config.oidc_issuer_format .CLUSTER.name) }}
```

### using logic

inline logic is great for a black-and-white situation, but quickly becomes difficult to read and write.
 instead, use variables to reference it. in this example, the variable is defined before the yaml start, but it can easily be defined in another .gtmpl and used as a datasource.

 ```yaml

replicas: {{- if eq .ENVIRONMENT.name "prod" }} 3 {{ else }} 1 {{ end }}


```

becomes:

```yaml

{{- $replicas := 1 -}}
{{- if eq .ENVIRONMENT.name "prod" }}{{- $replicas = 3 -}}{{- end }}
replicas: {{ $replicas }}


```

### my "central config" has me typing too much. what gives?

following the widely-known best practice to group similar configurations together and "put them where they belong" so to speak (i.e. if they are PART of something thent they go below a level of its parent) (whatever that's called),
 it would make sense that an inexperienced gomplater would iterate down the configs. this will become difficult to read after a few levels, and adjusting the template should the location of these configs change would be rather cumbersome.

```yaml

envVars:
  - name: REGION
    value: {{ .CLUSTER.type.crnRegion }}
  - name: PARTITION
    value: {{ .CLUSTER.type.crnPartition }}
  - name: SERVICEDEPENDENCY_HOST
    value: {{ .CLUSTER.type.dep.host }}
  - name: SERVICEDEPENDENCY_PORT
    value: {{ .CLUSTER.type.dep.port }}

```

becomes

```yaml

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

### okay, but now do i have to do it all over again on the next template?

no you do not! imagine, if you will, that every single one of your microservices has a region and a partition, and it lives two levels deep in a config (i.e. .CLUSTER.type.region/.partition as demonstrated).