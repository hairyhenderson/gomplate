package kubernetes

var TestHealthy = `
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
