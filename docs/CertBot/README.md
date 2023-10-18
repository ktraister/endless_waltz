# CertBot
CertBot is provided by LetsEncrypt to manage free 90 day TLS certificates. 
Endless waltz uses this for load balancer certificates for TLS encryption. 

## setup and operation
https://certbot.eff.org/instructions?ws=other&os=ubuntufocal&tab=wildcard

## AWS Route53 specific docs
https://certbot-dns-route53.readthedocs.io/en/stable/

## Operation
Because certbot requires sudo, you'll have to configure the access keys for the
AWS account for the root user.
```
(ins)[~][none]> sudo certbot certonly   --dns-route53   -d '*.endlesswaltz.xyz'
Saving debug log to /var/log/letsencrypt/letsencrypt.log
Requesting a certificate for *.endlesswaltz.xyz

Successfully received certificate.
Certificate is saved at: /etc/letsencrypt/live/endlesswaltz.xyz/fullchain.pem
Key is saved at:         /etc/letsencrypt/live/endlesswaltz.xyz/privkey.pem
This certificate expires on 2024-01-16.
These files will be updated when the certificate renews.
Certbot has set up a scheduled task to automatically renew this certificate in the background.

- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
If you like Certbot, please consider supporting our work by:
 * Donating to ISRG / Let's Encrypt:   https://letsencrypt.org/donate
 * Donating to EFF:                    https://eff.org/donate-le
- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
```

## Renewals
Renewals can be done from the command line the same way the cert was requested.
Ensure you use both of the names for the certificate
```
(ins)[~][none]> sudo certbot certonly   --dns-route53  -d 'endlesswaltz.xyz' -d '*.endlesswaltz.xyz'
Saving debug log to /var/log/letsencrypt/letsencrypt.log

- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
You have an existing certificate that contains a portion of the domains you
requested (ref: /etc/letsencrypt/renewal/endlesswaltz.xyz.conf)

It contains these names: *.endlesswaltz.xyz

You requested these names for the new certificate: endlesswaltz.xyz,
*.endlesswaltz.xyz.

Do you want to expand and replace this existing certificate with the new
certificate?
- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
(E)xpand/(C)ancel: E
Renewing an existing certificate for endlesswaltz.xyz and *.endlesswaltz.xyz

Successfully received certificate.
Certificate is saved at: /etc/letsencrypt/live/endlesswaltz.xyz/fullchain.pem
Key is saved at:         /etc/letsencrypt/live/endlesswaltz.xyz/privkey.pem
This certificate expires on 2024-01-16.
These files will be updated when the certificate renews.
Certbot has set up a scheduled task to automatically renew this certificate in the background.

- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
If you like Certbot, please consider supporting our work by:
 * Donating to ISRG / Let's Encrypt:   https://letsencrypt.org/donate
 * Donating to EFF:                    https://eff.org/donate-le
- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
```

## Certificate Installation
Certificates will need to be `base64` encoded with the `-w 0` option and put 
into the appropriate secret in kubernetes. Then the nginx service will need 
to be restarted, After these steps, check the certificate in a web browser. 

Profit.
