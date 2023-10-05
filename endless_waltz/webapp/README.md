# WebApp
The webapp service is used to allow end users to manage their accounts, as well as get information about Endless Waltz. 

## PII
To protect end users privacy and be more easily GDPR-compliant, this webapp
will serve only minimum cookies, and take as little PII as possible. Currently,
the end user is not required to enter an email address or phone number. This
may change in the future.

## Operation

### Env Variables
This application requires the following env variables to be set:
  - MongoURI: String value for mongo protocol/hostname/port
  - MongoUser: Login user for mongo
  - MongoPass: Login password for Mongo
  - SessionKey: Unique key to be used for session store in containers
  - CaptchaKey: Key provided by google to check if captcha responses are valid

### On Start
On startup, the webapp binary reads in configuration variables from the 
environment. These variables deal with Mongo and Logging. A new router mutex 
is created, and the routes listed below are configured with desired methods. 
The HTTP server is then started on port 8080 with the router mutex. 

### Routes
