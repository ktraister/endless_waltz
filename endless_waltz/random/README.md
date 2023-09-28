# Random
This service is used to serve random one-time pads from the MongoDB instance 
to end users. It also serves to authenticate the messenger applications on 
startup before they try to connect to Exchange. 

## Operation
### On Start
On startup, the API binary reads in configuration variables from the 
environment. These variables deal with Mongo and Logging. A new router mutex 
is created, and the routes listed below are configured with desired methods. 
The HTTP server is then started on port 8090 with the router mutex. 

### Routes
#### /api/healthcheck (GET)
The purpose of this route is to provide the client with an easy way to test 
it's credentials for authentication & authorization. Hitting this route also
proves to the client that the RandomAPI is prepared to handle traffic.

#### /api/otp (POST)
This route provides clients with raw one-time pads. The JSON body of the
request should contain the "host" identifier, set either to "client or 
"server", denoting the requester's position in the EW Connection. 

If "server" is set, the db is queried for an item that is unlocked 
({"LOCK": false}). a UUID and raw pad will be returned to the requestor, 
and the item in the db will be locked by setting {"LOCK": true}.

If "client" is set, the db is queried for an item that has the UUID matching
the UUID in the JSON body. The raw pad in the record is then returned to the 
requester. The db record that was referenced is then deleted. If there is no 
UUID in the body, the requester is warned and conn is terminated. 
