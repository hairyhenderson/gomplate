package kubernetes

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

var TestPodRaw = `{
  "apiVersion": "v1",
  "kind": "Pod",
  "metadata": {
    "creationTimestamp": "2024-01-03T05:04:33Z",
    "labels": {
      "name": "myapp"
    },
    "name": "myapp",
    "namespace": "default",
    "resourceVersion": "274103",
    "selfLink": "/api/v1/namespaces/default/pods/myapp",
    "uid": "e8330f3c-66ca-11e9-b6fa-0800271788ca"
  },
  "spec": {
    "containers": [
      {
        "image": "nginx",
        "imagePullPolicy": "Always",
        "name": "myapp",
        "ports": [
          {
            "containerPort": 1234,
            "protocol": "TCP"
          }
        ],
        "resources": {},
        "terminationMessagePath": "/dev/termination-log",
        "terminationMessagePolicy": "File",
        "volumeMounts": [
          {
            "mountPath": "/var/run/secrets/kubernetes.io/serviceaccount",
            "name": "default-token-nmshj",
            "readOnly": true
          }
        ]
      }
    ],
    "dnsPolicy": "ClusterFirst",
    "enableServiceLinks": true,
    "nodeName": "minikube",
    "priority": 0,
    "restartPolicy": "Always",
    "schedulerName": "default-scheduler",
    "securityContext": {},
    "serviceAccount": "default",
    "serviceAccountName": "default",
    "terminationGracePeriodSeconds": 30,
    "tolerations": [
      {
        "effect": "NoExecute",
        "key": "node.kubernetes.io/not-ready",
        "operator": "Exists",
        "tolerationSeconds": 300
      },
      {
        "effect": "NoExecute",
        "key": "node.kubernetes.io/unreachable",
        "operator": "Exists",
        "tolerationSeconds": 300
      }
    ],
    "volumes": [
      {
        "name": "default-token-nmshj",
        "secret": {
          "defaultMode": 420,
          "secretName": "default-token-nmshj"
        }
      }
    ]
  },
  "status": {
    "conditions": [
      {
        "lastProbeTime": null,
        "lastTransitionTime": "2024-01-03T05:04:33Z",
        "status": "True",
        "type": "Initialized"
      },
      {
        "lastProbeTime": null,
        "lastTransitionTime": "2024-01-03T05:04:33Z",
        "status": "True",
        "type": "Ready"
      },
      {
        "lastProbeTime": null,
        "lastTransitionTime": "2024-01-03T05:04:33Z",
        "status": "True",
        "type": "ContainersReady"
      },
      {
        "lastProbeTime": null,
        "lastTransitionTime": "2024-01-03T05:04:33Z",
        "status": "True",
        "type": "PodScheduled"
      }
    ],
    "containerStatuses": [
      {
        "containerID": "docker://92d7dc7a851453c2f1e75c4af42a9e72fea50127fede62dfbd5fbb6fb0481fcc",
        "image": "nginx:latest",
        "imageID": "docker-pullable://nginx@sha256:96fb261b66270b900ea5a2c17a26abbfabe95506e73c3a3c65869a6dbe83223a",
        "lastState": {
          "terminated": {
            "containerID": "docker://288fc0a2b98708d6a4661f59c54c4ae366c1acea642f000ba9615932dbff411f",
            "exitCode": 0,
            "finishedAt": "2024-01-03T05:04:33Z",
            "reason": "Completed",
            "startedAt": "2024-01-03T05:04:33Z"
          }
        },
        "name": "myapp",
        "ready": true,
        "restartCount": 3,
        "state": {
          "running": {
            "startedAt": "2024-01-03T05:04:33Z"
          }
        }
      }
    ],
    "hostIP": "10.0.2.15",
    "phase": "Running",
    "podIP": "172.17.0.2",
    "qosClass": "BestEffort",
    "startTime": "2024-01-03T05:04:33Z"
  }
}
`

