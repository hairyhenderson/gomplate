#!/usr/bin/env bats

load helper

function setup () {
  mkdir -p /var/run/secrets/kubernetes.io/serviceaccount
  touch /var/run/secrets/kubernetes.io/serviceaccount/token
  cp /tmp/ca.crt  /var/run/secrets/kubernetes.io/serviceaccount/
  start_k8s_svc
  export KUBERNETES_SERVICE_PORT=8083 KUBERNETES_SERVICE_HOST=127.0.0.1

}

function teardown () {
  stop_k8s_svc
  rm -rf /var/run/secrets/kubernetes.io
}

@test "Testing k8s" {
  gomplate --datasource data=k8s://default/test -i '{{datasource "data" "test.property.1" "default"}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "foo" ]]
}
