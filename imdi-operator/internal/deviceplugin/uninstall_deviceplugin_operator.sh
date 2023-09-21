#! /bin/bash

. ~/.profile
echo "kubectl delete -k $4"
set -e
kubectl delete -k $4
set +e
until [[ $(kubectl get pod -l control-plane=controller-manager -n inteldeviceplugins-system 2>&1 |grep "No resources found") != "" ]]
do
 echo "checking device plugin operator pod"
 sleep 2
done

kubectl delete -f $3
kubectl delete -k $2
kubectl delete -k $1

until [[ $(kubectl get pod -l app.kubernetes.io/name=webhook -n cert-manager 2>&1 |grep "No resources found") != "" ]]
do
 echo "checking webhook pod"
 sleep 2
done

until [[ $(kubectl get pod -l app.kubernetes.io/name=cainjector -n cert-manager 2>&1 |grep "No resources found") != "" ]]
do
 echo "checking cainjector pod"
 sleep 2
done

until [[ $(kubectl get pod -l app=cert-manager -n cert-manager 2>&1 |grep "No resources found") != "" ]]
do
 echo "checking cert-manage pod"
 sleep 2
done

until [[ $(kubectl get pod -l app=nfd-master -n node-feature-discovery 2>&1 |grep "No resources found") != "" ]]
do
 echo "checking nfd-master pod"
 sleep 2
done