var TestPodNeat = `{
  "apiVersion": "v1",
  "kind": "Pod",
  "metadata": {"labels":{"name":"myapp"},"name":"myapp","namespace":"default"},
  "spec": {
    "containers": [
      {
        "image": "nginx",
        "name": "myapp",
        "ports": [
          {
            "containerPort": 1234
          }
        ]
      }
    ],
    "priority": 0,
    "serviceAccountName": "default",
    "tolerations": [
      {
        "effect": "NoExecute",
        "key": "node.kubernetes.io/not-ready",
        "operator": "Exists",
        "tolerationSeconds": 300
      },
      {
        "effect": "NoExecute",
        "key": "node.kubernetes.io/unreachable",
        "operator": "Exists",
        "tolerationSeconds": 300
      }
    ]
  }
}
`

var TestPVJsonRaw = `{
  "apiVersion": "v1",
  "kind": "PersistentVolume",
  "metadata": {
    "annotations": {
      "hostPathProvisionerIdentity": "7de69121-4d7a-11e9-8684-0800271788ca",
      "pv.kubernetes.io/provisioned-by": "k8s.io/minikube-hostpath"
    },
    "creationTimestamp": "2019-03-23T14:52:51Z",
    "finalizers": ["kubernetes.io/pv-protection"],
    "name": "pvc-54fad2fe-4d7b-11e9-9172-0800271788ca",
    "resourceVersion": "186863",
    "selfLink": "/api/v1/persistentvolumes/pvc-54fad2fe-4d7b-11e9-9172-0800271788ca",
    "uid": "5527dbad-4d7b-11e9-9172-0800271788ca"
  },
  "spec": {
    "accessModes": ["ReadWriteOnce"],
    "capacity": {
      "storage": "2Gi"
    },
    "claimRef": {
      "apiVersion": "v1",
      "kind": "PersistentVolumeClaim",
      "name": "prom-prometheus-alertmanager",
      "namespace": "default",
      "resourceVersion": "860",
      "uid": "54fad2fe-4d7b-11e9-9172-0800271788ca"
    },
    "hostPath": {
      "path": "/tmp/hostpath-provisioner/pvc-54fad2fe-4d7b-11e9-9172-0800271788ca",
      "type": ""
    },
    "persistentVolumeReclaimPolicy": "Delete",
    "storageClassName": "standard",
    "volumeMode": "Filesystem"
  },
  "status": {
    "phase": "Released"
  }
}`

var TestPVYAMLRaw = `apiVersion: v1
kind: PersistentVolume
metadata:
  annotations:
    hostPathProvisionerIdentity: 7de69121-4d7a-11e9-8684-0800271788ca
    pv.kubernetes.io/provisioned-by: k8s.io/minikube-hostpath
  name: pvc-54fad2fe-4d7b-11e9-9172-0800271788ca
spec:
  accessModes:
  - ReadWriteOnce
  capacity:
    storage: 2Gi
  claimRef:
    apiVersion: v1
    kind: PersistentVolumeClaim
    name: prom-prometheus-alertmanager
    namespace: default
    resourceVersion: "860"
    uid: 54fad2fe-4d7b-11e9-9172-0800271788ca
  hostPath:
    path: /tmp/hostpath-provisioner/pvc-54fad2fe-4d7b-11e9-9172-0800271788ca
  persistentVolumeReclaimPolicy: Delete
  storageClassName: standard
`

var TestServiceRaw = `{
  "apiVersion": "v1",
  "kind": "Service",
  "metadata": {
    "name": "myappservice",
    "namespace": "default",
    "resourceVersion": "187503",
    "selfLink": "/api/v1/namespaces/default/services/myappservice",
    "uid": "409de7fb-66cd-11e9-b6fa-0800271788ca"
  },
  "spec": {
    "clusterIP": "None",
    "ports": [
      {
        "port": 2222,
        "protocol": "TCP",
        "targetPort": 2222
      }
    ],
    "selector": {
      "name": "myapp"
    },
    "sessionAffinity": "None",
    "type": "ClusterIP"
  },
  "status": {
    "loadBalancer": {}
  }
}
`

