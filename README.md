# endless_waltz

POC in python, working code in go
----------------------------------

## Phase I
### just create something
***Complete :)***
- [x] implement api, server, and client
- [x] implement xor and pad code
- [x] implement redis for caching

## Phase II
### get basic functionality working in GoLang
***80% Complete***
- [x] rewrite API in Go
   - moved UUID generation for connection from server into API
- [x] implement xor library
   - having a prototype made this way easier
   - need xor lib to pass around strings
   - need xor lib to import smoothly
- [x] rewrite Server in Go
   - server is working up to after a pad is requested and UUID sent to client
- [x] rewrite Client in Go
   - need one set of random APIs to garuntee MITM attacks are mitigated
- [x] create dockerfiles to host each component
- [x] add configuration to daemons 
- [x] get the compose file working for local dev
- [x] cleanup output from containers, useful logging
- [x] fix conn.close on server/client transaction
- [x] add flags to send message to client
- [x] Remove Redis and update to use Mongo
- [x] Add API route to upload pads -- make the service cloud agnostic!!!
   - Mongo should be used to store pads as well (!!!)
- [ ] add DH handshake!!
   - dh handshake with rediculous values will be used for pad transformation and message signing
   - two different values will need to be calculated
- [ ] **CODE CLEANUP** 
   - BEFORE PROCEEDING:
   - [ ] ensure all dead code is removed
   - [ ] make sure server/api/client do not quit prematurely
   - [ ] implement at least basic unit test coverage -- run on commit

## Phase III
### get something working with hardware to deploy to cloud
***10% Complete***
- [x] order hardware
- [x] confirm operation of ew-rtl-entropy binary and containerize
   - need to make sure this binary will work, output works, containerize
- [ ] write reaper in Go to live on physical hardware
   - reaper will depend on a C executable for RNG using an SDR for randomness
   - [x] Make a pipe inside a volume and mount it on `/dev/urandom` where the Go lib reads -- dont fight it lol
   - reaper and ew-rtl-entropy are working together in docker -- compose file works
- [ ] move upload api to seperate port
- [ ] setup automation for CI/CD
   - GitHub Actions is the easiest and closest, and I'm already paying for pro - try this first
   - need to choose a provider -- CircleCI provides most free build minutes per month
- [ ] Create K8s helm charts for services
   - API
   - Reaper
- [ ] further logging improvements
- [ ] create infrastructure for project in AWS
   - terraform IAC
- [ ] update docker build to respect any env (API, Server)
   - a dockerfile with VARIABLE references to config files should be sufficient
- [ ] start padding the message with random data to prevent length attacks
   - pad should be random, use delimeter like "###" to signify padding
- [ ] setup server to interact with cloud/prod env
   - I wont be looking at the logs for messages, and wont want them decrypted until I'm ready
   - add a flag, localdev is good as is

### At this point, we're ready for the deepweb and linux users

## Architecture
![alt text](./EndlessWaltz.png)


## Phase IV
### needed to make the tech accesible
- [ ] setup website
- [ ] need Android client
   - ugh. This is gonna be intense
   - To make it accessable, it needs to work easily on your phone. Workflow may look like
     - Sally opens app. Tunnel to server established.
     - Betty is offline. Sally pings Betty.
     - Betty opens app, establishes unique tunnel to server
     - apps perform DH, get pad, send message
     - tunnels are closed, msg deleted
- [ ] needs IOS client
- [ ] turn it loose

