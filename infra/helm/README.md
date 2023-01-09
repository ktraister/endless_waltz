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
The Deployment secret is created like so at cluster creation:
kubectl create secret docker-registry ghcrCred --docker-server=ghcr.io --docker-username=ktraister --docker-password=<your-pword> --docker-email=kayleigh.traister@gmail.com

Actual service deployment:
helm install ew-reaper reaper

Actual upgrade:
helm upgrade ew-reaper reaper


TODOs:
------------------
- [x] deploy mongo to cluster

Reaper:
- [x] mount /dev/usb for entropy container in helm chart
- [x] reaper helm chart should run without shutting down
- [ ] improve logging for reaper to be useful

Random:
- [ ] get Helm chart working (with mongo dep)
