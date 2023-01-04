#!/bin/sh

/etc/kamailio/genconf.py
/usr/sbin/kamailio_network_setup.sh
touch /etc/kamailio/dispatcher.list
kamailio -DD -E
