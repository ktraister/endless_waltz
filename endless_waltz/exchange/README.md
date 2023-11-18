# Exchange
This service is used to serve websocket connections used to communicate by
the EW Messenger clients. It is served as a single instance in the cloud, and 
depends on the MongoDB database to authenticate users. 

## Operation
### On Start
On startup, the Exchange binary reads in configuration variables from the 
environment. These variables deal with Mongo and Logging. A goroutine is then
started which runs the Broadcaster function. A new router mutex is created, 
and the routes listed below are configured with desired methods. The HTTP 
server is then started on port 8081 with the router mutex. 

### Routes
Both routes require credentials to be set using the correct headers.
Headers "User" and "Passwd" should be set with credentials that are 
configured in the db "auth" collection. 

#### /listUsers (GET)
When this route is hit, it creates a map of unique users from the currently 
connected websocket clients. Each EW Messenger in use should have at least 
one connection to the server. The userList is returned as a string 
delimited by a colon (':').

#### /ws (GET,WEBSOCKET)
When this route is hit, the GET header "User" is checked against current 
clients. If the new request user is not unique, it is denied from 
upgrading. If the request user is unique, the request is upgraded to a 
websocket connection. The connection is then listened to indefinitely to 
receive messages using the receiver function. 

In the receiver function, messages from the client are read. If the message
is type "startup", the client username will be mapped to this connection. 
Otherwise, the message is passed to the "broadcast" channel. 

### GoRoutines

#### Broadcaster
The broadcaster routine is an indefinite loop, listening for messages on 
the "broadcast" chan. If "message.To" matches "message.From", we consider it
spam and pass further operation. Then, the broadcaster iterates over all 
clients in the map until a matching "client.Username" is found, and the 
message is sent to the matching "client.Conn". If the message is not sent
to any users, the message is considered "blackholed". The client that sent
the blackholed message is then informed of this by the "SYSTEM" user.

If any websocket connections are encountered, the errored client is removed
from the websocket exchange. 
