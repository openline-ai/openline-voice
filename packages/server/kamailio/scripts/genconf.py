#!/usr/bin/python3
import configparser
import os

with open("/etc/kamailio/local.conf", "w") as f:
    f.write("#!substdef \"!EPHEMERAL_AUTH_SECRET!%s!g\"\n" % (os.getenv("AUTH_SECRET")))
    f.write("#!substdef \"!DBURL!postgres://%s:%s@%s/%s!g\"\n" % (os.getenv("SQL_USER"), os.getenv("SQL_PASSWORD"), os.getenv("SQL_HOST"), os.getenv("SQL_DATABASE")))

config = configparser.ConfigParser()
config['database'] = {'host': os.getenv("SQL_HOST"),
                      'database': os.getenv("SQL_DATABASE"),
                      'user': os.getenv("SQL_USER"),
                      'password': os.getenv("SQL_PASSWORD")}
with open('/etc/kamailio/config.ini', 'w') as configfile:
    config.write(configfile)