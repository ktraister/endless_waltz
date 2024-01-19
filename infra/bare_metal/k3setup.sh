#!/bin/bash

if [[ `whoami` != "root" ]]; then
    echo "Script must be run as root..."
    exit 1
fi

usage() {
        echo "Usage: k3setup.sh [master/secondary/node]"
	echo "       if 'secondary' OR 'node': master_ip master_node_token"
    }

if [[ ! `which docker` ]]; then
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
fi

if [[ $1 == "master" ]]; then
	#setup master k3s node
	curl -sfL https://get.k3s.io | sh -s - 
	systemctl enable k3s
	echo k3s master token:
	echo "----------------------------------------------------------------------------------------------------"
	cat /etc/rancher/k3s/k3s.yaml
	echo "----------------------------------------------------------------------------------------------------"
	echo
	echo k3s node token:
	echo "----------------------------------------------------------------------------------------------------"
	cat /var/lib/rancher/k3s/server/node-token
	echo "----------------------------------------------------------------------------------------------------"
elif [[ $1 == "secondary" ]]; then
        echo "THIS IS NOT YET SUPPORTED"
	exit 128
        if [[ -z $2 ]] || [[ -z $3 ]]; then
	    usage
	    exit 2
        fi
        curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="server" K3S_URL="https://$2:6443" K3S_TOKEN="$3" sh -
	systemctl enable k3s
elif [[ $1 == "node" ]]; then
        if [[ -z $2 ]] || [[ -z $3 ]]; then
	    usage
	    exit 2
        fi
        curl -sfL https://get.k3s.io | K3S_URL="https://$2:6443" K3S_TOKEN="$3" sh -
	systemctl enable --now k3s-agent
else 
        echo "*****************************************************"
        echo "unrecognized input, skipping master AND node setup :)"
        echo "*****************************************************"
	usage
fi
