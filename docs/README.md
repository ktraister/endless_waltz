# Documentation
This directory contains all of the documentation for the Endless Waltz 
messenger. Individual services are split out into seperate directories, while 
common things are given their own directories. 

LAST UPDATED: 19 January 24

## Naming Conventions
The project gets its name from the movie Gundam Wing: Endless Waltz, and hosts
are named to reflect this. The environment contains:
  - DeathScythe: My laptop :) 
  - Ultron: AWS prod Cluster hosting K3s primary (Altron)
  - Epion: AWS prod Singular hosting K3s secondary
  - *Zero: former physical machine, hosting dev and prod infra*
  - *HeavyArms: former AWS Cloud server hosting k3s for dev infra work*
  - *Shenlong: Defunct free GCP Cloud server hosting init k3s*
  - *Sandrock: Available*
  - *Wing: Available*
  - *Talgeese: Available*
  - *Mercurius: Available*
  - *Vayeate: Available*

## Applications
The EW Messenger has six discreet applications:
  - [RandomAPI](../endless_waltz/random/README.md)
  - [Exchange](../endless_waltz/exchange/README.md)
  - [WebApp](../endless_waltz/webapp/README.md)
  - [Billing](../endless_waltz/billing/README.md)
  - [Proxy](../endless_waltz/proxy/README.md)
  - [Messenger](https://github.com/ktraister/ew_messenger)

Everything except Messenger lives in this repository and is 
written in GoLang.
EW_Messenger lives in ktraister/ew_messenger, and is written in GoLang.

Random, Exchange, and Webapp share common functions within the 
`common` directory. Functionality is documented in the [common README](../endless_waltz/common/README.md)

### Utilities
Some utilites exist for use in the infrastructure or other locations to 
support the Endless Waltz ecosystem. These are also written in GoLang. 

  - [ewSysMgr](../endless_waltz/utils/ewSysMgr/README.md)

For more information, read the individual READMEs for the appropriate service.

## Infrastructure
The standard paradigm of the EW infrastructure is to deploy Alpine containers 
on Kubernetes in both the Cloud and on physical machines. K3s is used on the 
small dev deployments, while a managed service should be used for production

For more information, read the [infrastructure READMEs](../infra/README.md).

## Database
A single Mongo database is used to host authentication information for users, 
as well as one-time pads to be used in messaging. 

For more information, read the [database README](./database/README.md).

## Supporting Docs
This is a collection of documents and scripts of various languages used to 
support development of the applications.
  - [CertBot](./CertBot/README.md): docs for TLS Certs using CertBot
  - [LiveISO](./LiveISO/README.md): docs for the Endless Waltz LiveISO setup
  - [automation](./automation/README.md): docs for the Endless Waltz LiveISO setup
  - [security](./security/README.md): docs around Endless Waltz security posture
  - [database](./database/README.md): docs around Endless Waltz database interaction and operation

### Context
Context is important to be able to understand what you read. Therefor, I plan
to keep the respective documentation as close to the code as possible. 

Bear this in mind as you read the documentation and navigate the repo. 

