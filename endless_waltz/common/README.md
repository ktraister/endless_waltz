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
the correct creds.

### LoggerMiddleware
LoggerMiddleware returns a generic Mux middleware func to be used within a
gorillaMux router. It injects a logger into the request that can then be
extracted by the executed function :)

## logging.go
### createLogger
createLogger is a generic function used to set logging level and return a 
configured logger. It also accepts an input to configure logging as JSON
or plain log strings. Any general logging updates should be made here
