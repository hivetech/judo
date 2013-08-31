#! /bin/bash
#
# run.sh
# Copyright (C) 2013 xavier <xavier@dokku-thib>
#
# Distributed under terms of the MIT license.
#

echo -e "__ Juju logger managemet"
echo -e "\n\t- Configuration available in /opt/logstash/conf"
echo -e "\t- Dashboard code in ~/Kibana, file config is KibanaConfig.rb"
echo -e "\t- Indexer used: elasticsearch in single node"
echo -e "\n\t- Default log files monitored: $HOME/.juju/local/log/*.log\n"

echo "Running logstash agent, it will take some time to initailize..."
sudo java -jar /opt/logstash/bin/logstash.jar agent -f /opt/logstash/conf/ -v &
sleep 10
echo "Running dashboard..."
sudo ruby $HOME/kibana/kibana.rb
