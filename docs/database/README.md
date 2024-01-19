# MongoDB
MongoDB is a NoSQL database known for high speed and flexibility. It serves as 
the persistent data store for one-time pads and user authentication data. 

## Application
The application uses one distinct database collections within Mongo:
`auth`. `auth` is used for storing user credentials, whose records
include notes and will eventually other information.  

### Data Structure
This is a single mongo record. Mongo does not force us into a rigid schema, 
so this is the maximum that can be set by all application paths.
```
  {
    //mongo specific id
    _id: ObjectId('654d20acd11f0aaed34ff843'),

    //account values -- active:false will disallow login for messenger
    SignupTime: '1699553450',
    Active: true,
    Premium: true,
    FriendsList: 'item1:item2',

    //user credentials
    Email: 'kayleigh.traister@gmail.com',
    EmailVerifyToken: 'Z8HOJHZ...FCmrI1PS37',
    User: 'zero53',
    Passwd: '',

    //password reset values
    passwordResetTime: Long('1699553760'),
    passwordResetToken: 'lnnpU...oSypNlf',

    //global billing values
    billingCycleEnd: '01-01-2024'        //MM-DD-YYYY

    //crypto billing values
    cryptoBilling: true
    billingEmailSent: false            //crypto specific
    billingReminderSent: false         //crypto specific
    billingCharge: '2E8YCQWQ',         //also crypto specific
    billingToken: 'lnnpU...oSypNlf',   //also crypto specific

    //card billing values
    cardBilling: true
    cardBillingId: "cus_P81rSXvuzrd44t" //subscription ID is used to tie customer to subscription in stripe
  },

db.keys.updateOne( { User: "zero53" }, { $set: { "cryptoBilling": true, billingCycleEnd: "12-06-2023", billingEmailSent: false, billingReminderSent: false }, $currentDate: { lastModified: true } } )
```

## Infrastructure
Currently, Mongo is served within K8s using a persistent volume claim to 
persist data. The service files used can be found in `../../infra/k8s/mongodb/`

In the future, I'd like to move to an operator that lives in the K3s cluster and allows for easy service configuration and operation.
This will be undertaken in https://github.com/ktraister/endless_waltz/issues/351

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
    EmailVerifyToken: 'N5ZJgRA...AOAUqJQPu',
    User: 'Kayleigh',
    Passwd: ''
})
db.keys.find({})
```

## Backup/Restore
Mongo Backups are handled by a github actions script, and then pushed to s3
```
#dump command
kubectl exec mongo-659c8dc68-mswzv -- /usr/bin/mongodump --archive --authenticationDatabase admin -u $USER -p $PASSWD > db.dump

#restore command
docker cp db.dump 864ee1eeb02c:/tmp/db.dump
  Successfully copied 5.12kB to 864ee1eeb02c:/tmp/db.dump
docker exec 864ee1eeb02c /usr/bin/mongorestore --authenticationDatabase admin -u $USER -p $PASS --archive=/tmp/db.dump
```
