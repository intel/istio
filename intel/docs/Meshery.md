# Use Meshery to deploy and manage Istio with Intel Features

## Meshery Introduction

Meshery is the open source, cloud native management plane that enables the adoption, operation, and management of Kubernetes, any service mesh, and their workloads.

### The benefit to use Meshery to deploy and manage Istio

Meshery offers us with an intuitive, visual, and convenient way to deploy Intel Accelerated Istio with one button and run performance tests to intuitively demonstrate the performance acceleration of Intel Istio.

## Architecture

![Architecture](./images/Architecture.png)

## Install Meshery

You can use one of below 3 options:

**Install on Kubernetes**:

```
$ curl -L https://meshery.io/install | PLATFORM=kubernetes bash -
```

**Install on Docker**:

```
$ curl -L https://meshery.io/install | PLATFORM=docker bash -  
```
 
**Install via Docker Compose**:

If you want more personalized configurations, such as setting proxies, please edit “.meshery/meshery.yaml”, then setup Meshery via docker compose:

```
$ docker compose -f ~/.meshery/meshery.yaml up
```

## Access and Login Meshery

Visit Meshery’s web-based user interface http://<hostname>:9081. 

![Access-meshery](./images/acess-meshery.png)

## Configure Connection to Kubernetes

Meshery attempts to auto detect your kubeconfig if it is stored in the default path ($HOME/.kube) on your system. In most deployments, Meshery will automatically connect to your Kubernetes cluster. If your config has not been auto-detected, or you want to connect remote Kubernetes cluster, you can manually locate and upload your kube config file and select the context name (docker-desktop, kind-clsuter, minikube etc.)

![Connect-k8s-cluster](./images/connect-k8s-cluster.png)

## Meshery Design 

Meshery Designs contain patterns and configurations that describe how we will deploy Istio.
Currently, we have published two Meshery Designs in Meshery Catalog website which you can import and deploy using Meshery directly:
- [CRYPTOMB-TLS-HANDSHAKE-ACCELERATION-FOR-ISTIO](https://raw.githubusercontent.com/meshery/meshery.io/master/catalog/28715e69-c6c1-4f96-bfa2-05113b00bae0.yaml): 

    CryptoMB means using Intel® Advanced Vector Extensions 512 (Intel® AVX-512) instructions using a SIMD (single instruction, multiple data) mechanism. Up to eight RSA or ECDSA operations are gathered into a buffer and processed at the same time, providing potentially improved performance. Intel AVX-512 instructions are available on recently launched 3rd generation Intel Xeon Scalable processor server processors, or later. With this Meshery Design, you can install Istio and enable CryptoMB to achieve performance improvements and accelerated handshakes.

- [QAT-TLS-HANDSHAKE-ACCELERATION-FOR-ISTIO](https://raw.githubusercontent.com/meshery/meshery.io/master/catalog/05e97933-90a6-4dd3-9b29-18e78eb4d3f1.yaml):
    
    Intel® QuickAssist Technology (QAT) provides hardware acceleration to offload the security and authentication burden from the CPU, significantly improving the performance and efficiency of standard platform solutions. With this Meshery design, you can install the Intel® QAT Device Plugin and Istio and enable QAT cryptographic acceleration for the TLS handshake in the Istio ingressgateway. This design is only available for Intel® Xeon CPUs with QAT devices enabled.

### Import Meshery Design

Take the CryptoMB TLS Handshake acceleration as example, you can import its url directly from Meshery Design page:

![meshery-design](./images/meshery-design.png)

Or, you can download the design from the url and edit it as you like, then upload it to Meshery.

### Deploy Meshery Design

After importing Meshery Design, you can deploy it in the current cluster: 

![design-deployment](./images/design-deployment.png)

It will install Istio using Istio Operator and enable CryptoMB TLS Handshake acceleration in Istio Ingressgateway.

![desing-check](./images/design-check.png)

## Applications/Backend Server

You can configure automatic sidecar injection for a namespace. Then, you can deploy sample applications:

![automatic-sidecar-injection](./images/automatic-sidecar-injection.png)
![sample-application.png](./images/sample-application.png)

You can also import your own applications from Meshery applications page via Helm Chart, Docker Compose and Kubernetes Manifests.
Here is an example file of Kuberenetes Manifests that contains a backend server for benchmarking tests called "Fortio Server":
[fortio-server-manifests.yaml](../yaml/fortio-server-manifests.yaml)

![fortio-server-manifests](./images/fortio-server-manifests.png)
![fortio-server-manifests-1](./images/fortio-server-manifests-1.png)

## Run a performance test

### Performance profile

Meshery UI provides an easy-to-use interface in which you can create performance profiles to run repeated tests with similar configuration and can also even schedule performance tests to be run at particular times through the calendar.

On the navigation menu, click on performance.

This will open the performance management dashboard and you can run the performance test with your own profile:

![performance-profile](./images/performance-profile.png)

More detailes: https://docs.meshery.io/guides/performance-management

### Run a performance test with TLS enabled in Istio Ingressgateway

From above example, we use Meshery design to deploy Istio and enable CryptoMB TLS Handshake acceleration in Istio Ingressgateway.
Therefore, we need to send HTTPs requests and upload the certificate to Istio Ingressgateway. Assuming you already have a application
backbend running in Istio and exposed them with HTTPs protocul using Istio gateway. You can upload the certificate in Meshery
performance profile and run the send the HTTPs load tests as shown below:

![http_load_tests](./images/https_load_tests.png)

## Grafana Dashboard Integration with Meshery

> **Prerequisites:**
> 
> Prometheus and Grafana installed: https://github.com/intel/istio/blob/release-1.18-intel/intel/docs/Grafana-Dashboard.md

- Login To Grafana and retrieve API Key from Settings.
- Access Meshery UI using Meshery service LB address.
- Go to Settings http://LB-IP-address/settings#metrics -> Grafana/Prometheus
- Input address of Prometheus service in Prometheus section. Check the connectivity by pinging the address button.

![prometheus-1](./images/prometheus-1.png)

- Load custom config file of the dashboard to fetch the panels.

![prometheus-2](./images/prometheus-2.png)

- Input address of Grafana service in Grafana section. Check the connectivity by pinging the address button.

All Grafana dashboards and its panels should appear in the menu.

![prometheus-3](./images/prometheus-3.png)

## Clean up

Clean up Meshery Design:

![clean-up-design](./images/clean-up-design.png)

Clean up Istio:

![clean-up-istio](./images/clean-up-istio.png)

Clean up Meshery:

```
mesheryctl system stop
```

or 

```
docker-compose -f ~/.meshery/meshery.yaml down
```