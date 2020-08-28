#!/bin/sh
mkdir -p /etc/satellite/keys
openssl req -nodes -new -x509 -subj "/C=US/ST=SC/L=Charleston/O=Hacker/CN=satellite" -keyout /etc/satellite/keys/key.pem -out /etc/satellite/keys/cert.pem -days 365 2> /dev/null
