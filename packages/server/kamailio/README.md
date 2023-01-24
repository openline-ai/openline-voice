# openline-voice-kamailio-server

## docker
currently this docker image only supports WebRTC to WebRTC

to build the docker image:

```
./build-docker.sh
```

to run the kamailio server:

```
./start.sh
```
to stop the kamailio server:

```
./stop.sh
```
## packer ami
to an aws ami image


to build the api image you must have AWS_ACCESS_KEY_ID AWS_SECRET_ACCESS_KEY AWS_REGION properly set with your AWS credentials

This script expects variables to be set in the aws parameter store
the build is set up so that it can be set up for multiple environments the default environment name is 'uat-ninja'

The following paramstore keys need to be set, if you are not using uat-ninja as an environemnt please replace 'uat-ninja' with your environment name

| key                                           | meaning                                                                                                                |
| --------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| /config/kamailio-server_{ENV}/auth_secret | the shared secret to use for ephemeral authentication, this needs to be set to the same value as inside the Oaisis app |
| /config/kamailio-server_{ENV}/db_database | the name of the postgres database to use                                                                               |
| /config/kamailio-server_{ENV}/db_host     | ip or hostname of the postgres database                                                                                |
| /config/kamailio-server_uat-{ENV}/db_password | password to use to connect to the database                                                                             |
| /config/kamailio-server_uat-{ENV}/db_user     | username to connect to the database with                                                                               |
| /config/kamailio-server_{ENV}/dmq_domain      | domain containining all kamailio (so DMK can idenfity all nodes in the culster) |
| /config/kamailio-server_{ENV}/homer_ip        | internal ip address of the homer server (to send traffic to) |

To build the packer image you can do as follows
```
packer init aws-debian.pkr.hcl
packer build aws-debian.pkr.hcl
```

To build the packer image for a different enviroment, you can specify the enviornment as a variable
```
packer init aws-debian.pkr.hcl
packer build -var "environment=development" aws-debian.pkr.hcl
```

For production you need to specify both the environment and the region
*DO NOT FORGET TO UPDATE THE AWS_REGION ENV VAR AS WELL*
```
export AWS_REGION=eu-west-1
packer init aws-debian.pkr.hcl
packer build -var "environment=development" -var "region=eu-west-1" aws-debian.pkr.hcl
```

## WebRTC

Webrtc requires the following variables to be set
| variable    | meaning                                                              |
| ----------- | -------------------------------------------------------------------- |
| AUTH_SECRET | This must match the value set in WEBRTC_AUTH_SECRET in the oasis-api |

## Database
The database schemas are included in the sql directory, it will also be available in the AMI image in the /tmp/kamailio/sql directory

To provision the database you need set following environment variables

| variable     | meaning                                      |
| ------------ | -------------------------------------------- |
| SQL_HOST     | IP or domain of the postgres server          |
| SQL_USER     | Username to run the sql as                   |
| SQL_PASSWORD | Password to run the sql as                   |
| SQL_DATABASE | Name of the database to insert the tables in |

```
./sql/build_db.sh
```

then run the script

## Testing
to run the unit tests, a postgress database on localhost will be required as well as the python package psycopg
```
cd tests
python3 -m unittest
```
