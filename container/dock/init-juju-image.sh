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

base_image=$1
target_image=$2

ID=$(/usr/bin/docker run -d $base_image /bin/bash -c "mkdir -p /var/log/juju && apt-get update && apt-get install -y cloud-init")
#ID=$(/usr/bin/docker run -d $base_image /bin/bash -c "adduser --disabled-password --gecos \"\" ubuntu")
#ID=$(/usr/bin/docker run -d $base_image /bin/bash -c "useradd --disabled-password ubuntu")
#ID=$(/usr/bin/docker run -d $base_image /bin/bash -c "test -d /var/log/juju || mkdir -p /var/log/juju")
/usr/bin/docker wait $ID 2>&1 >> /tmp/init-juju.logs

# The only output is the new id generated
/usr/bin/docker commit $ID $target_image
