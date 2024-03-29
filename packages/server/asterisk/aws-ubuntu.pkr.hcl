variable "region" {
	type=string
	default="eu-west-2"
	sensitive=false
}

variable "environment" {
	type=string
	default="openline-dev"
	sensitive=false
}


data "amazon-parameterstore" "channels_api_key" {
  name = "/config/asterisk-server_${var.environment}/channels_api_key"
  with_decryption = true
}

data "amazon-parameterstore" "channels_api_service" {
  name = "/config/asterisk-server_${var.environment}/channels_api_service"
  with_decryption = false
}

data "amazon-parameterstore" "gladia_api_key" {
  name = "/config/asterisk-server_${var.environment}/gladia_api_key"
  with_decryption = true
}


# usage example of the data source output
locals {
  channels_api_key   = data.amazon-parameterstore.channels_api_key.value
  channels_api_service   = data.amazon-parameterstore.channels_api_service.value
  gladia_api_key   = data.amazon-parameterstore.gladia_api_key.value
}
packer {
  required_plugins {
    amazon = {
      version = ">= 0.0.2"
      source  = "github.com/hashicorp/amazon"
    }
  }
}

locals {
  opus_codec  = "asterisk-18.0/x86-64/codec_opus-18.0_current-x86_64"
}

source "amazon-ebs" "ubuntu" {
  ami_name      = "asterisk-server-ami_${var.environment}"
  instance_type = "t2.micro"
  region        = "${var.region}"
  source_ami_filter {
    filters = {
      name                = "ubuntu/images/*ubuntu-jammy-22.04-amd64-server-*"
      root-device-type    = "ebs"
      virtualization-type = "hvm"
    }
    most_recent = true
    owners      = ["099720109477"]
  }
  ssh_username = "ubuntu"
}

build {
  name    = "build-asterisk-image"
  sources = [
    "source.amazon-ebs.ubuntu"
  ]
  provisioner "shell" {
    inline = [
      "sudo sh -c 'add-apt-repository universe && apt-get update'",
      "sudo sh -c 'apt-get install -y asterisk sox python2 golang'",
      "sudo sh -c 'mkdir -p /usr/src/codecs/opus'",
      "sudo sh -c 'cd /usr/src/codecs/opus && curl -sL http://downloads.digium.com/pub/telephony/codec_opus/${local.opus_codec}.tar.gz | tar --strip-components 1 -xz'",
      "sudo sh -c 'cp /usr/src/codecs/opus/*.so /usr/lib/x86_64-linux-gnu/asterisk/modules/'",
      "sudo sh -c 'cp /usr/src/codecs/opus/codec_opus_config-en_US.xml /usr/share/asterisk/documentation/'",
      "sudo sh -c 'rm /usr/lib/x86_64-linux-gnu/asterisk/modules/format_ogg_opus_open_source.so'",
      "mkdir /tmp/asterisk/",

    ]
  }
  provisioner "file" { 
	source = "conf"
	destination = "/tmp/asterisk/"
  }
  provisioner "file" { 
	source = "scripts"
	destination = "/tmp/asterisk/"
  }
  provisioner "file" { 
	source = "ari"
	destination = "/tmp/asterisk/"
  }
  provisioner "file" { 
	source = "awslogs"
	destination = "/tmp/asterisk/"
  }
  provisioner "shell" {
    inline = [
      "sudo sh -c 'cp -v /tmp/asterisk/conf/* /etc/asterisk/'",
      "sudo sh -c 'cp -v /tmp/asterisk/scripts/asterisk_network_setup.sh /usr/sbin/'",
      "sudo sh -c 'chmod a+x /tmp/asterisk/scripts/asterisk_network_setup.sh'",
      "sudo sh -c 'cp -v /tmp/asterisk/scripts/asterisk_config.sh /usr/sbin/'",
      "sudo sh -c 'chmod a+x /tmp/asterisk/scripts/asterisk_config.sh'",
      "sudo sh -c 'CHANNELS_API_SERVICE=\"${local.channels_api_service}\" CHANNELS_API_KEY=\"${local.channels_api_key}\" GLADIA_API_KEY=\"${local.gladia_api_key}\"  /usr/sbin/asterisk_config.sh'",
      "sudo sh -c 'mv /tmp/asterisk/scripts/asterisk.service /etc/systemd/system'",
      "sudo sh -c 'chmod 644 /etc/systemd/system/asterisk.service'",
      "sudo sh -c 'mv /tmp/asterisk/scripts/asterisk_ari.service /etc/systemd/system'",
      "sudo sh -c 'chmod 644 /etc/systemd/system/asterisk_ari.service'",
      "sudo sh -c 'systemctl enable asterisk_ari.service'",
      "sudo sh -c 'cd /tmp/; curl https://s3.amazonaws.com/aws-cloudwatch/downloads/latest/awslogs-agent-setup.py -O; chmod a+x awslogs-agent-setup.py'",
      "sudo sh -c 'cd /tmp/; python2 ./awslogs-agent-setup.py -r ${var.region} -n -c /tmp/asterisk/awslogs/awslogs.conf'",
      "sudo sh -c 'cd /tmp/asterisk/ari;go mod download;go build -o /usr/local/bin/record_agi'",
    ]
  }
}

