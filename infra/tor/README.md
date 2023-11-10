# Tor
The Tor service is used to create a `.onion` website with the 
ew-webapp service. The Tor application proxies to the NGINX
host on port 8080, allowing for http connections without redirect.
Tor isn't crazy about HTTPS, and it makes it superflous. 

## Dockerfile
The dockerfile exists to avoid depending on external containers, 
the contents of which are unknown. Due to the sensitive nature
of tor traffic, this is the most secure approach. 
