#! /bin/bash

. ~/.profile

kubectl delete -k sgx-psw/sgx_aesmd -n default

until [[ $(kubectl get pod -l app=intel-sgx-aesmd -n default 2>&1 |grep "No resources found") != "" ]]
do
 echo "checking intel-sgx-aesmd pod"
 sleep 2
done
