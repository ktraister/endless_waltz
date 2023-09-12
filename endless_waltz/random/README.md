Init readme for service

#Need to create a script to add api keys to mongo
run script troubleshoot_mongo.sh

mongosh --username adminuser mongodb://localhost:27017

use auth
db.keys.insertOne({"API-Key":"arandomnumber","User":"Kayleigh","Comments":"Init"})

profit
