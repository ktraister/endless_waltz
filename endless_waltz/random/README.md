# Random
The Random service is a catchall API that is used to serve a hodgepodge of functionality.
It is deployed independently from the web application for scaling concerns and 
future decoupling. 

This service was initially developed to serve random one-time pads from the MongoDB instance 
to end users. 

After that functionality was removed, logic was added to authenticate the messenger applications on 
startup before they try to connect to Exchange. Additional functionality was 
added as time went on for various billing functionalities.  

## Operation
### On Start
On startup, the API binary reads in configuration variables from the 
environment. These variables deal with Mongo and Logging. A new router mutex 
is created, and the routes listed below are configured with desired methods. 
The HTTP server is then started on port 8090 with the router mutex. 

### Routes
Routes require credentials to be set using the correct headers.
Headers "User" and "Passwd" should be set with credentials that are 
configured in the db "auth" collection. 

#### /api/healthcheck (GET)
The purpose of this route is to provide the client with an easy way to test 
it's credentials for authentication & authorization. Hitting this route also
proves to the client that the RandomAPI is prepared to handle traffic.

#### /api/cryptoPayment (GET)
This route is used to create a crypto payment in CoinBase using the CoinBase
API. We use a One Time Password created in the database to authenticate the 
users, which is passed as a get param at execution time. If everything checks
out, the API is hit and returns a URL and billing charge that we store in the 
database for the billing binary to check for payment. The user will then
be redirected to the payment page returned by the CoinBase API. 

#### /api/create-checkout-session
Stripe Code. Creates a checkout session using the Stripe SDK and returns 
a JSON blob with the particulars. This is used by the Webapp to create an 
Iframe that is embedded into the page.

#### /api/modify-checkout-session
Stripe Code. Creates a checkout session using the Stripe SDK and returns 
a JSON blob with the particulars. This is used by the Webapp to create an 
Iframe that is embedded into the page. 

Modify is distinct from create in that there isn't a 30 day free trial, 
but instead a trial length dictated by the number of days the user should
be allowed to operate before charging. 

#### /api/session-status
Checks with Stripe for the provided session ID. If found, it returns a json
blob with the status of the provided ID. Used by the web app. 


---
This code was mothballed when Kyber was rolled out :)
```
#### /api/otp (POST)
This route provides clients with raw one-time pads. The JSON body of the
request should contain the "host" identifier, set either to "client or 
"server", denoting the requester's position in the EW Connection. All 
records accessed in this path are stored in the "otp_db" collection.

If "server" is set, the db is queried for an item that is unlocked 
({"LOCK": nil}). a UUID and raw pad will be returned to the requestor, and 
the item in the db will be locked by setting {"LOCK": true}.

If "client" is set, the db is queried for an item that has the UUID matching
the UUID in the JSON body. The raw pad in the record is then returned to the 
requester. The db record that was referenced is then deleted. If there is no 
UUID in the body, the requester is warned and conn is terminated. 
```