var TestServiceNeat = `{
  "apiVersion": "v1",
  "kind": "Service",
  "metadata": {"name":"myappservice","namespace":"default"},
  "spec": {
    "clusterIP": "None",
    "ports": [
      {
        "port": 2222
      }
    ],
    "selector": {
      "name": "myapp"
    }
  }
}
`

var TestHealthyCertificate = `
apiVersion: cert-manager.io/v1alpha2
kind: Certificate
metadata:
  creationTimestamp: "2019-02-15T18:17:06Z"
  generation: 1
  name: test-cert
  namespace: argocd
  resourceVersion: "68337322"
  selfLink: /apis/cert-manager.io/v1alpha2/namespaces/argocd/certificates/test-cert
  uid: e6cfba50-314d-11e9-be3f-42010a800011
spec:
  acme:
    config:
    - domains:
      - cd.apps.argoproj.io
      http01:
        ingress: http01
  commonName: cd.apps.argoproj.io
  dnsNames:
  - cd.apps.argoproj.io
  issuerRef:
    kind: Issuer
    name: argo-cd-issuer
  secretName: test-secret
status:
  acme:
    order:
      url: https://acme-v02.api.letsencrypt.org/acme/order/45250083/316944902
  conditions:
  - lastTransitionTime: "2019-02-15T18:21:10Z"
    message: Order validated
    reason: OrderValidated
    status: "False"
    type: ValidateFailed
  - lastTransitionTime: null
    message: Certificate issued successfully
    reason: CertIssued
    status: "True"
    type: Ready
`

var TestDegradedCertificate = `
apiVersion: cert-manager.io/v1alpha2
kind: Certificate
metadata:
  creationTimestamp: "2019-02-15T18:17:06Z"
  generation: 1
  name: test-cert
  namespace: argocd
  resourceVersion: "68337322"
  selfLink: /apis/cert-manager.io/v1alpha2/namespaces/argocd/certificates/test-cert
  uid: e6cfba50-314d-11e9-be3f-42010a800011
spec:
  acme:
    config:
      - domains:
          - cd.apps.argoproj.io
        http01:
          ingress: http01
    commonName: cd.apps.argoproj.io
    dnsNames:
      - cd.apps.argoproj.io
    issuerRef:
      kind: Issuer
      name: argo-cd-issuer
    secretName: test-secret
status:
  acme:
    order:
      url: https://acme-v02.api.letsencrypt.org/acme/order/45250083/316944902
  conditions:
    - lastTransitionTime: "2019-02-15T18:21:10Z"
      message: Order validated
      reason: OrderValidated
      status: "False"
      type: ValidateFailed
    - lastTransitionTime: null
      message: Certificate issuance failed
      reason: Failed
      status: "False"
      type: Ready
`

var TestHealthySvc = `
apiVersion: v1
kind: Service
metadata:
  name: argocd-server
spec:
  clusterIP: 100.69.46.185
  externalTrafficPolicy: Cluster
  ports:
  - name: http
    nodePort: 30354
    port: 80
    protocol: TCP
    targetPort: 8080
  - name: https
    nodePort: 31866
    port: 443
    protocol: TCP
    targetPort: 8080
  selector:
    app: argocd-server
  sessionAffinity: None
  type: LoadBalancer
status:
  loadBalancer:
    ingress:
    - hostname: abc123.us-west-2.elb.amazonaws.com
`

var TesthealthyUnstructured = GetUnstructured(TestHealthySvc)
var TestProgressing = `
apiVersion: v1
kind: Service
metadata:
  name: argo-artifacts
spec:
  clusterIP: 10.105.70.181
  externalTrafficPolicy: Cluster
  ports:
  - name: service
    nodePort: 32667
    port: 9000
    protocol: TCP
    targetPort: 9000
  selector:
    app: minio
    release: argo-artifacts
  sessionAffinity: None
  type: LoadBalancer
status:
  loadBalancer: {}
`

