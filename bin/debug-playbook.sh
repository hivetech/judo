#! /bin/bash
#
# debug-playbook.sh
# Copyright (C) 2013 xavier <xavier@dokku-thib>
#
# Distributed under terms of the MIT license.
#

id=$1
ansible-playbook /var/lib/juju/ansible/cloudinit.yml -u root --extra-vars="hosts=xavier-hive-machine-$id config_vars=/var/lib/juju/containers/xavier-hive-machine-$id/cloud-init"
