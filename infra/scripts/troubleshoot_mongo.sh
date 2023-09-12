#!/bin/bash

kubectl apply -f ./mongo_troubleshooter.yaml
echo Waiting for pod ready....
sleep 20
kubectl exec -it mongoclient /bin/bash
