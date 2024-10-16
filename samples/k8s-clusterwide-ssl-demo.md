## Kubernetes Clusterwide SSL Demo

### Setup

Create a Kind cluster by using an ansible playbook from the project `ansible` folder.

```bash
cd ../ansible && ansible-playbook playbooks/kind-cluster.yaml
```

```bash
cd ../ansible && ansible-playbook playbooks/setup-ecosystem.yaml
```

```bash
helm upgrade --install seldon-core-v2-crds  ../k8s/helm-charts/seldon-core-v2-crds -n seldon-mesh
```

```
Release "seldon-core-v2-crds" has been upgraded. Happy Helming!
NAME: seldon-core-v2-crds
LAST DEPLOYED: Thu Jun 22 09:11:26 2023
NAMESPACE: seldon-mesh
STATUS: deployed
REVISION: 2
TEST SUITE: None

```

```bash
kubectl create namespace ns1
```

```
namespace/ns1 created

```

```bash
kubectl create namespace ns2
```

```
namespace/ns2 created

```

```bash
helm install seldon-v2-certs ../k8s/helm-charts/seldon-core-v2-certs/ -n ns1
```

```yaml
NAME: seldon-v2-certs
LAST DEPLOYED: Thu Jun 22 10:00:32 2023
NAMESPACE: ns1
STATUS: deployed
REVISION: 1
TEST SUITE: None

```

```bash
helm install seldon-v2-certs ../k8s/helm-charts/seldon-core-v2-certs/ -n ns2
```

```yaml
NAME: seldon-v2-certs
LAST DEPLOYED: Thu Jun 22 09:11:36 2023
NAMESPACE: ns2
STATUS: deployed
REVISION: 1
TEST SUITE: None

```

```bash
cat ../k8s/samples/values-tls-dataplane-controlplane-example.yaml
```

```yaml
security:
  controlplane:
    protocol: SSL
    ssl:
      server:
        secret: seldon-controlplane-server
        clientValidationSecret: seldon-controlplane-client
      client:
        secret: seldon-controlplane-client
        serverValidationSecret: seldon-controlplane-server
  envoy:
    protocol: SSL
    ssl:
      upstream:
        server:
          secret: seldon-upstream-server
          clientValidationSecret: seldon-upstream-client
        client:
          secret: seldon-upstream-client
          serverValidationSecret: seldon-upstream-server
      downstream:
        client:
          serverValidationSecret: seldon-downstream-server
        server:
          secret: seldon-downstream-server

```

```bash
helm install seldon-v2 ../k8s/helm-charts/seldon-core-v2-setup/ -n seldon-mesh --set controller.clusterwide=true --values ../k8s/samples/values-tls-dataplane-controlplane-example.yaml
```

```yaml
NAME: seldon-v2
LAST DEPLOYED: Thu Jun 22 09:11:40 2023
NAMESPACE: seldon-mesh
STATUS: deployed
REVISION: 1
TEST SUITE: None

```

```bash
helm install seldon-v2-runtime ../k8s/helm-charts/seldon-core-v2-runtime  -n ns1
```

```yaml
NAME: seldon-v2-runtime
LAST DEPLOYED: Thu Jun 22 10:00:36 2023
NAMESPACE: ns1
STATUS: deployed
REVISION: 1
TEST SUITE: None

```

```bash
helm install seldon-v2-servers ../k8s/helm-charts/seldon-core-v2-servers  -n ns1
```

```yaml
NAME: seldon-v2-servers
LAST DEPLOYED: Thu Jun 22 10:00:36 2023
NAMESPACE: ns1
STATUS: deployed
REVISION: 1
TEST SUITE: None

```

```bash
helm install seldon-v2-runtime ../k8s/helm-charts/seldon-core-v2-runtime  -n ns2
```

```yaml
NAME: seldon-v2-runtime
LAST DEPLOYED: Thu Jun 22 09:11:49 2023
NAMESPACE: ns2
STATUS: deployed
REVISION: 1
TEST SUITE: None

```

```bash
helm install seldon-v2-servers ../k8s/helm-charts/seldon-core-v2-servers  -n ns2
```

