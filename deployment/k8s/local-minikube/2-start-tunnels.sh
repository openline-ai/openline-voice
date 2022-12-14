#!/bin/bash

export NAMESPACE_NAME=voice-dev
minikube tunnel --bind-address 127.0.0.1 &
if [ $(uname -m) == "x86_64" ];
then
  kubectl port-forward --namespace $NAMESPACE_NAME svc/voice-plugin-service 11010:11010 &
fi

wait
