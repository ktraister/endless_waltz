# Billing
This service deals with batch actions for card and crypto billing. It is not externally exposed :)
This service should be run on a cron job in k8s, as it doesn't run in an indefinite loop.

## Operation
On start, the binary takes configuration information through environment variables. 5 top level functions
are executed to deal with batch processing of various workflows

### CryptoResolvePayments(logger)
Find all database records that have an associated coinbase charge. If the charge was paid, set user active
and reset the crypto billing values.

### CryptoBillingInit
Finds all records with a billing cycle end in the next 7 days, and crypto billing set. Send the email to 
invite the user to make the payment and update the database to reflect this.

### CryptoBillingReminder
Finds all records with a billing cycle end in the next 2 days, and crypto billing set. Send the email to 
invite the user to make the payment and update the database to reflect this. 

### CryptoDisableAccount
Find all active records with billing email flags set, as well as expired billing cycle. Lock all matching 
accounts. We also send a courtesy email to our users if we disable their accounts. 

### StripeSubscriptionChecks
Find all records with cardBilling set to true. Check subscription status. If not trialing or not active,
set active false if not set. Do the opposite for trialing or active true. Send an email to the user if we
disable their account :) 
