### K8s Infra

# Purpose
--------------
The purpose of this is to create a set of scripts to build a cluster on any host.

HOST SETUP:
--------------------------
## install k8s
get your new cluster set up:
```
./k3setup.sh
```

By default, k8s hosts traefik on port 443 for the cluster. 
If the cluster is to host the web stack, you'll need to disable traefik on k3s:
```
vim /etc/systemd/system/k3s.service
#append --disable=traefik to the systemd run cmd
```

## VPN Server
setup openVPN server on remote host IF REQUIRED:
```
openvpn-install.sh
```

This will configure your first client at setup.
Re-run to generate new client configs for reaper boxes
Re-run to delete existing client configs

## VPN Client
```
sudo apt install openvpn
```

Next, copy iphone.ovpn as follows:
```
sudo cp iphone.ovpn /etc/openvpn/client.conf
```

Test connectivity from the CLI:
```
sudo openvpn --client --config /etc/openvpn/client.conf
sudo systemctl start openvpn@client
sudo systemctl enable openvpn@client
```

The VPN client has issues being long lived. To address this:
Setup VPN Restart CronJob
```
#install cron -- no longer standard on Ubuntu
sudo apt update && sudo apt install cron

sudo systemctl enable cron

sudo crontab -e 
#0 * * * * /usr/bin/systemctl restart openvpn@client
```

NEXT STEPS:
--------------------------
proceed to `../k8s` and follow readme


When Bouncing HeavyArms:
-------------------------
 - fixup client vpn config on Zero
 - fixup reaper env config on Zero
 - fixup ssh config on Deathscythe
 - fixup kubeconfig on Deathscythe
