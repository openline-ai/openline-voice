#!/bin/sh

sed "s/#MY_PUBLIC_IP_ADDR#/$PUBLIC_IP/g" /etc/kamailio/network.conf.template| sed "s/#MY_PRIVATE_IP_ADDR#/$LOCAL_IP/g" > /etc/kamailio/network.conf

if [ -n "$ASTERISK_HOST" ]
then
echo "INSERT INTO kamailio_dispatcher (id, setid, destination, flags, priority) VALUES (1, 0, 'sip:$ASTERISK_HOST', 8, 0);" |PGPASSWORD="$SQL_PASSWORD"  psql -h $SQL_HOST $SQL_USER $SQL_DATABASE
fi
