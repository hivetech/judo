#!/bin/sh

echo "Stopping running containers"
docker ps | grep -v "ID" | awk '{print $1}' | xargs docker kill
echo "Deleting stopped containers"
#FIXME
#if [ -n "$IDS" -a "$IDS" != ^"Usage"* ]; then
    #echo "Ids: $IDS"
    #echo "$IDS" | xargs docker rm
#fi
docker ps -a | grep -v "ID" | awk '{print $1}' | xargs docker rm
echo "Removing previous images"
docker images | grep 'machine' |  awk '{print $3}' | xargs docker rmi

sudo sh -c "
    export GOMAXPROCS=\"$GOMAXPROCS\"
    export GOPATH=\"$GOPATH\"
    export GOROOT=\"$GOROOT\"
    export PATH=\"$PATH\"
    /opt/go/bin/go test $*
"