```yaml
NAME: seldon-v2-servers
LAST DEPLOYED: Thu Jun 22 09:12:16 2023
NAMESPACE: ns2
STATUS: deployed
REVISION: 1
TEST SUITE: None

```

```bash
kubectl wait --for condition=ready --timeout=300s server --all -n ns1
```

```
server.mlops.seldon.io/mlserver condition met
server.mlops.seldon.io/triton condition met

```

```bash
kubectl wait --for condition=ready --timeout=300s server --all -n ns2
```

```
server.mlops.seldon.io/mlserver condition met
server.mlops.seldon.io/triton condition met

```

### Setup TLS Config

```python
MESH_IP=!kubectl get svc seldon-mesh -n ns1 -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
MESH_IP_NS1=MESH_IP[0]
import os
os.environ['MESH_IP_NS1'] = MESH_IP_NS1
MESH_IP_NS1
```

```
'172.21.255.2'

```

```python
MESH_IP=!kubectl get svc seldon-mesh -n ns2 -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
MESH_IP_NS2=MESH_IP[0]
import os
os.environ['MESH_IP_NS2'] = MESH_IP_NS2
MESH_IP_NS2
```

```
'172.21.255.4'

```

```bash
seldon config add tls ${PWD}/config-dataplane-tls.json
```

```bash
seldon config activate tls
```

### Launch model in namespace ns1

```bash
kubectl create -f ./models/sklearn-iris-gs.yaml -n ns1
```

```
model.mlops.seldon.io/iris created

```

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n ns1
```

```
model.mlops.seldon.io/iris condition met

```

```bash
seldon model infer iris --inference-host ${MESH_IP_NS1}:80 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```json
{
	"model_name": "iris_1",
	"model_version": "1",
	"id": "77f21360-b6cf-4c50-a000-1f219f5e9e29",
	"parameters": {},
	"outputs": [
		{
			"name": "predict",
			"shape": [
				1,
				1
			],
			"datatype": "INT64",
			"parameters": {
				"content_type": "np"
			},
			"data": [
				2
			]
		}
	]
}

```

### Launch model in namespace ns2

```bash
kubectl create -f ./models/sklearn-iris-gs.yaml -n ns2
```

```
model.mlops.seldon.io/iris created

```

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n ns2
```

```
model.mlops.seldon.io/iris condition met

```

```bash
seldon model infer iris --inference-host ${MESH_IP_NS2}:80 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```json
{
	"model_name": "iris_1",
	"model_version": "1",
	"id": "902b8483-0ede-4136-97be-dbf266f05a0c",
	"parameters": {},
	"outputs": [
		{
			"name": "predict",
			"shape": [
				1,
				1
			],
			"datatype": "INT64",
			"parameters": {
				"content_type": "np"
			},
			"data": [
				2
			]
		}
	]
}

```

## TearDown

```bash
kubectl delete -f ./models/sklearn-iris-gs.yaml -n ns1
```

```
model.mlops.seldon.io "iris" deleted

```

```bash
kubectl delete -f ./models/sklearn-iris-gs.yaml -n ns2
```

```
model.mlops.seldon.io "iris" deleted

```

```bash
helm delete seldon-v2-servers -n ns1
helm delete seldon-v2-runtime -n ns1
helm delete seldon-v2-certs -n ns1
```

```
release "seldon-v2-servers" uninstalled
release "seldon-v2-runtime" uninstalled
release "seldon-v2-certs" uninstalled

```

```bash
helm delete seldon-v2-runtime -n ns2
helm delete seldon-v2-runtime -n ns2
helm delete seldon-v2-certs -n ns2
```

```
release "seldon-v2-runtime" uninstalled
Error: uninstall: Release not loaded: seldon-v2-runtime: release: not found
release "seldon-v2-certs" uninstalled

```

```bash
helm delete seldon-v2  -n seldon-mesh
```

```
release "seldon-v2" uninstalled

```

```bash
helm delete seldon-core-v2-crds -n seldon-mesh
```

```
release "seldon-core-v2-crds" uninstalled

```

```bash
kubectl delete namespace ns1
kubectl delete namespace ns2
```

```
namespace "ns1" deleted
namespace "ns2" deleted

```

```python

```
