#Need to create a script to add api keys to mongo
run script troubleshoot_mongo.sh

troubleshooter --> mongo --username adminuser mongodb://mongo-svc:27017
ubuntu --> mongosh --username adminuser mongodb://localhost:27017

use auth
db.keys.insertOne({"Passwd":"f57ae22905021c0bcc0e9fad532af2787256bdbdc20f57cb4c63303e2bbd4c562a2c9ca6d79da6c02602b2b2faea41cbda8953020d0b92e0b1cecd3bd75029bb","User":"Kayleigh","Comments":"Init"})

profit

