#! /bin/bash

. ~/.profile

kubectl delete -f qatdeviceplugin.yaml

until [[ $(kubectl get pod -l app=intel-qat-plugin -n inteldeviceplugins-system 2>&1 |grep "No resources found") != "" ]]
do
 echo "checking intel-qat-plugin pod"
 sleep 2
done
