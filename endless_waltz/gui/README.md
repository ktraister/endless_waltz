# Messenger
This service is used by end users to send messages to each other encrypted 
with one-time pads served by the RandomAPI. Different instances communicate 
with each other through the Exchange websockets. 

## Operation
All the GUI components herein are created using the Fyne framework in Go. 

### Startup
On startup, the Messenger checks if the file `~/.ew/config.yml` exists, 
creating it with default contents if it does not, along with directories. 
The config is then read by viper into a `Configurations` struct and 
returned to main. A new GUI app is created, and a login window is spawned.
To check authentication, the GUI reaches out to the RandomAPI healthcheck
path. Once the user authenticates correctly, a new messenger window is 
opened.

### Messaging
Once login is complete, the Messenger GUI is configured with all the 
appropriate windows. A GoRoutine to refresh the online users panel is 
then started. Buttons to send messages and clear the screen are then 
configured. The GoRoutines for listening for messages,sending messages, 
and posting new messages to the GUI are then started. The window is then 
resized and run. 

### GoRoutines

#### Listen
This routine spawns a connection with the exchange to listen for HELOs.
Then, it listens indefinitely for HELOs, and spawns GoRoutines to handle
individual connections.

#### HandleConnection
GoRoutines for handling connections are spawned when a user HELOs the 
server side. The routine creates a new connection to the exchange, and 
responds with a HELO to the initiating connection. The server goes
through the EW Protocol to receive a single message. After the message 
is received, the message is placed in the incomingMsgChan channel, and the
connection is closed. 

#### Send
The send GoRoutine listens for messages being placed on the outgoingMsgChan
channel in an indefinite loop. When one is received, the send button and 
input text box are hidden. A new websocket connection to the exchange is 
created, and the client HELOs to the target user. If the message is 
blackholed or HELO is not received within 5 seconds, the client exits. 
Otherwise, the responding user is used for the rest of the transaction. 
Sent messages are placed on the incomingMsgChan to be posted by the Post
routine. 

#### Post
In an indefinite loop, the post GoRoutine listens for messages on the 
incomingMsgChan. These messages are used to show successful and failed 
messages, as well as the target user. If the current target user is not 
selected, the messages are stashed until the user is selected. 

#### RefreshUsers
In an indefinite loop, the client reaches out to the Exchange API path
`/listUsers` every five seconds. The returned current client list is then
alphabetized and returned to the UI. The UI is refreshed accordingly.
