#! /bin/bash

. ~/.profile

kubectl delete -f SgxDevicePlugin.yaml 

until [[ $(kubectl get pod -l app=intel-sgx-plugin -n inteldeviceplugins-system 2>&1 |grep "No resources found") != "" ]]
do
 echo "checking intel-sgx-plugin pod"
 sleep 2
done
