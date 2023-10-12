# Security
This page is intended to inform the reader of security actions undertaken on 
the EW Circut and messenger.

## WebApp
To prevent DB injection attacks, we're going to disallow any entries of curly
braces. In fact, we're going to disallow the username from containing any special characters at all. It'd just be a huge mess.
