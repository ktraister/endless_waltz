### K8s Infra

# Purpose
--------------
The purpose of this is to create a set of scripts to build a cluster on any host.

HOST SETUP:
--------------------------
## install k8s
get your new cluster set up, start with a master node:
```
./k3setup.sh master
```

then, on N number of nodes, run the node installer with output from the master:
```
./k3setup.sh node "10.0.0.10" "fu...ck::server:lo...ol"
```

By default, k8s hosts traefik on port 443 for the cluster. 
If the cluster is to host the web stack, you'll need to disable traefik on k3s:
This should only be done on the master!
```
vim /etc/systemd/system/k3s.service
#append --disable=traefik to the systemd run cmd
```

## Firewall
As of today, the only ports that should be exposed through the firewall are 
80TCP/443TCP. 8080TCP is plaintext only for the TOR router.
