### K8s Infra

Last updated: 19Jan24

# Purpose
--------------
The purpose of this is to create a set of files and scripts to build a local cluster on any physical host with k8s.

### Back to the Shack
---------------------------------------------------------------------------------
>>> Should I Use Rancher?
>>> Rancher is a good tool to use if you have a lot of clusters to manage, with users that are in multiple projects across clusters. This allows you to manage the users in one location and apply to all the projects. It also provides a “single pane of glass” for looking at clusters and configurations.
>>> 
>>> When Not to Use Rancher
>>> If you only have one cluster, with only a few users, or it is only managed with CLI tools, Rancher may not be an appropriate tool. It adds a little bit of complexity, in addition to at least one more node for the Rancher cluster, plus its infrastructure such as load balancers, DNS entries, backups, and others.

### k3s
---------------------------------------------------------------------------------
So it turns out, I used rancher prematurely. It hurts, but it's true. A lot of the issues i've faced on local has been due to this, and I'll experience more and limit configuration with this approach.

Instead, use k3s, which will be better on small devices and bare metal. It also comes witih needed k8s components
https://github.com/k3s-io/k3s/releases/tag/v1.26.0+k3s1

>>> just a quick switch to k3 and it's working. Incredible :) Though I may have had it working in rancher and not realized it because of API responses not being what I expected :sweatsmile:
>>> fuck it, it'll be better on a smaller k8s solution anyway.

### EKS
---------------------------------------------------------------------------------
The decision to use k3s over EKS for EW Production infra was not made lightly. The cost and complexity of the service was prohibitive, as well as the lack of control over the control 
plane base hosts. Because of this approach, I'm saving myself $140/month (70 per cluster), and getting smaller and more granular control over my k8s clusters. This decision was also made to avoid
the vendor lock-in that this project is architected to ignore. 

SVC SETUP:
--------------------------
This section outlines how to set up the EW specific services on the underlying k3s hosts created under the bare_metal section. Current architecture calls for two seperate k3s clusters to allow
proxy ingress and HTTPS ingress on the same port. The current paradigm also has all configuration information living in secrets under the `config` directory, to be mounted on the running containers
in the case of configuration files (Tor/NGINX), or used in environment secrets for applications that take configuration through the environment. 

ON BOTH CLUSTERS:

#set up pull secret
```
kubectl create secret docker-registry ghcrcred \
  --docker-server=ghcr.io \
  --docker-username=ktraister \
  --docker-password= $GH_PAT\
  --docker-email=kayleigh.traister@gmail.com
```

#setup common config secrets
```
cd config && echo "edit" && kubectl apply -f .
```

ON PRIMARY CLUSTER:

#deploy mongo
```
cd mongo_single && kubectl apply -f .
```

#deploy billing
```
cd billing && kubectl apply -f .
```

#deploy random
```
cd random && kubectl apply -f .
```

#deploy webapp
```
cd webapp && kubectl apply -f .
```

#deploy exchange
```
cd exchange && kubectl apply -f .
```

#deploy nginx
```
cd nginx && kubectl apply -f .
```

#deploy tor
```
cd tor && kubectl apply -f .
```

#expose load balancer (NGINX) ports
```
kubectl expose deployment nginx --type=LoadBalancer --name=nginx-lb
```

#expose mongo service ports
```
kubectl expose deployment mongo --type=LoadBalancer --name=mongo-lb --port 27017
```

ON SECONDARY CLUSTER:

#setup proxy config secrets
```
cd config && echo "edit" && kubectl apply -f .
```

#deploy proxy
```
cd proxy && kubectl apply -f .
```

#expose proxy service ports
```
kubectl expose deployment ew-proxy --type=LoadBalancer --name=proxy-lb
```

