# Proxy
Proxy is used to make websocket connections from behind a packet-inspecting
firewall. Like the one here in rm 224 Hilton. 

Firewall admins are able to disable Websockets at the firewall, but NOT when 
those websockets are tunneled through an SSH connection. Deep Packet Inspection
(DPI) can also be used to inspect your websocket traffic, leaking passwords,
who you are talking to, at what time, and more. These details are usually hidden 
by HTTPS, but DPI removes that encryption. 

## Prototyping
The proxy binary performs as a very basic VPN solution. A true VPN would be too 
heavy for our purposes, and allow users to abuse their connections easily.
Because of this, a basic SSH solution was chosen. 

The client side basically performs as stated below:
```
# Connection using ssh:
Ended up using this command instead to forward a single port-to-port:
ssh -v -L 9090:endlesswaltz.xyz:443 zero53@localhost -p 443
```

## Operation
### On Start
On start, configuration is read in through environment variables. DB config
and the SSH server keys are of primary importance. The clients are configured
to get upset if the host key they are offered does not match what they expect, 
so this is an important detail. The pubkey is compiled into the binary to
prevent exploitation. 

The SSH Server config is of special importance here. RateLimit is checked to 
prevent brute forcing attacks, but this is proving to be ineffective in production, 
and will need to be revisited. We will need to find a better way to rate limit 
incoming connections rather than using the same function we pass around for the 
web and api binaries. 

### GoRoutines
For each incoming connection, a new goroutine is spawned to handle the connection. 

### handleConnection
This function attempts to create a new connection by performing an SSH handshake. 
If the client fails to perform authentication or conform to the protocol, the 
client connection is bounced. A thread is spawned to discard requests, and the 
channels returned by the SSH connection function are then iterated through. For 
each channel, we call handleChannel.

### handleChannel
handleChannel discards all channels that are not of the type 'direct-tcpip'. 
We then accept the channel and requests. Only then do we create an outgoing connection 
to the remote. On docker-compose, the remote is 'localhost:443', and on production, it 
is 'endlesswaltz.xyz:443'. This HTTPS connection is used for proxying request traffic
by creating two distict go functions. One copies data from the destConn to the channel, 
while the other replies to requests. The main body of this thread then copies data from 
the channel to the destConn, completing the circut. 
