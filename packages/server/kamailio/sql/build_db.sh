#!/bin/bash

NAMESPACE_NAME="voice-dev"
FILES="standard-create.sql permissions-create.sql auth_db.sql carriers.sql"
if [ "x$1" == "xlocal-kube" ]; then
  while [ -z "$pod" ]; do
    pod=$(kubectl get pods -n $NAMESPACE_NAME|grep voice-postgres|grep Running| cut -f1 -d ' ')
    if [ -z "$pod" ]; then
      echo "database not ready waiting"
      sleep 1
    fi
    sleep 1
  done
  echo $FILES |xargs cat|kubectl exec -n $NAMESPACE_NAME -it $pod -- psql -U $SQL_USER $SQL_DATABASE
else
  echo $FILES |xargs cat| PGPASSWORD=$SQL_PASSWORD  psql -h $SQL_HOST $SQL_USER $SQL_DATABASE
fi
