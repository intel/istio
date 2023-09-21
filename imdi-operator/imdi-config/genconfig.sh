#! /bin/bash

Helplogging="
genconfig [install/sample]\n
install: print imdi-operator configuration\n
sample: print available sample custom resource
"

if [[ $1 = "genconfig" ]];
then
 if [[ $2 = "install" ]];
 then
  oldimage='docker.io\/imdi-operator:latest'
  newimage=$3
  sed -i "s|${oldimage}|${newimage}|g" imdi-all.yaml
  cat imdi-all.yaml
 elif [[ $2 = "sample" ]];
 then
  for s in $(ls ./samples); do
      echo ${s} 
  done
 else
  echo -e ${Helplogging}
 fi
else
 echo -e ${Helplogging}
fi
