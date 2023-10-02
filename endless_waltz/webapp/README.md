# WebApp

## Operation
### On Start
On startup, the webapp binary reads in configuration variables from the 
environment. These variables deal with Mongo and Logging. A new router mutex 
is created, and the routes listed below are configured with desired methods. 
The HTTP server is then started on port 8080 with the router mutex. 

### Routes
