#!/bin/bash

rm -rf /var/lib/apt/lists/* && apt-get update
DEBIAN_FRONTEND=noninteractive apt-get install -qq --assume-yes libluajit-5.1-common libluajit-5.1-dev lsb-release wget curl git wget

cd /tmp/
wget https://github.com/sipcapture/homer-installer/raw/master/homer_installer.sh 
chmod +x homer_installer.sh
./homer_installer.sh
sed -i -e "s/AlegIDs\s*=\s*\[\]/AlegIDs        = \[\"X-Openline-UUID\"]/g" /etc/heplify-server.toml
sed -i -e "s/CustomHeader\s*=\s*\[\]/CustomHeader        = \[\"X-Openline-Origin-Carrier\", \"X-Openline-Dest\", \"X-Openline-Origin-Carrier\", \"X-Openline-Dest-Carrier\", \"X-Openline-Dest-Endpoint-Type\", \"X-Openline-Endpoint-Type\", \"X-Openline-CallerID\"]/g" /etc/heplify-server.toml
