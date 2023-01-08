charts contained here are to deploy the services contained within. 
Charts:
  - Reaper:
      Reaper should deploy a pod with the reaper and ETL-entropy containers
      Should contain shared volume for passthrough
  - Random:
      Random should deploy the API. It will depend on a seperate Mongo Helm chart

The Deployment secret is created like so at cluster creation:
kubectl create secret docker-registry ghcrCred --docker-server=ghcr.io --docker-username=ktraister --docker-password=<your-pword> --docker-email=kayleigh.traister@gmail.com

Actual service deployment:
helm install ew-reaper reaper

Actual upgrade:
helm upgrade ew-reaper reaper

TODOs:
------------------
Reaper:
- [ ] mount /dev/usb for entropy container in helm chart
- [ ] reaper helm chart should run without shutting down
- [ ] improve logging for reaper to be useful

Random:
- [ ] deploy mongo to cluster
- [ ] get Helm chart working (with mongo dep)
