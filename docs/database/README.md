### MongoDB
MongoDB is a NoSQL database known for high speed and flexibility. It serves as 
the persistent data store for one-time pads and user authentication data. 

## Infrastructure
Currently, Mongo is served within K8s using a persistent volume claim to 
persist data. The service files used can be found in `../../infra/k8s/mongodb/`

## Troubleshooting
This script allows you to get a mongo shell even if the port for the service 
is not exposed outside the k8s host. `troubleshoot_mongo.sh`

These commands are used to connect to a mongo instance as the `adminuser`:
```
troubleshooter --> mongo --username adminuser mongodb://mongo-svc:27017
ubuntu --> mongosh --username adminuser mongodb://localhost:27017
```

## Adding a User
The Mongo shell is built on top of a JavaScript interface. The following 
commands can be used to switch to the `auth` database and add a user:
```
use auth
db.keys.insertOne({"Passwd":"f57ae22905021c0bcc0e9fad532af2787256bdbdc20f57cb4c63303e2bbd4c562a2c9ca6d79da6c02602b2b2faea41cbda8953020d0b92e0b1cecd3bd75029bb","User":"Kayleigh","Comments":"Init"})
```
