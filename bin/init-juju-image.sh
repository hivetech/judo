#!/bin/sh
#
# init-juju-image.sh
# Copyright (C) 2013 xavier <xavier@dokku-thib>
#
# Distributed under terms of the MIT license.
#

if [ $# != 2 ]; then 
    exit 58
fi

#FIXME juju is not yet able to choose a docker custome image
BASE_IMAGE="ubuntu:$1"
#BASE_IMAGE="ubuntu:precise"
#BASE_IMAGE="base"
TARGET_IMAGE=$2
DOCKER_BIN=$(whereis docker | cut -d" " -f2)

#ID=$(/usr/bin/docker run -d $base_image /bin/bash -c "mkdir -p /var/log/juju && apt-get update && apt-get install -y cloud-init")
#ID=$(/usr/bin/docker run -d $BASE_IMAGE /bin/bash -c "test -d /var/log/juju || mkdir -p /var/log/juju")

# For now, each machine uses a common modified base image
# So if already built before, re-uses it
JUJU_MACHINE_ID=$($DOCKER_BIN ps -a | grep "juju/$1" | cut -d" " -f1)

if [ -z "$JUJU_MACHINE_ID" ]; then
    JUJU_MACHINE_ID=$($DOCKER_BIN run -d $BASE_IMAGE /bin/bash -c "apt-get install -y python python-apt openssh-server && mkdir -p /var/{log/juju,run/sshd} && echo \"root:quant\" | chpasswd")
    $DOCKER_BIN wait $JUJU_MACHINE_ID 2>&1 >> /tmp/init-juju.logs
    $DOCKER_BIN commit $JUJU_MACHINE_ID juju/$1
fi

# Create requested specific image for later process
# Note: the new machine id is the only output of this script
$DOCKER_BIN commit $JUJU_MACHINE_ID $TARGET_IMAGE
#sleep 30
