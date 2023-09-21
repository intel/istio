#! /bin/bash

if [[ $(cat ~/.profile | grep "export http_proxy") = "" ]];
then
echo "export http_proxy=$1" >> ~/.profile
else
sed -i "s|http_proxy=.*|http_proxy=$1|g"  ~/.profile
fi

if [[ $(cat ~/.profile | grep "export https_proxy") = "" ]];
then
echo "export https_proxy=$2" >> ~/.profile
else
sed -i "s|https_proxy=.*|https_proxy=$2|g"  ~/.profile
fi

if [[ $(cat ~/.profile | grep "export no_proxy") = "" ]];
then
echo "export no_proxy=$3" >> ~/.profile
else
sed -i "s|no_proxy=.*|no_proxy=$3|g"  ~/.profile
fi