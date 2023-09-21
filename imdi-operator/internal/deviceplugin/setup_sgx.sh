# https://github.com/istio-ecosystem/hsm-sds-server/blob/main/README.md#getting-started

## Getting started
# This section covers how to install Istio mTLS and gateway private keys protection with SGX. We use Cert Manager as default K8s CA in this document. If you want to use TCS for workload remote attestaion, please refer to this [Document](Install-with-TCS.md).

### Create signer
. ~/.profile

kubectl apply -f SgxDevicePlugin.yaml 

until [[ $(kubectl get pod -l app=intel-sgx-plugin -n inteldeviceplugins-system 2>&1 | awk '{print $3}' | grep "Running") = "Running" ]]
do
 echo "trying intel-sgx-plugin"
 kubectl apply -f SgxDevicePlugin.yaml 
 sleep 2 
done

sleep 10



