# Tor
The Tor service is used to create a `.onion` website with the 
ew-webapp service. The Tor application proxies to the NGINX
host on port 8080, allowing for http connections without redirect.
Tor isn't crazy about HTTPS, and it makes it superflous. 

## Dockerfile
The dockerfile exists to avoid depending on external containers, 
the contents of which are unknown. Due to the sensitive nature
of tor traffic, this is the most secure approach. 

## Config File
The hostname to stash in the NGINX config header to point tor
clients to the correct relay hostname can be found in the NGINX
container.
```
(ins)[~][none]> kubectl exec -it tor-699c67c656-dx7hx -- /bin/ash
/ # cat /var/lib/tor/endlesswaltz/hostname
g74rg24wiyj3ut4rdakz6cvrvvnncwvgyyyibn6y465d6y5em6husfqd.onion
```

Due to recent updates to support the prod k8s cluster, this wont
work on the localhost. Honestly, I don't want the local Tor relay
to be live, so this works out. 
