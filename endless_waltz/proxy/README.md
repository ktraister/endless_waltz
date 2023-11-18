# Proxy
Proxy is used to make websocket connections from behind a packet-inspecting
firewall. Like the one here in rm 224 Hilton. 

Connection using ssh:
```
Ended up using this command instead to forward a single port-to-port:
ssh -v -L 9090:endlesswaltz.xyz:443 zero53@localhost -p 2222
```

## Operation
### On Start

### GoRoutines
