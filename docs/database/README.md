### MongoDB
MongoDB is a NoSQL database known for high speed and flexibility. It serves as 
the persistent data store for one-time pads and user authentication data. 

## Application
The application uses two distinct database collections within Mongo:
`auth` and `otp`. `auth` is used for storing user credentials, whose records
include notes and will eventually other information. `otp` is used to store 
records with UUIDs and one-time pads, written by the reaper service.

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
db.keys.insertOne({
    SignupTime: '1697504433',
    Active: true,
    Email: 'Kayleigh.traister@gmail.com',
    EmailVerifyToken: 'N5ZJgRAWO0keQI81YGawv9xNxMc0EKjFgFAvzVwxPm9srDn7WzDS8M66AdAPpSiIWN3V4cF7BjD8VoLemsDa9P1bCeGGGnHHClDefAaIDcaZ6qFhQuCsqWSAOAUqJQPu',
    User: 'Kayleigh',
    Passwd: ''
})
db.keys.find({})
```

## Backup/Restore
Mongo Backups are handled by a github actions script, and then pushed to s3
```
#dump command
kubectl exec mongo-659c8dc68-mswzv -- /usr/bin/mongodump --archive --authenticationDatabase admin -u $USER -p $PASSWD --db keys > db.dump

#restore command
kubectl exec mongo-659c8dc68-mswzv -- /usr/bin/mongorestore --archive --authenticationDatabase admin -u $USER -p $PASSWD --db keys < db.dump
```
