#! /bin/bash

# Deploy Images
NAMESPACE_NAME="voice-dev"
OASIS_NAME_SPACE="oasis-dev"
CUSTOMER_OS_NAME_SPACE="openline"
echo "script is $0"
VOICE_HOME="$(dirname $(readlink -f $0))/../../../"
echo "VOICE_HOME=$VOICE_HOME"
CUSTOMER_OS_HOME="$VOICE_HOME/../openline-customer-os"
OASIS_HOME="$VOICE_HOME/../openline-oasis"

function getCustomerOs () {
  if [ ! -d $CUSTOMER_OS_HOME ];
  then
    cd "$VOICE_HOME/../"
    git clone https://github.com/openline-ai/openline-customer-os.git
  fi
  cd $CUSTOMER_OS_HOME/deployment/scripts/old/
  ../0-get-config.sh
  cd "$VOICE_HOME/../"
}

function getOasis () {
  if [ ! -d $OASIS_HOME ];
  then
    cd "$VOICE_HOME/../"
    git clone https://github.com/openline-ai/openline-oasis.git
  fi
}
if [ $(uname -m) != "x86_64" ];
then
    echo "CPU not supported x86_64 required"
    exit -1
fi
if [ -z "$(which kubectl)" ] || [ -z "$(which docker)" ] || [ -z "$(which minikube)" ] ; 
then
  if [ -z "$(which docker)" ]; 
  then
    INSTALLED_DOCKER=1
  else
    INSTALLED_DOCKER=0
  fi
  getCustomerOs
  if [ "x$(lsb_release -i|cut -d: -f 2|xargs)" == "xUbuntu" ];
  then
    echo "missing base dependencies, installing"
    cd $CUSTOMER_OS_HOME/deployment/scripts/old/
    $CUSTOMER_OS_HOME/deployment/scripts/old/1-ubuntu-dependencies.sh
    if [ $INSTALLED_DOCKER == 1 ];
	  then 
	    echo "Docker has just been installed"
	    echo "Please logout and log in for the group changes to take effect"
	    echo "Once logged back in, re-run this script to resume the installation"
	    exit
    fi
  fi
  if [ "x$(uname -s)" == "xDarwin" ]; 
  then
    echo "missing base dependencies, installing"
    cd $CUSTOMER_OS_HOME/deployment/scripts/old/
    $CUSTOMER_OS_HOME/deployment/scripts/old/1-mac-dependencies.sh
  fi
fi

MINIKUBE_STATUS=$(minikube status)
MINIKUBE_STARTED_STATUS_TEXT='Running'
if [[ "$MINIKUBE_STATUS" == *"$MINIKUBE_STARTED_STATUS_TEXT"* ]];
  then
     echo " --- Minikube already started --- "
  else
     eval $(minikube docker-env)
     minikube start &
     wait
fi

if [[ $(kubectl get namespaces) == *"$CUSTOMER_OS_NAME_SPACE"* ]];
then
  echo "Customer OS Base already installed"
else
  echo "Installing Customer OS Base"
  getCustomerOs
  cd $CUSTOMER_OS_HOME/deployment/scripts/old/
  $CUSTOMER_OS_HOME/deployment/scripts/old/2-base-install.sh
fi

if [ -z "$(kubectl get deployment customer-os-api -n $CUSTOMER_OS_NAME_SPACE)" ]; 
then
  echo "Installing Customer OS Aplicaitons"
  getCustomerOs
  cd $CUSTOMER_OS_HOME/deployment/scripts/old/
  $CUSTOMER_OS_HOME/deployment/scripts/old/3-deploy.sh
fi  

if [[ $(kubectl get namespaces) == *"$OASIS_NAME_SPACE"* ]];
then
  echo "Oasis already installed"
else
  echo "Installing Oasis"
  getOasis
  $OASIS_HOME/deployment/k8s/local-minikube/0-build-deploy-openline-oasis-local-images.sh $1
fi

if [[ $(kubectl get namespaces) == *"$NAMESPACE_NAME"* ]];
  then
    echo " --- Continue deploy on namespace $NAMESPACE_NAME --- "
  else
    echo " --- Creating $NAMESPACE_NAME namespace in minikube ---"
    kubectl create -f "$VOICE_HOME/deployment/k8s/local-minikube/voice-dev.json"
    wait
