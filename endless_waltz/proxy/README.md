# Proxy
Proxy is used to make websocket connections from behind a packet-inspecting
firewall. Like the one here in rm 224 Hilton. 

Connection using ssh:
```
#this command works for proxying through ssh
ssh -v -NTD 127.0.0.1:9090 shenlong

#this command should work for my service
ssh -v -NTD 127.0.0.1:9090 zero@localhost -p 2222
```

## Operation
### On Start

### GoRoutines
