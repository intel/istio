# Imdi-operator

## Building images

1. Build and push `imdi-operator` image to the location specified by `OPERATOR-IMG`:

```sh
TAG=$(date "+%Y%m%d-%H%M") ; make docker-build-operator docker-push-operator OPERATOR-IMG=registry.fi.intel.com/yizhouxu/imdi-operator:$TAG
```

2. Build and push `imdi-config` image to the location specified by `CONFIG-IMG`:

```sh
TAG=$(date "+%Y%m%d-%H%M") ;
REGISTRY_PATH=registry.fi.intel.com/yizhouxu; make docker-build-config docker-push-config CONFIG-IMG=$REGISTRY_PATH/imdi-config:latest
```

## Deployment

### Oneclick deployment

1. Deploy/Undeploy the operator to k8s cluster by using `kubectl apply` from `imdi-config` image:

```sh
REGISTRY_PATH=docker.io/intel
TAG=latest

# deploy operator
kubectl run imdi-config -qit --rm --image=$REGISTRY_PATH/imdi-config:$TAG --restart=Never -- genconfig install $REGISTRY_PATH/imdi-operator:$TAG | kubectl apply -f -

# delete operator
kubectl run imdi-config -qit --rm --image=$REGISTRY_PATH/imdi-config:$TAG --restart=Never -- genconfig install $REGISTRY_PATH/imdi-operator:$TAG | kubectl delete -f -
```



2. Get sample CR:

1) print available custom resource:
```sh
kubectl run imdi-config -qit --rm --image=$REGISTRY_PATH/imdi-config:$TAG --restart=Never -- genconfig sample
```

2) deploy sample CR you want to use:
```sh
# deploy Istio with no hardware feature
kubectl run imdi-config -qit --rm --image=$REGISTRY_PATH/imdi-config:$TAG --restart=Never --command cat samples/imdi-servicemesh_v1_imdioperator.yaml | kubectl apply -f -
# uninstall Istio with no hardware feature
kubectl run imdi-config -qit --rm --image=$REGISTRY_PATH/imdi-config:$TAG --restart=Never --command cat samples/imdi-servicemesh_v1_imdioperator.yaml | kubectl delete -f -

# deploy Istio with Qat
kubectl run imdi-config -qit --rm --image=$REGISTRY_PATH/imdi-config:$TAG --restart=Never --command cat samples/imdi-servicemesh_v1_imdioperator_with_qat.yaml | kubectl apply -f -
kubectl run imdi-config -qit --rm --image=$REGISTRY_PATH/imdi-config:$TAG --restart=Never --command cat samples/imdi-servicemesh_v1_imdioperator_with_qat.yaml >imdi-servicemesh_v1_imdioperator_with_qat.yaml 
# uninstall Istio with Qat
kubectl run imdi-config -qit --rm --image=$REGISTRY_PATH/imdi-config:$TAG --restart=Never --command cat samples/imdi-servicemesh_v1_imdioperator_with_qat.yaml | kubectl delete -f -

# deploy Istio with CryptoMB 
kubectl run imdi-config -qit --rm --image=$REGISTRY_PATH/imdi-config:$TAG --restart=Never --command cat samples/imdi-servicemesh_v1_imdioperator_with_cryptomb.yaml | kubectl apply -f -
# uninstall Istio with CryptoMB 
kubectl run imdi-config -qit --rm --image=$REGISTRY_PATH/imdi-config:$TAG --restart=Never --command cat samples/imdi-servicemesh_v1_imdioperator_with_cryptomb.yaml | kubectl delete -f -
```

### Deployment manually

```sh
make deploy OPERATOR-IMG=<some-registry>/imdi-operator:tag
```

```sh
make undeploy
```
