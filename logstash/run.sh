#! /bin/bash
#
# run.sh
# Copyright (C) 2013 xavier <xavier@dokku-thib>
#
# Distributed under terms of the MIT license.
#

sudo java -jar /opt/logstash/bin/logstash.jar agent -f /opt/logstash/conf/ -v
