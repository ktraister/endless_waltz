# Common
Common directory for common code shared throughout the backend services

## api_lib.go

### rateLimit
The rateLimit function is to be called after the checkAuth function. This
ensures that the user exists, and that someone without creds isn't trying
to block a user's access. Each unique user is allowed to make up to 5 
requests per second. The user is then throttled, returning 429. 

### checkAuth
checkAuth performs a db lookup to ensure the user is active and is passing
the correct creds. A flag can be passed to ignore the active check, which 
is used for the web application, but not the rest of the infrastructure. 

### LoggerMiddleware
LoggerMiddleware returns a generic Mux middleware func to be used within a
gorillaMux router. It injects a logger into the request that can then be
extracted by the executed function :)

## logging.go

### createLogger
createLogger is a generic function used to set logging level and return a 
configured logger. It also accepts an input to configure logging as JSON
or plain log strings. Any general logging updates should be made here

## security.go
This was moved out of the web application and into common for mobility of
functions. 

### generateToken
Used to return a string of random chars containing upper lower and numbers.
Used for generating one time passwords for email forms. 

### isEmailValid
Performs regex check of user email input and returns pass/fail

### isPasswordValid
Performs check on each character in input string. Returns true if it finds 
a number, upper, and special chars as well as length >= 8 chars. 

### checkUserInput
Checks user input for possible injection attacks. Returns false if the input
contains any special characters that could be used in an injection attack. 

### nextBillingCycle
Helper function that returns the next month-day-year date that the billing 
cycle will end based on input string. This function takes extra care to 
ensure that the date is returned in the specific 2-2-4 digit format so as 
not to break time marshalling in the application. 
