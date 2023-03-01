#!/bin/sh
export PUBLIC_IP_EXTERNAL=$(curl 'https://api.ipify.org')
/usr/sbin/asterisk_network_setup.sh
/usr/sbin/asterisk_config.sh
(sleep 5; /usr/local/bin/record_agi) &
/usr/sbin/asterisk -T -W -U asterisk -p -vvvdddf