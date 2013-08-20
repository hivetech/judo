#!/bin/sh

#go build dock.go
cp -r ~/dev/judo/juju-core ~/.go/src/launchpad.net/
#cp * ${HOME}/.go/src/launchpad.net/juju-core/container/dock
#rm ${HOME}/.go/pkg/linux_amd64/github.com/judo/

#sudo -E /opt/go/bin/go test $*
#export GOMAXPROCS=\"$GOMAXPROCS\"
sudo bash -E -c "
    export GOPATH=\"$GOPATH\"
    export GOROOT=\"$GOROOT\"
    export PATH=\"$PATH\"
    /opt/go/bin/go test $*
"
