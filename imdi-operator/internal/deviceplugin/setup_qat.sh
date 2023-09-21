#! /bin/bash

. ~/.profile

kubectl apply -f qatdeviceplugin.yaml 


until [[ $(kubectl get pod -l app=intel-qat-plugin -n inteldeviceplugins-system 2>&1 | awk '{print $3}' | grep "Running") = "Running" ]]
do
 echo "trying intel-qat-plugin"
 kubectl apply -f qatdeviceplugin.yaml
 sleep 2 
done

sleep 10
