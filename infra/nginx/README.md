# NGINX
NGINX is a fast proxy service that is used to load balance and redirect 
incoming traffic to the EW circut. The configuration file is mounted on 
on the appropriate path in all environments. The localhost directory is
used for keys and certs used in local development.

## Development Cert
LocalDev cert was generated using this command:
```
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -sha256 -days 3650 -nodes -subj "/C=EW/ST=SancKingdom/L=ESUN/O=OperationMeteor/OU=XXXG-01D/CN=Deathscythe"
```
