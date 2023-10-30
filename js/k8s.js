k8s = {
  conditions: {
    getMessage: function(v) {
      return v.healthStatus.message.trim();
    },
    getError: function(v) {
      active = []
      if (v.status == null) {
        return "No status found"
      }
      status = v.status
      if (status.conditions == null) {
        return "no conditions found"
      }
      status.conditions.forEach(function(state) {
        if (state.status == "False") {
          active.push(state)
        }
      })
      active.sort(function(a, b) { a.lastTransitionTime > b.lastTransitionTime && 1 || -1 })
      errorMessage = ""
      active.forEach(function(state) {
        if (errorMessage != "") {
          errorMessage += ', '
        }
        errorMessage += state.lastTransitionTime + ': ' + state.type + ' is ' + state.reason
        if (state.message != null) {
          errorMessage += ' with ' + state.message
        }
      })
      return errorMessage
    },
    isReady: function(v) {
      return v.healthStatus.status.toLowerCase() === "healthy" ? true : false;
    },
  },
  getAlertName: function(v) {
    name = v.alertname
    if (startsWith(v.alertname, "KubeDeployment")) {
      return name + "/" + v.deployment
    }
    if (startsWith(v.alertname, "KubePod") || startsWith(v.alertname, "ExcessivePod")) {
      return name + "/" + v.pod
    }
    if (startsWith(v.alertname, "KubeDaemonSet")) {
      return name + "/" + v.daemonset
    }
    if (v.alertname == "CertManagerInvalidCertificate") {
      return name + "/" + v.name
    }
    if (startsWith(v.alertname, "KubeStatefulSet")) {
      return name + "/" + v.statefulset
    }
    if (startsWith(v.alertname, "Node") || startsWith(v.alertname, "KubeNode")) {
      return name + "/" + v.node
    }
  },
  getAlertLabels: function(v) {
    function ignoreLabel(k) {
      return k == "severity" || k == "job" || k == "alertname" || k == "alertstate" || k == "__name__" || k == "value" || k == "namespace"
    }
    function parseLabels(v) {
      results = {}
      v.namespace = v.namespace || v.exported_namespace
      v.instance = v.exported_instance
      delete (v.exported_namespace)
      delete (v.exported_instance)
      for (k in v) {
        newKey = k.replace("label_", "")
        newKey = newKey.replace("apps_kubernetes_io_", "apps/kubernetes.io/")
        results[newKey] = v[k]
        delete v[k]
      }
      return results
    }
    v = parseLabels(v)
    if (v.alertname == "CertManagerInvalidCertificate") {
      delete (v.condition)
      delete (v.container)
      delete (v.endpoint)
      delete (v.instance)
      delete (v.service)
      delete (v.pod)
    }
    return v
  },
  getAlerts: function(results) {
    function ignoreLabel(k) {
      return k == "severity" || k == "job" || k == "alertname" || k == "alertstate" || k == "__name__" || k == "value" || k == "namespace"
    }
    function getLabels(v) {
      s = ""
      for (k in v) {
        if (ignoreLabel(k)) {
          continue
        }
        if (s != "") {
          s += " "
        }
        s += k + "=" + v[k]
      }
      return s
    }
    function getLabelMap(v) {
      out = {}
      for (k in v) {
        if (ignoreLabel(k)) {
          continue
        }
        out[k] = v[k] + ""
      }
      return out
    }
    var out = _.map(results, function(v) {
      v = k8s.getAlertLabels(v)
      return {
        pass: v.severity == "none",
        namespace: v.namespace,
        labels: getLabelMap(v),
        message: getLabels(v),
        name: k8s.getAlertName(v)
      }
    })
    JSON.stringify(out)
  },
  getNodeMetrics: function(results) {
    components = []
    for (i in results) {
      node = results[i].Object
      components.push({
        name: node.metadata.name,
        properties: [
          {
            name: "cpu",
            value: fromMillicores(node.usage.cpu)
          },
          {
            name: "memory",
            value: fromSI(node.usage.memory)
          }
        ]
      })
    }
    return components
  },

  getPodMetrics: function(results) {
    components = []
    for (i in results) {
      node = results[i].Object
      cpu = 0
      mem = 0
      for (j in node.containers) {
        cpu += fromMillicores(node.containers[j].usage.cpu)
        mem += fromSI(node.containers[j].usage.memory)
      }
      components.push({
        name: node.metadata.name,
        properties: [
          {
            name: "cpu",
            value: cpu
          },
          {
            name: "memory",
            value: mem
          }
        ]
      })
    }
    return components
  },

  filterLabels: function(labels) {
    var filtered = {}
    for (label in labels) {
      if (endsWith(label, "-hash")) {
        continue
      }
      filtered[label] = labels[label]
    }
    return filtered
  },
  getPodTopology: function(results) {
    var pods = []
    for (i in results) {
      pod = results[i].Object
      labels = k8s.filterLabels(pod.metadata.labels)
      labels.namespace = pod.metadata.namespace
      pod_mem_limit = 0
      pod_cpu_limit = 0
      if (pod.spec.containers[0].resources.limits) {
        pod_mem_limit = fromSI(pod.spec.containers[0].resources.limits.memory || 0)
        pod_cpu_limit = fromMillicores(pod.spec.containers[0].resources.limits.cpu || 0)
      }
      if (pod_mem_limit === 0) {
        pod_mem_limit = null
      }
      if (pod_cpu_limit === 0) {
        pod_cpu_limit = null
      }

      _pod = {
        name: pod.metadata.name,
        namespace: pod.metadata.namespace,
        type: "KubernetesPod",
        labels: labels,
        logs: [
          { name: "Kubernetes", type: "KubernetesPod" },
        ],
        external_id: pod.metadata.namespace + "/" + pod.metadata.name,
        configs: [
          {
            name: pod.metadata.name,
            type: "Kubernetes::Pod",
          }
        ],
        properties: [
          {
            name: "version",
            text: pod.spec.containers[0].image.split(':')[1],
            headline: true
          },
          {
            name: "cpu",
            headline: true,
            unit: "millicores",
            max: pod_cpu_limit,
          },
          {
            name: "memory",
            headline: true,
            unit: "bytes",
            max: pod_mem_limit,
          },
          {
            name: "node",
            text: pod.spec.nodeName
          },
          {
            name: "created",
            text: pod.metadata.creationTimestamp,
          },
          {
            name: "ip",
            text: pod.status.IPs != null && pod.status.IPs.length > 0 ? pod.status.IPs[0].ip : ""
          }
        ]
      }

      if (k8s.conditions.isReady(pod)) {
        _pod.status = "healthy"
      } else {
        _pod.status = "unhealthy"
        _pod.status_reason = k8s.conditions.getMessage(pod)
      }

      pods.push(_pod)
    }
    return pods
  },


  getNodeTopology: function(results) {
    var nodes = []
    for (i in results) {
      node = results[i].Object
      _node = {
        name: node.metadata.name,
        type: "KubernetesNode",
        external_id: node.metadata.name,
        labels: k8s.filterLabels(node.metadata.labels),
        selectors: [{
          name: "",
          labelSelector: "",
          fieldSelector: "node=" + node.metadata.name
        }],
        logs: [
          { name: "Kubernetes", type: "KubernetesNode" },
        ],
        configs: [
          {
            name: node.metadata.name,
            type: "Kubernetes::Node",
          }
        ],
        properties: [
          {
            name: "cpu",
            min: 0,
            unit: "millicores",
            headline: true,
            max: fromMillicores(node.status.allocatable.cpu)
          },
          {
            name: "memory",
            unit: "bytes",
            headline: true,
            max: fromSI(node.status.allocatable.memory)
          },
          {
            name: "ephemeral-storage",
            unit: "bytes",
            max: fromSI(node.status.allocatable["ephemeral-storage"])
          },
          {
            name: "instance-type",
            text: node.metadata.labels["beta.kubernetes.io/instance-type"]
          },
          {
            name: "zone",
            text: node.metadata.labels["topology.kubernetes.io/zone"]
          },
          {
            name: "ami",
            text: node.metadata.labels["eks.amazonaws.com/nodegroup-image"]
          }
        ]
      }
      internalIP = _.find(node.status.addresses, function(a) { a.type == "InternalIP" })
      if (internalIP != null) {
        _node.properties.push({
          name: "ip",
          text: internalIP.address
        })
      }
      externalIP = _.find(node.status.addresses, function(a) { a.type == "ExternalIP" })
      if (externalIP != null) {
        _node.properties.push({
          name: "externalIp",
          text: externalIP.address
        })
      }
      _node.properties.push({
        name: "os",
        text: node.status.nodeInfo.osImage + "(" + node.status.nodeInfo.architecture + ")"
      })
      for (k in node.status.nodeInfo) {
        if (k == "bootID" || k == "machineID" || k == "systemUUID" || k == "architecture" || k == "operatingSystem" || k == "osImage") {
          continue
        }
        v = node.status.nodeInfo[k]
        _node.properties.push({
          name: k.replace("Version", ""),
          text: v
        })
      }

      if (k8s.conditions.isReady(node)) {
        _node.status = "healthy"
      } else {
        _node.status = "unhealthy"
        _node.status_reason = k8s.conditions.getMessage(node)
      }

      nodes.push(_node)
    }
    return nodes
  }
}