var TestProgressingUnstructured = GetUnstructured(TestProgressing)

var TestUnhealthy = `
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: 2018-12-02T09:19:36Z
  name: my-pod
  namespace: argocd
  resourceVersion: "151454"
  selfLink: /api/v1/namespaces/argocd/pods/my-pod
  uid: 63674389-f613-11e8-a057-fe5f49266390
spec:
  containers:
  - command:
    - sh
    - -c
    - exit 1
    image: alpine:latest
    imagePullPolicy: Always
    name: main
    resources: {}
    terminationMessagePath: /dev/termination-log
    terminationMessagePolicy: File
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: default-token-f9jvj
      readOnly: true
  dnsPolicy: ClusterFirst
  nodeName: minikube
  restartPolicy: Always
  schedulerName: default-scheduler
  securityContext: {}
  serviceAccount: default
  serviceAccountName: default
  terminationGracePeriodSeconds: 30
  tolerations:
  - effect: NoExecute
    key: node.kubernetes.io/not-ready
    operator: Exists
    tolerationSeconds: 300
  - effect: NoExecute
    key: node.kubernetes.io/unreachable
    operator: Exists
    tolerationSeconds: 300
  volumes:
  - name: default-token-f9jvj
    secret:
      defaultMode: 420
      secretName: default-token-f9jvj
status:
  conditions:
  - lastProbeTime: null
    lastTransitionTime: 2018-12-02T09:19:36Z
    status: "True"
    type: Initialized
  - lastProbeTime: null
    lastTransitionTime: 2018-12-02T09:19:36Z
    message: 'containers with unready status: [main]'
    reason: ContainersNotReady
    status: "False"
    type: Ready
  - lastProbeTime: null
    lastTransitionTime: 2018-12-02T09:19:36Z
    status: "True"
    type: PodScheduled
  containerStatuses:
  - containerID: docker://c3aa0064b95a26045999b99c268e715a1c64201e816f1279ac06638778547bb8
    image: alpine:latest
    imageID: docker-pullable://alpine@sha256:621c2f39f8133acb8e64023a94dbdf0d5ca81896102b9e57c0dc184cadaf5528
    lastState:
      terminated:
        containerID: docker://c3aa0064b95a26045999b99c268e715a1c64201e816f1279ac06638778547bb8
        exitCode: 1
        finishedAt: 2018-12-02T09:20:25Z
        reason: Error
        startedAt: 2018-12-02T09:20:25Z
    name: main
    ready: false
    restartCount: 3
    state:
      waiting:
        message: Back-off 40s restarting failed container=main pod=my-pod_argocd(63674389-f613-11e8-a057-fe5f49266390)
        reason: CrashLoopBackOff
  hostIP: 192.168.64.41
  phase: Running
  podIP: 172.17.0.9
  qosClass: BestEffort
  startTime: 2018-12-02T09:19:36Z
`

var TestUnhealthyUnstructured = GetUnstructured(TestUnhealthy)

var TestLuaStatus = `
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: cluster-git
  namespace: argocd
spec:
  generators:
    - merge:
        generators: []
        mergeKeys:
          - server
  template:
    metadata:
      name: '{{name}}'
    spec:
      destination:
        namespace: default
        server: '{{server}}'
      project: default
      source:
        path: helm-guestbook
        repoURL: https://github.com/argoproj/argocd-example-apps/
        targetRevision: HEAD
status:
  conditions:
    - lastTransitionTime: "2021-11-12T14:28:01Z"
      message: found less than two generators, Merge requires two or more
      reason: ApplicationGenerationFromParamsError
      status: "True"
      type: ErrorOccurred
`

var TestLuaStatusUnstructured = GetUnstructured(TestLuaStatus)

var TestUnstructuredList = []unstructured.Unstructured{
	*TesthealthyUnstructured, *TestProgressingUnstructured, *TestUnhealthyUnstructured, *TestLuaStatusUnstructured,
}
