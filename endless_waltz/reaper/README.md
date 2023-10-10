# Reaper
This service is used to generate random one-time pads and upload them to the 
MongoDB instance where they will be served by the RandomAPI service. 

## Operation
### On Start
On startup, the Reaper binary reads in configuration variables from the 
environment. These variables deal with Mongo and Logging. The variable
`WriteThreshold` deserves special attention: it sets the number of pads
reaper will maintain in the OTP database.The application then enters an 
indefinite loop.

### Main
In each loop, it connects to Mongo and reads how many items are in the "otp" 
collection. If the count is under the threshold set in the environment, the 
diff is calculated and passed to the insertItems func.

In the insertItems func, the diff is limited to 100 to use network bandwidth 
effeciently. New UUIDs are generated and checked for uniqueness against the 
database. If they are unique, they are combined with a raw otp generated
by the createOTP function. The pair is the added to the slice used to batch
upload items to the database. Once the correct count of items has been 
aggregated, it is written to the db using the BulkWrite method.

If the count meets the correct limit, the process sleeps for 10 seconds 
before looping.

## Dockerfile
The Dockerfile has a special command to make the container compatable with
the `ew-rtl-entropy` container. On start, the command will remove 
`/dev/urandom` and link `/{tmp,dev}/urandom`, and then run the reaper binary.
Note: /tmp is passed through a volume from `ew-rtl-entropy`
