#!/usr/bin/python3
import configparser
import os

def subsdef_escape(str):
    return str.replace("!", "\!")


with open("/etc/kamailio/local.conf", "w") as f:
    f.write("#!substdef \"!EPHEMERAL_AUTH_SECRET!%s!g\"\n" % (subsdef_escape(os.getenv("AUTH_SECRET"))))
    f.write("#!substdef \"!DMQ_DOMAIN!%s!g\"\n" % (subsdef_escape(os.getenv("DMQ_DOMAIN"))))
    f.write("#!substdef \"!HOMER_IP_ADDRESS!%s!g\"\n" % (subsdef_escape(os.getenv("HOMER_IP"))))
    f.write("#!substdef \"!DBURL!postgres://%s:%s@%s/%s!g\"\n" % (subsdef_escape(os.getenv("SQL_USER")), subsdef_escape(os.getenv("SQL_PASSWORD")), subsdef_escape(os.getenv("SQL_HOST")), subsdef_escape(os.getenv("SQL_DATABASE"))))


def escape(str):
    #only char that needs to be escaped according to
    #https://docs.python.org/3/library/configparser.html#interpolation-of-values
    return str.replace("%", "%%")


config = configparser.ConfigParser()
config['database'] = {'host': escape(os.getenv("SQL_HOST")),
                      'database': escape(os.getenv("SQL_DATABASE")),
                      'user': escape(os.getenv("SQL_USER")),
                      'password': escape(os.getenv("SQL_PASSWORD"))}

config['apiban'] = {'key': escape(os.getenv("APIBAN_KEY"))}

with open('/etc/kamailio/config.ini', 'w') as configfile:
    config.write(configfile)
