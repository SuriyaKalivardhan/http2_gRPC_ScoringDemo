#!/bin/bash

for i in $(eval echo {1..$1})
do 
	curl suriyakvm.westus2.cloudapp.azure.com:5001/healthcheck
done
