
# 🧬 Kratos Assign Operator

The **Kratos Assign Operator** is a Kubernetes controller that automatically mutates Deployments by applying rules defined in custom `Assign` resources. It supports patching fields like `nodeSelector` and `tolerations` at the Deployment level by analyzing the current state of Pods and tracing them back to their owning ReplicaSets and Deployments.


## ✨ Features

* ✅ Patch `nodeSelector` fields on Deployments
* ✅ Add `tolerations` to target Deployments
* ✅ Namespace-aware resource matching using label selectors
* ✅ Conflict-resilient update mechanism using exponential backoff
* ✅ Fully RBAC-secured and extendable design


## 📦 Custom Resource Definition

The operator introduces a new custom resource: `Assign`.

### CRD: `Assign`

```yaml
apiVersion: mutations.kratos.dev/v1alpha1
kind: Assign
metadata:
  name: example-assign
spec:
  applyTo:
    - groups:
        - ""
      kinds:
        - Pod
      versions:
        - v1
  location: spec.nodeSelector # or spec.tolerations
  match:
    kinds:
      - apiGroups:
          - "*"
        kinds:
          - Pod
    namespaceSelector:
      matchLabels:
        kubernetes.io/metadata.name: example-namespace
    scope: Namespaced
  parameters:
    assign:
      value:
        key: value # either a nodeSelector map or toleration array
```


## 📂 Example Resources

### 🧲 Apply Node Selector

```yaml
apiVersion: mutations.kratos.dev/v1alpha1
kind: Assign
metadata:
  name: example-nodeselector
spec:
  applyTo:
    - groups: [""]
      kinds: ["Pod"]
      versions: ["v1"]
  location: spec.nodeSelector
  match:
    kinds:
      - apiGroups: ["*"]
        kinds: ["Pod"]
    namespaceSelector:
      matchLabels:
        kubernetes.io/metadata.name: k8scostai
    scope: Namespaced
  parameters:
    assign:
      value:
        k8scostai.com/spot: "true"
```

### 🧱 Apply Tolerations

```yaml
apiVersion: mutations.kratos.dev/v1alpha1
kind: Assign
metadata:
  name: example-tolerations
spec:
  applyTo:
    - groups: [""]
      kinds: ["Pod"]
      versions: ["v1"]
  location: spec.tolerations
  match:
    kinds:
      - apiGroups: [""]
        kinds: ["Pod"]
    namespaceSelector:
      matchLabels:
        kubernetes.io/metadata.name: k8scostai
    scope: Namespaced
  parameters:
    assign:
      value:
        - key: "k8scostai.com/scalesetpriority"
          operator: "Equal"
          value: "spot"
          effect: "NoSchedule"
```



## ⚙️ How It Works

1. The controller watches `Assign` custom resources.
2. For each `Assign`, it selects Pods based on `match.namespaceSelector` and `match.kinds`.
3. It identifies the Deployment via the Pod → ReplicaSet → Deployment chain.
4. Depending on the `location`, it updates the Deployment spec accordingly:

   * `spec.nodeSelector`: Adds or updates key-value pairs
   * `spec.tolerations`: Appends toleration rules
5. Uses exponential backoff retry logic to handle conflicts during updates.


## 🔐 RBAC Permissions

The operator requires access to the following resources:

```yaml
- groups: [mutations.kratos.dev]
  resources: [assigns, assigns/status, assigns/finalizers]
  verbs: [get, list, watch, create, update, patch, delete]
- groups: [apps]
  resources: [deployments, replicasets]
  verbs: [get, list, watch, create, update, patch]
- groups: [core]
  resources: [pods]
  verbs: [get, list, watch, update, patch]
```


## 🚀 Getting Started

1. **Install the CRDs**:

   ```bash
   kubectl apply -f config/crd/bases/mutations.kratos.dev_assigns.yaml
   ```

2. **Deploy the controller**:

   ```bash
   make docker-build docker-push IMG=<your-image>
   make deploy IMG=<your-image>
   ```

3. **Create `Assign` resources** to define your mutation logic.


## 🛠️ Development

* Requires:

  * Go 1.20+
  * Kubernetes v1.25+
  * [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)

```bash
make install
make run
```




