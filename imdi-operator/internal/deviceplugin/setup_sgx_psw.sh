# This section covers how to install SGX PSW and AESMD Daemonset.
. ~/.profile

kubectl apply -k sgx-psw/sgx_aesmd -n default

until [[ $(kubectl get pod -l app=intel-sgx-aesmd -n default 2>&1 | awk '{print $3}' | grep "Running") = "Running" ]]
do
 echo "trying intel-sgx-aesmd"
 kubectl apply -k sgx-psw/sgx_aesmd
 sleep 2 
done

sleep 10