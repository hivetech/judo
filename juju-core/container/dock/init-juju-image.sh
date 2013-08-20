#!/bin/sh
#
# init-juju-image.sh
# Copyright (C) 2013 xavier <xavier@dokku-thib>
#
# Distributed under terms of the MIT license.
#

if [ $# != 2 ]; then 
    exit 1
fi

base_image=$1
target_image=$2

#ID=$(docker run -d $base_image /bin/bash -c "apt-get update && apt-get install -y cloud-init")
#ID=$(/usr/bin/docker run -d $base_image /bin/bash -c "apt-get install -y cloud-init")
ID=$(/usr/bin/docker run -d $base_image /bin/bash -c "ps")
/usr/bin/docker wait $ID
/usr/bin/docker commit $ID $target_image
