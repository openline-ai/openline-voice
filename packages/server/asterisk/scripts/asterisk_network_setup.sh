#!/bin/sh

if [ -z "$PUBLIC_IP_EXTERNAL" ]; then
PUBLIC_IP_EXTERNAL=$PUBLIC_IP
fi

sed "s/!PUBLIC_IP!/$PUBLIC_IP_EXTERNAL/g" /etc/asterisk/pjsip.conf.template| sed "s/!LOCAL_IP!/$LOCAL_IP/g" > /etc/asterisk/pjsip.conf

sed "s/!PUBLIC_IP!/$PUBLIC_IP/g" /etc/asterisk/rtp.conf.template| sed "s/!LOCAL_IP!/$LOCAL_IP/g" > /etc/asterisk/rtp.conf
