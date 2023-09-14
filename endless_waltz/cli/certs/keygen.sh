#!/bin/bash

unique=$(date +%s)
openssl req -x509 -sha256 -nodes -newkey rsa:4096 -keyout $unique.com.key -days 730 -out $unique.com.pem -config san.cnf
