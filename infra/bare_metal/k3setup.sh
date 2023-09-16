#!/bin/bash
#

if [[ `whoami` != "root" ]]; then
    echo "Script must be run as root..."
    exit 1
fi

#install deps
apt update
apt install apt-transport-https ca-certificates curl software-properties-common -y
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu focal stable"

#install docker
apt update
apt install docker-ce -y
systemctl start docker
systemctl enable docker

if [[ ! -z $1 ]]; then
    #fix for MEEEE
    usermod -aG docker $1
    newgrp docker
fi

#setup master k3s node
curl -sfL https://get.k3s.io | sh -s - --docker
systemctl enable k3s

echo k3s master token:
echo "----------------------------------------------------------------------------------------------------"
cat /etc/rancher/k3s/k3s.yaml
echo "----------------------------------------------------------------------------------------------------"
