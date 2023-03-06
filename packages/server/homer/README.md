# openline-voice-homer-server

To build the packer image you can do as follows
```
packer init aws-debian.pkr.hcl
packer build aws-debian.pkr.hcl
```

To build for production you must specify the region
```
packer init aws-debian.pkr.hcl
packer build -var 'region=eu-west-1' aws-debian.pkr.hcl
```

Default login:
admin / sipcapture
