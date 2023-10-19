# Security
This page is intended to inform the reader of security actions undertaken on 
the EW Circut and messenger.

## Scanning 
A great source of info on website and infra vulns and issues has been 
my personal account at `hostedscan.com`. This could be extended and leveraged
in the future. NMap is also helpful, and Nessus has been helpful in the past.

## WebApp
### Injection attacks
To prevent DB injection attacks, we check user input for the `Email` and 
`username` fields. The password field is hashed, so no risk of an injection 
attack there. Checks exist for email, username, and password correctness 
both on the server side(Go) and client side(JS). Checks are also performed
on password resets (all inputs, all forms).

### User Security
To ensure a certain level of security, password complexity requirements are
enforced on Server/Client side by Go/JS. This is enforced on reset as well.
