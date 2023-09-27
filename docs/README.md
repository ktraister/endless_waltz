# Documentation
This directory contains all of the documentation for the Endless Waltz messenger. 
Individual services are split out into seperate directories, while common things
are given their own directories. 

## Naming Conventions
The project gets its name from the movie Gundam Wing: Endless Waltz, and hosts
are named to reflect this. As of 26 September 2023, the environment contains:
  - Zero: Physical development machine, hosting dev and prod infra
  - HeavyArms: Cloud server hosting k3s for dev infra work
  - DeathScythe: My laptop :) 

## Applications
As of 26 September 2023, the EW Messenger has five discreet applications:
  - [RTL-Entropy](https://github.com/ktraister/ew-rtl-entropy) 
  - [Reaper](../endless_waltz/reaper/README.md)
  - [RandomAPI](../endless_waltz/random/README.md)
  - [Exchange](../endless_waltz/exchange/README.md)
  - [Messenger](../endless_waltz/gui/README.md)

Everything except RTL-Entropy lives in this repository and is written in GoLang.
RTL-Entropy lives in ktraister/ew-rtl-entropy, and is written in C.

For more information, read the individual READMEs for the appropriate service.

## Infrastructure
The standard paradigm of the EW infrastructure is to deploy Alpine containers 
on Kubernetes in both the Cloud and on physical machines. K3s is used on the 
physical machine side, and in the Cloud for the POC. The Physical machine is 
also outfitted with an RTL-SDR (Software Defined Radio) to collect random data.
A VPN is used to connect the physical host to the cloud hosts for one time pad
transmission.

Note: One radio/antenna on Zero is for dev, the other for prod. 

For more information, read the [infrastructure READMEs](../infra/README.md).

## Database
A single Mongo database is used to host authentication information for users, 
as well as one-time pads to be used in messaging. 

For more information, read the [database README](./database/README.md).

## Supporting Docs
This is a collection of documents and scripts of various languages used to 
support development of the applications.
  - [random_numbers](./random_numbers/README.md): used to generate new prime numbers for rn.go in the 
    messenger application. 
  - [passwd](./passwd/README.md): used to generate password hashes for user that match go output

### Context
Context is important to be able to understand what you read. Therefor, I plan
to keep the respective documentation as close to the code as possible. 

Bear this in mind as you read the documentation and navigate the repo. 
