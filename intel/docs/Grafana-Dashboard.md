# Grafana Dashboard
Prerequisites:
* Kubernetes cluster with at least one Ice Lake node
* Cert-Manager
* Istioctl
* Helm

This README contains:
- [Installing Grafana Dashboard](#deployment)
- [Uninstallation](#uninstallation)

## Deployment

Install Istio with the BOOTSTRAP_XDS_AGENT environment variable and Envoy Filter with the cryptomb values.
```
istioctl install -f intel/yaml/intel-istio-cryptomb.yaml
kubectl apply -f intel/yaml/envoy-filter-cryptomb-stats.yaml
```

Install Cert-Manager issuer and certificate for the Istio-IngressGateway TLS endpoint.
```
kubectl apply -f intel/yaml/grafana-gw-certificates.yaml
```

Create a separate namespace to isolate monitoring services and applications.
```
kubectl create ns monitoring
```

Install Kube-Prometheus-Stack.
```
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm install kube-prometheus-stack prometheus-community/kube-prometheus-stack -n monitoring
helm upgrade kube-prometheus-stack prometheus-community/kube-prometheus-stack -f intel/yaml/grafana-dashboard-values.yaml -n monitoring
```

Install Istio Ingress Gateway and Virtual Services for Grafana and Prometheus.
```
kubectl apply -f intel/yaml/grafana-istio-gw-vs.yaml
```
To access Grafana and Prometheus UIs, use:
* a) Istio-IngressGateway service load balancing address.
* b) Kubectl port-forwarding to the service and connect localhost.
  Paths are /grafana/ and /prometheus/.
Grafana UI Credentials:
 * login: admin
 * password: prom-operator

Restart Envoy of the Istio Ingress Gateway.
```
kubectl rollout restart deployment -n istio-system istio-ingressgateway
```
Import the dashboard using Grafana UI.

1. Login to Grafana
2. Create -> Import -> Upload JSON file (intel/json/intel-distribution-of-istio.json)

## Uninstallation
Use the following command to uninstall kube-prometheus-stack from cluster.
```
helm uninstall kube-prometheus-stack -n monitoring
```

CRDs created by this chart are not removed by default and should be manually cleaned up:

```
kubectl delete crd alertmanagerconfigs.monitoring.coreos.com
kubectl delete crd alertmanagers.monitoring.coreos.com
kubectl delete crd podmonitors.monitoring.coreos.com
kubectl delete crd probes.monitoring.coreos.com
kubectl delete crd prometheusagents.monitoring.coreos.com
kubectl delete crd prometheuses.monitoring.coreos.com
kubectl delete crd prometheusrules.monitoring.coreos.com
kubectl delete crd scrapeconfigs.monitoring.coreos.com
kubectl delete crd servicemonitors.monitoring.coreos.com
kubectl delete crd thanosrulers.monitoring.coreos.com
```

Remove deployment.
```
kubectl delete -f intel/yaml/envoy-filter-cryptomb-stats.yaml -f intel/yaml/grafana-gw-certificates.yaml -f intel/yaml/grafana-istio-gw-vs.yaml
istioctl x uninstall --purge # to delete all Istio components from the cluster
```