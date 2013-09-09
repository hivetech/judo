#! /bin/bash
#
# clean.sh
# Copyright (C) 2013 xavier <xavier@laptop-300E5A>
#
# Distributed under terms of the MIT license.
#

# Beware: docker containers and images are not cleaned

echo "Destroying juju environment"
sudo $GOPATH/bin/juju -v destroy-environment

echo "Cleaning /tmp directory (cid and agent.conf file)"
sudo rm /tmp/*.{cid,conf}

echo "Cleaning ansible hosts inventory"
sudo bash -c "echo  > /etc/ansible/hosts"

echo "Cleaning deprecated containers"
sudo rm -r /var/lib/juju/containers/*

echo "Stopping juju processes"
ps -ef | grep juju | grep -v grep | awk '{print $2}' | xargs sudo kill

echo "Destroying $HOME/.juju configuration"
sudo rm -r $HOME/.juju

#echo "Cleaning deprecated services config"
#sudo rm /etc/init/juju-*