fi

## Build Images
cd $VOICE_HOME/deployment/k8s/local-minikube

minikube image load postgres:13.4 --pull

kubectl apply -f postgres/postgresql-configmap.yaml --namespace $NAMESPACE_NAME
kubectl apply -f postgres/postgresql-storage.yaml --namespace $NAMESPACE_NAME
kubectl apply -f postgres/postgresql-deployment.yaml --namespace $NAMESPACE_NAME
kubectl apply -f postgres/postgresql-service.yaml --namespace $NAMESPACE_NAME

cd  $VOICE_HOME

if [ "x$1" == "xbuild" ]; then
  if [ "x$(lsb_release -i|cut -d: -f 2|xargs)" == "xUbuntu" ];
  then
    if [ -z "$(which protoc)" ]; 
    then
	    sudo apt-get update
	    sudo apt-get install -y unzip wget
	    cd /tmp/
	    wget https://github.com/protocolbuffers/protobuf/releases/download/v21.9/protoc-21.9-linux-x86_64.zip
	    unzip protoc-21.9-linux-x86_64.zip
	    sudo mv bin/protoc /usr/local/bin
	    sudo mv include/* /usr/local/include/
    fi
    if [ -z "$(which go)" ]; 
    then
	    sudo apt-get update
	    sudo apt-get install -y golang-go
	    mkdir -p ~/go/{bin,src,pkg}
	    export GOPATH="$HOME/go"
	    export GOBIN="$GOPATH/bin"
    fi
    if [ -z "$(which make)" ]; 
    then
	    sudo apt-get install make
    fi
  fi
  if [ "x$(uname -s)" == "xDarwin" ]; 
  then
	  brew install protobuf
  fi
  
  cd  $VOICE_HOME
  if [ $(uname -m) == "x86_64" ];
  then
    cd packages/server/kamailio/;docker build -t ghcr.io/openline-ai/openline-voice/openline-kamailio-server:otter .;cd $VOICE_HOME
    cd packages/server/asterisk/;docker build -t ghcr.io/openline-ai/openline-voice/openline-asterisk-server:otter .;cd $VOICE_HOME
    cd packages/apps/voice-plugin/;docker build -t ghcr.io/openline-ai/openline-voice/voice-plugin:otter .;cd $VOICE_HOME
  fi
else
  
  if [ $(uname -m) == "x86_64" ];
  then
    docker pull ghcr.io/openline-ai/openline-voice/openline-kamailio-server:otter
    docker pull ghcr.io/openline-ai/openline-voice/openline-asterisk-server:otter
    docker pull ghcr.io/openline-ai/openline-voice/voice-plugin:otter
  fi


fi

if [ $(uname -m) == "x86_64" ];
then
  minikube image load ghcr.io/openline-ai/openline-voice/openline-kamailio-server:otter --daemon
  minikube image load ghcr.io/openline-ai/openline-voice/openline-asterisk-server:otter --daemon
  minikube image load ghcr.io/openline-ai/openline-voice/voice-plugin:otter --daemon
fi

if [ $(uname -m) == "x86_64" ];
then
  cd $VOICE_HOME/packages/server/kamailio/sql
  SQL_USER=openline-voice SQL_DATABABASE=openline-voice ./build_db.sh local-kube
fi
  
cd $VOICE_HOME/deployment/k8s/local-minikube

if [ $(uname -m) == "x86_64" ];
then
  kubectl apply -f apps-config/asterisk.yaml --namespace $NAMESPACE_NAME
  kubectl apply -f apps-config/asterisk-k8s-service.yaml --namespace $NAMESPACE_NAME
  kubectl apply -f apps-config/kamailio.yaml --namespace $NAMESPACE_NAME
  kubectl apply -f apps-config/kamailio-k8s-service.yaml --namespace $NAMESPACE_NAME
  kubectl apply -f apps-config/kamailio-k8s-loadbalancer-service.yaml --namespace $NAMESPACE_NAME
  kubectl apply -f apps-config/voice-plugin.yaml --namespace $NAMESPACE_NAME
  kubectl apply -f apps-config/voice-plugin-k8s-service.yaml --namespace $NAMESPACE_NAME
fi

cd $VOICE_HOME/deployment/k8s/local-minikube
