#!/bin/bash


sed "s/!CHANNELS_API_SERVICE!/$CHANNELS_API_SERVICE/g" /etc/asterisk/ari_record.conf.template| sed "s/!CHANNELS_API_SERVICE!/$CHANNELS_API_SERVICE/g" > /etc/asterisk/ari_record.conf
