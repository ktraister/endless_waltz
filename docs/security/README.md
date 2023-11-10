# Security
This page is intended to inform the reader of security actions undertaken on 
the EW Circut and messenger.

# Scanning 
A great source of info on website and infra vulns and issues has been 
my personal account at `hostedscan.com`. This could be extended and leveraged
in the future. NMap is also helpful, and Nessus has been helpful in the past.

# Applications
Applications are writen in GoLang, which is thread-safe and memory safe.
Per the NSA. lol. 
https://www.nsa.gov/Press-Room/News-Highlights/Article/Article/3215760/nsa-releases-guidance-on-how-to-protect-against-software-memory-safety-issues/

Specifically for the messenger (but really for all applications), accessing
messages within their memory space is realistically possible for privileged
users. However, there is nothing I can do about a user getting their root
(admin) users popped except my TAILS-like os. 
https://stackoverflow.com/questions/1989783/how-is-it-possible-to-access-memory-of-other-processes

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

### CSRF
CSRF was possible with the old webapp implementation. Addressed in issue #199.

## NGINX
