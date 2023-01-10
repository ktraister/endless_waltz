charts contained here are to deploy the services contained within.
Charts:
  - Reaper:
      Reaper should deploy a pod with the reaper and ETL-entropy containers
      Should contain shared volume for passthrough
  - Random:
      Random should deploy the API. It will depend on a seperate Mongo Helm chart

MONGO SETUP:
--------------------------
//we creating a new chart to deploy to local mongo
//I'm not crazy about it, but here's the non-helm yaml for now
cd kubernetes-mongodb && kubectl apply -f .

//instructions
https://devopscube.com/deploy-mongodb-kubernetes/

SVC SETUP:
--------------------------
//#gotta have that coredns.... maybe ...
$ helm repo add coredns https://coredns.github.io/helm
$ helm --namespace=kube-system install coredns coredns/coredns

The Deployment secret is created like so at cluster creation:
kubectl create secret docker-registry ghcrCred --docker-server=ghcr.io --docker-username=ktraister --docker-password=<your-pword> --docker-email=kayleigh.traister@gmail.com

Actual service deployment:
helm install ew-reaper reaper

Actual upgrade:
helm upgrade ew-reaper reaper


Above works for helm, but I'm not continuing that paradigm.
Need to setup local cluster load balancer service metallb
https://metallb.universe.tf/installation/

** NEED TO FIX kube-router SERVICE ON MY CLUSTER
#metallb pre-req
kubectl apply -f https://raw.githubusercontent.com/cloudnativelabs/kube-router/master/daemonset/kube-router-all-service-daemonset.yaml

#metallb install
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.13.7/config/manifests/metallb-native.yaml



TODOs:
------------------
- [x] deploy mongo to cluster

Reaper:
- [x] mount /dev/usb for entropy container in helm chart
- [x] reaper helm chart should run without shutting down
   - needs the ew-entropy container updated
- [x] improve logging for reaper to be useful

Random:
- [0] get Helm chart working (with mongo dep)

