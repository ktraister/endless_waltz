### K8s Infra

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

So it turns out, I used rancher prematurely. It hurts, but it's true. A lot of the issues i've faced on local has been due to this, and I'll experience more and limit configuration with this approach.

Instead, use k3s, which will be better on small devices and bare metal. It also comes witih needed k8s components
https://github.com/k3s-io/k3s/releases/tag/v1.26.0+k3s1

>>> just a quick switch to k3 and it's working. Incredible :) Though I may have had it working in rancher and not realized it because of API responses not being what I expected :sweatsmile:
>>> fuck it, it'll be better on a smaller k8s solution anyway.

SVC SETUP:
--------------------------
This is the process for a new dev namespace with all the trimmings.
Pick and choose what you need to deploy for each unique part

#set up pull secret
```
kubectl create secret docker-registry ghcrcred \
  --docker-server=ghcr.io \
  --docker-username=ktraister \
  --docker-password= \
  --docker-email=kayleigh.traister@gmail.com
```

#deploy mongo
```
cd mongo && kubectl apply -f .
```

#deploy reaper
```
cd reaper && kubectl apply -f .
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

#expose random port
```
kubectl expose deployment ew-random --type=LoadBalancer --name=local-ew-random
```

#expose exchange port
```
kubectl expose deployment ew-exchange --type=LoadBalancer --name=local-ew-exchange
```

#expose webapp port
```
kubectl expose deployment ew-webapp --type=LoadBalancer --name=local-ew-webapp
```
