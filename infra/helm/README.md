charts contained here are to deploy the services contained within. 
Charts:
  - Reaper:
      Reaper should deploy a pod with the reaper and ETL-entropy containers
      Should contain shared volume for passthrough
  - Random:
      Random should deploy the API. It will depend on a seperate Mongo Helm chart
