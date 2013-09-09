#!/bin/sh

echo "Removing remaining files"
sudo rm /tmp/*.{conf,cid}

echo "Stopping running containers"
docker ps | grep -v "CREATED" | grep -v "d9cd412ff77a" | awk '{print $1}' | xargs docker kill
echo "Deleting stopped containers"
docker ps -a | grep -v "CREATED" | grep -v "d9cd412ff77a" | awk '{print $1}' | xargs docker rm
#FIXME
#if [ -n "$IDS" -a "$IDS" != ^"Usage"* ]; then
    #echo "Ids: $IDS"
    #echo "$IDS" | xargs docker rm
#fi
echo "Removing previous images"
docker images | grep 'machine' |  awk '{print $3}' | xargs docker rmi

sudo -E sh -c "
    export GOMAXPROCS=\"$GOMAXPROCS\"
    export GOPATH=\"$GOPATH\"
    export GOROOT=\"$GOROOT\"
    export PATH=\"$PATH\"
    /opt/go/bin/go test $*
"
