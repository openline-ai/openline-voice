# openline-voice-asterisk-server


## docker
currently this docker image only supports WebRTC to WebRTC

to build the docker image:

```
./build-docker.sh
```
## packer ami
to an aws ami image

to build the api image you must have AWS_ACCESS_KEY_ID AWS_SECRET_ACCESS_KEY AWS_REGION properly set with your AWS credentials

then you can run the following commands to build the image

```
packer init aws-ubuntu.pkr.hcl
packer validate aws-ubuntu.pkr.hcl
packer build aws-ubuntu.pkr.hcl
```

for production builds you also need to specify the region

```
packer init aws-ubuntu.pkr.hcl
packer validate aws-ubuntu.pkr.hcl
packer build -var 'region=eu-west-1' aws-ubuntu.pkr.hcl
```
