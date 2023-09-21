#! /bin/bash

. ~/.profile
env | grep proxy
set -e
kubectl apply -k $1 
set +e
kubectl apply -k $2
kubectl apply -f $3

echo "checking webhook pod"
until [[ $(kubectl get pod -l app.kubernetes.io/name=webhook -n cert-manager| awk '{print $3}' | grep "Running") = "Running" ]]
do
 echo "checking webhook pod"
 sleep 5
done

echo "checking cainjector pod"
until [[ $(kubectl get pod -l app.kubernetes.io/name=cainjector -n cert-manager| awk '{print $3}' | grep "Running") = "Running" ]]
do
 echo "checking cainjector pod"
 sleep 5
done

echo "checking cainjector pod log"
until [[ $(kubectl logs $(kubectl get pod -l app.kubernetes.io/name=cainjector -o name -n cert-manager) -n cert-manager 2>&1 | grep "Updated object") != "" ]]
do
 echo "checking cainjector pod log"
 sleep 5
done

echo "checking webhook pod log"
until [[ $(kubectl logs $(kubectl get pod -l app.kubernetes.io/name=webhook -o name -n cert-manager) -n cert-manager 2>&1 | grep "Updated cert-manager webhook TLS") != "" ]]
do
 echo "checking webhook pod log"
 sleep 5
done

echo "trying to install device-plugin-operator"
until [[ $(kubectl apply -k $4 2>&1 | grep "certificate.cert-manager.io/inteldeviceplugins-serving-cert") != "" ]]
do
 echo "trying to install device-plugin-operator"
 sleep 10
done

echo "checking device-plugin-operator pod"
until [[ $(kubectl get pod -l control-plane=controller-manager -n inteldeviceplugins-system | awk '{print $3}' | grep "Running") = "Running" ]]
do
 echo "checking device-plugin-operator pod"
 sleep 5
done

echo "checking device-plugin-operator pod log"
until [[ $(kubectl logs $(kubectl get pod -l control-plane=controller-manager -o name -n inteldeviceplugins-system) -n inteldeviceplugins-system 2>&1 | grep "Starting workers") != "" ]]
do
 echo "checking device-plugin-operator pod log"
 sleep 5
done

echo "install deviceplugin successful."